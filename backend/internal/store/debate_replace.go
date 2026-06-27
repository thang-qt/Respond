package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"respond/internal/model"
)

var (
	ErrDebateNotWaitingReplacement = errors.New("debate not in waiting_replacement status")
	ErrDebateIsRemainingDebater    = errors.New("user is the remaining debater")
	ErrDebateIsResignedDebater     = errors.New("user is the debater who resigned")
)

type ReplaceDebateResult struct {
	DebateID     uuid.UUID           `json:"debate_id"`
	Side         string              `json:"side"`
	AnonymousID  string              `json:"anonymous_id"`
	Status       string              `json:"status"`
	TurnDeadline *time.Time          `json:"turn_deadline"`
	Events       []model.DebateEvent `json:"-"`
}

func (s *Store) ReplaceDebate(ctx context.Context, debateID, userID uuid.UUID) (ReplaceDebateResult, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return ReplaceDebateResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const fetchQuery = `
		SELECT status, side_a_user_id, side_b_user_id, open_side, resigned_user_id, time_mode, topic, current_turn_side
		FROM debates
		WHERE id = $1
		FOR UPDATE
	`

	var (
		status          string
		sideAUser       uuid.UUID
		sideBUser       sql.NullString
		openSide        sql.NullString
		resignedUser    sql.NullString
		timeMode        string
		topic           string
		currentTurnSide string
	)

	if err := tx.QueryRowContext(ctx, fetchQuery, debateID).Scan(
		&status, &sideAUser, &sideBUser, &openSide, &resignedUser, &timeMode, &topic, &currentTurnSide,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ReplaceDebateResult{}, ErrNotFound
		}
		return ReplaceDebateResult{}, fmt.Errorf("fetch debate: %w", err)
	}

	if status != "waiting_replacement" {
		return ReplaceDebateResult{}, ErrDebateNotWaitingReplacement
	}

	if !openSide.Valid {
		return ReplaceDebateResult{}, fmt.Errorf("debate in waiting_replacement but open_side is null")
	}

	side := openSide.String

	// Check user is not the remaining debater
	var remainingUserID uuid.UUID
	if side == "a" {
		// Side A is open, Side B is remaining
		if sideBUser.Valid {
			parsed, err := uuid.Parse(sideBUser.String)
			if err != nil {
				return ReplaceDebateResult{}, fmt.Errorf("parse side_b_user_id: %w", err)
			}
			remainingUserID = parsed
		}
	} else {
		// Side B is open, Side A is remaining
		remainingUserID = sideAUser
	}

	if userID == remainingUserID {
		return ReplaceDebateResult{}, ErrDebateIsRemainingDebater
	}

	// Check user is not the debater who resigned
	if resignedUser.Valid {
		parsed, err := uuid.Parse(resignedUser.String)
		if err != nil {
			return ReplaceDebateResult{}, fmt.Errorf("parse resigned_user_id: %w", err)
		}
		if userID == parsed {
			return ReplaceDebateResult{}, ErrDebateIsResignedDebater
		}
	}

	blocked, err := isEitherUserBlockedTx(ctx, tx, userID, remainingUserID)
	if err != nil {
		return ReplaceDebateResult{}, err
	}
	if blocked {
		return ReplaceDebateResult{}, ErrDebateUserBlocked
	}

	// Generate new anonymous ID for the open side
	prefix := strings.ToUpper(side[:1])
	anonID, err := generateAnonymousID(prefix)
	if err != nil {
		return ReplaceDebateResult{}, fmt.Errorf("generate anonymous id: %w", err)
	}

	now := time.Now().UTC()
	deadline := now.Add(turnWindow(timeMode))

	// Update debate: set the new user on the open side, reset to active
	var updateQuery string
	if side == "a" {
		updateQuery = `
			UPDATE debates
			SET status = 'active',
				side_a_user_id = $2,
				side_a_anonymous_id = $3,
				turn_deadline = $4,
				open_side = NULL,
				resigned_user_id = NULL,
				current_turn_side = $5
			WHERE id = $1
		`
	} else {
		updateQuery = `
			UPDATE debates
			SET status = 'active',
				side_b_user_id = $2,
				side_b_anonymous_id = $3,
				turn_deadline = $4,
				open_side = NULL,
				resigned_user_id = NULL,
				current_turn_side = $5
			WHERE id = $1
		`
	}

	if _, err := tx.ExecContext(ctx, updateQuery, debateID, userID, anonID, deadline, currentTurnSide); err != nil {
		return ReplaceDebateResult{}, fmt.Errorf("update debate: %w", err)
	}

	stintID, err := createSeatStintTx(ctx, tx, debateID, side, userID, anonID, now)
	if err != nil {
		return ReplaceDebateResult{}, err
	}
	if err := linkReplacementStintTx(ctx, tx, debateID, side, stintID); err != nil {
		return ReplaceDebateResult{}, err
	}
	event, err := insertDebateEventTx(ctx, tx, insertDebateEventParams{
		debateID:    debateID,
		eventType:   debateEventReplacementJoin,
		side:        &side,
		actorUserID: &userID,
		payload: map[string]any{
			"anonymous_id": anonID,
		},
		createdAt: &now,
	})
	if err != nil {
		return ReplaceDebateResult{}, err
	}

	// Notify the remaining debater
	message := fmt.Sprintf("A replacement debater joined \"%s\"", topic)
	if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
		UserID:   remainingUserID,
		Type:     "replacement_joined",
		Message:  message,
		DebateID: &debateID,
	}); err != nil {
		return ReplaceDebateResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return ReplaceDebateResult{}, fmt.Errorf("commit tx: %w", err)
	}

	return ReplaceDebateResult{
		DebateID:     debateID,
		Side:         side,
		AnonymousID:  anonID,
		Status:       "active",
		TurnDeadline: &deadline,
		Events:       []model.DebateEvent{event},
	}, nil
}
