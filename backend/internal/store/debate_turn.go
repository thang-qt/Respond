package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"respond/internal/model"
)

var (
	ErrDebateNotActive = errors.New("debate not in active status")
	ErrTurnNotYourTurn = errors.New("not your turn")
)

type SubmitTurnParams struct {
	DebateID   uuid.UUID
	UserID     uuid.UUID
	Content    string
	AIAssisted bool
	AINote     *string
}

func (s *Store) SubmitTurn(ctx context.Context, params SubmitTurnParams) (model.DebateTurn, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.DebateTurn{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Lock the debate row and fetch state
	const fetchQuery = `
		SELECT status, current_turn_side, turn_count, turn_limit, time_mode,
			side_a_user_id, side_b_user_id,
			side_a_anonymous_id, side_b_anonymous_id,
			topic
		FROM debates
		WHERE id = $1
		FOR UPDATE
	`

	var (
		status      string
		currentSide string
		turnCount   int
		turnLimit   int
		timeMode    string
		sideAUser   uuid.UUID
		sideBUser   sql.NullString
		sideAAnon   string
		sideBAnon   sql.NullString
		topic       string
	)

	if err := tx.QueryRowContext(ctx, fetchQuery, params.DebateID).Scan(
		&status, &currentSide, &turnCount, &turnLimit, &timeMode,
		&sideAUser, &sideBUser,
		&sideAAnon, &sideBAnon,
		&topic,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.DebateTurn{}, ErrNotFound
		}
		return model.DebateTurn{}, fmt.Errorf("fetch debate: %w", err)
	}

	if status != "active" {
		return model.DebateTurn{}, ErrDebateNotActive
	}

	// Determine which side this user is on
	var userSide string
	var anonID string
	switch {
	case sideAUser == params.UserID:
		userSide = "a"
		anonID = sideAAnon
	case sideBUser.Valid && sideBUser.String == params.UserID.String():
		userSide = "b"
		anonID = sideBAnon.String
	default:
		return model.DebateTurn{}, ErrTurnNotYourTurn
	}

	if currentSide != userSide {
		return model.DebateTurn{}, ErrTurnNotYourTurn
	}

	nextTurnNumber := turnCount + 1
	nextSide := "a"
	if userSide == "a" {
		nextSide = "b"
	}

	// Get the actual max turn_number from turns. Legacy debates may include
	// historical system rows, so turn_count may be lower than MAX(turn_number).
	// We keep turn_count for limit logic and use rowTurnNumber for uniqueness.
	var maxTurnNumber int
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(turn_number), 0) FROM turns WHERE debate_id = $1`,
		params.DebateID,
	).Scan(&maxTurnNumber); err != nil {
		return model.DebateTurn{}, fmt.Errorf("get max turn number: %w", err)
	}
	rowTurnNumber := maxTurnNumber + 1

	now := time.Now().UTC()
	deadline := now.Add(turnWindow(timeMode))
	extensionDeadline := now.Add(48 * time.Hour)

	// Insert the turn
	const insertTurn = `
		INSERT INTO turns (debate_id, turn_number, side, user_id, anonymous_id, content, ai_assisted, ai_note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	var turn model.DebateTurn
	turn.TurnNumber = rowTurnNumber
	turn.Side = userSide
	turn.AnonymousID = anonID
	turn.Content = params.Content
	turn.AIAssisted = params.AIAssisted
	turn.AINote = params.AINote

	if err := tx.QueryRowContext(ctx, insertTurn,
		params.DebateID, rowTurnNumber, userSide, params.UserID, anonID, params.Content, params.AIAssisted, params.AINote,
	).Scan(&turn.ID, &turn.CreatedAt); err != nil {
		return model.DebateTurn{}, fmt.Errorf("insert turn: %w", err)
	}

	// Update debate state
	// If we've reached the turn limit, pause for extension decision
	if nextTurnNumber >= turnLimit {
		const endDebate = `
			UPDATE debates
			SET turn_count = $2,
				current_turn_side = $3,
				turn_deadline = NULL,
				status = 'pending_extension',
				extension_deadline = $4
			WHERE id = $1
		`
		if _, err := tx.ExecContext(ctx, endDebate,
			params.DebateID, nextTurnNumber, nextSide, extensionDeadline,
		); err != nil {
			return model.DebateTurn{}, fmt.Errorf("end debate: %w", err)
		}

		if _, err := insertDebateEventTx(ctx, tx, insertDebateEventParams{
			debateID:    params.DebateID,
			eventType:   debateEventExtensionPropose,
			side:        &userSide,
			actorUserID: &params.UserID,
			payload: map[string]any{
				"turn_limit": turnLimit,
			},
			createdAt: &now,
		}); err != nil {
			return model.DebateTurn{}, err
		}

	} else {
		const updateDebate = `
			UPDATE debates
			SET turn_count = $2,
				current_turn_side = $3,
				turn_deadline = $4,
				turn_nudge_sent = false
			WHERE id = $1
		`
		if _, err := tx.ExecContext(ctx, updateDebate,
			params.DebateID, nextTurnNumber, nextSide, deadline,
		); err != nil {
			return model.DebateTurn{}, fmt.Errorf("update debate: %w", err)
		}
	}

	if nextTurnNumber < turnLimit {
		var opponentID uuid.UUID
		if userSide == "a" && sideBUser.Valid {
			parsed, err := uuid.Parse(sideBUser.String)
			if err != nil {
				return model.DebateTurn{}, fmt.Errorf("parse side_b_user_id: %w", err)
			}
			opponentID = parsed
		} else if userSide == "b" {
			opponentID = sideAUser
		}

		if opponentID != uuid.Nil {
			if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
				UserID:      opponentID,
				Type:        "your_turn",
				MessageKey:  "notification.debate.turn",
				MessageVars: map[string]any{"topic": topic},
				DebateID:    &params.DebateID,
				TurnNumber:  &rowTurnNumber,
			}); err != nil {
				return model.DebateTurn{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return model.DebateTurn{}, fmt.Errorf("commit tx: %w", err)
	}

	return turn, nil
}
