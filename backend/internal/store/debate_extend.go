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
	ErrDebateNotPendingExtension = errors.New("debate not in pending_extension status")
	ErrExtensionAlreadyResponded = errors.New("already responded to extension")
)

type ExtendRespondResult struct {
	DebateID   uuid.UUID           `json:"debate_id"`
	Status     string              `json:"status"`
	TurnLimit  *int                `json:"turn_limit,omitempty"`
	Outcome    *string             `json:"outcome,omitempty"`
	WinnerSide *string             `json:"winner_side"`
	EndedAt    *time.Time          `json:"ended_at,omitempty"`
	Events     []model.DebateEvent `json:"-"`
}

func (s *Store) RespondExtension(ctx context.Context, debateID, userID uuid.UUID, accept bool) (ExtendRespondResult, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return ExtendRespondResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const fetchQuery = `
		SELECT status, side_a_user_id, side_b_user_id,
			extension_a_accepted, extension_b_accepted,
			turn_limit, time_mode, topic
		FROM debates
		WHERE id = $1
		FOR UPDATE
	`

	var (
		status       string
		sideAUser    uuid.UUID
		sideBUser    sql.NullString
		extAAccepted sql.NullBool
		extBAccepted sql.NullBool
		turnLimit    int
		timeMode     string
		topic        string
	)

	if err := tx.QueryRowContext(ctx, fetchQuery, debateID).Scan(
		&status, &sideAUser, &sideBUser,
		&extAAccepted, &extBAccepted,
		&turnLimit, &timeMode, &topic,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ExtendRespondResult{}, ErrNotFound
		}
		return ExtendRespondResult{}, fmt.Errorf("fetch debate: %w", err)
	}

	if status != "pending_extension" {
		return ExtendRespondResult{}, ErrDebateNotPendingExtension
	}

	// Determine which side this user is on
	var userSide string
	switch {
	case sideAUser == userID:
		userSide = "a"
	case sideBUser.Valid && sideBUser.String == userID.String():
		userSide = "b"
	default:
		return ExtendRespondResult{}, ErrDebateNotParticipant
	}

	// Check if already responded
	if userSide == "a" && extAAccepted.Valid {
		return ExtendRespondResult{}, ErrExtensionAlreadyResponded
	}
	if userSide == "b" && extBAccepted.Valid {
		return ExtendRespondResult{}, ErrExtensionAlreadyResponded
	}

	// If declining, end the debate as a draw immediately
	if !accept {
		now := time.Now().UTC()
		outcome := "draw"

		const declineQuery = `
			UPDATE debates
			SET status = 'finished',
				outcome = 'draw',
				winner_side = NULL,
				ended_at = $2,
				extension_deadline = NULL
			WHERE id = $1
		`
		if _, err := tx.ExecContext(ctx, declineQuery, debateID, now); err != nil {
			return ExtendRespondResult{}, fmt.Errorf("decline extension: %w", err)
		}
		if err := closeAllActiveSeatStintsTx(ctx, tx, debateID, "finished", now); err != nil {
			return ExtendRespondResult{}, err
		}
		event, err := insertDebateEventTx(ctx, tx, insertDebateEventParams{
			debateID:    debateID,
			eventType:   debateEventExtensionDecline,
			side:        &userSide,
			actorUserID: &userID,
			createdAt:   &now,
		})
		if err != nil {
			return ExtendRespondResult{}, err
		}

		// Update ELO ratings as a draw
		if sideBUser.Valid {
			sideBID, err := uuid.Parse(sideBUser.String)
			if err != nil {
				return ExtendRespondResult{}, fmt.Errorf("parse side_b_user_id: %w", err)
			}
			sideADelta, sideBDelta, err := updateRatingsDraw(ctx, tx, sideAUser, sideBID)
			if err != nil {
				return ExtendRespondResult{}, err
			}
			if err := setDebateRatingDeltas(ctx, tx, debateID, sideADelta, sideBDelta); err != nil {
				return ExtendRespondResult{}, err
			}
		}

		// Notify both participants
		message := fmt.Sprintf("Extension declined in \"%s\". The debate ended in a draw.", topic)
		for _, uid := range []uuid.UUID{sideAUser} {
			if uid != userID {
				_ = s.createNotificationTx(ctx, tx, CreateNotificationParams{
					UserID:   uid,
					Type:     "debate_ended",
					Message:  message,
					DebateID: &debateID,
				})
			}
		}
		if sideBUser.Valid {
			sideBID, _ := uuid.Parse(sideBUser.String)
			if sideBID != userID {
				_ = s.createNotificationTx(ctx, tx, CreateNotificationParams{
					UserID:   sideBID,
					Type:     "debate_ended",
					Message:  message,
					DebateID: &debateID,
				})
			}
		}

		if err := tx.Commit(); err != nil {
			return ExtendRespondResult{}, fmt.Errorf("commit tx: %w", err)
		}

		return ExtendRespondResult{
			DebateID:   debateID,
			Status:     "finished",
			Outcome:    &outcome,
			WinnerSide: nil,
			EndedAt:    &now,
			Events:     []model.DebateEvent{event},
		}, nil
	}

	// Accepting: record the acceptance
	if userSide == "a" {
		const q = `UPDATE debates SET extension_a_accepted = true WHERE id = $1`
		if _, err := tx.ExecContext(ctx, q, debateID); err != nil {
			return ExtendRespondResult{}, fmt.Errorf("accept extension (a): %w", err)
		}
		extAAccepted = sql.NullBool{Bool: true, Valid: true}
	} else {
		const q = `UPDATE debates SET extension_b_accepted = true WHERE id = $1`
		if _, err := tx.ExecContext(ctx, q, debateID); err != nil {
			return ExtendRespondResult{}, fmt.Errorf("accept extension (b): %w", err)
		}
		extBAccepted = sql.NullBool{Bool: true, Valid: true}
	}

	// Check if both sides have now accepted
	bothAccepted := extAAccepted.Valid && extAAccepted.Bool && extBAccepted.Valid && extBAccepted.Bool
	acceptedAt := time.Now().UTC()
	event, err := insertDebateEventTx(ctx, tx, insertDebateEventParams{
		debateID:    debateID,
		eventType:   debateEventExtensionAccept,
		side:        &userSide,
		actorUserID: &userID,
		createdAt:   &acceptedAt,
	})
	if err != nil {
		return ExtendRespondResult{}, err
	}

	if bothAccepted {
		newLimit := turnLimit + 5
		now := time.Now().UTC()
		deadline := now.Add(turnWindow(timeMode))

		const extendQuery = `
			UPDATE debates
			SET status = 'active',
				turn_limit = $2,
				turn_deadline = $3,
				extension_deadline = NULL,
				extension_a_accepted = NULL,
				extension_b_accepted = NULL
			WHERE id = $1
		`
		if _, err := tx.ExecContext(ctx, extendQuery, debateID, newLimit, deadline); err != nil {
			return ExtendRespondResult{}, fmt.Errorf("extend debate: %w", err)
		}

		// Notify both participants
		message := fmt.Sprintf("Extension accepted in \"%s\". The debate continues!", topic)
		for _, uid := range []uuid.UUID{sideAUser} {
			if uid != userID {
				_ = s.createNotificationTx(ctx, tx, CreateNotificationParams{
					UserID:   uid,
					Type:     "your_turn",
					Message:  message,
					DebateID: &debateID,
				})
			}
		}
		if sideBUser.Valid {
			sideBID, _ := uuid.Parse(sideBUser.String)
			if sideBID != userID {
				_ = s.createNotificationTx(ctx, tx, CreateNotificationParams{
					UserID:   sideBID,
					Type:     "your_turn",
					Message:  message,
					DebateID: &debateID,
				})
			}
		}

		if err := tx.Commit(); err != nil {
			return ExtendRespondResult{}, fmt.Errorf("commit tx: %w", err)
		}

		return ExtendRespondResult{
			DebateID:  debateID,
			Status:    "active",
			TurnLimit: &newLimit,
			Events:    []model.DebateEvent{event},
		}, nil
	}

	// Only one side accepted so far — notify the other participant.
	message := fmt.Sprintf("Your opponent agreed to extend \"%s\". Accept to continue!", topic)
	var opponentID uuid.UUID
	if userSide == "a" && sideBUser.Valid {
		opponentID, _ = uuid.Parse(sideBUser.String)
	} else if userSide == "b" {
		opponentID = sideAUser
	}
	if opponentID != uuid.Nil {
		_ = s.createNotificationTx(ctx, tx, CreateNotificationParams{
			UserID:   opponentID,
			Type:     "extension_proposed",
			Message:  message,
			DebateID: &debateID,
		})
	}

	if err := tx.Commit(); err != nil {
		return ExtendRespondResult{}, fmt.Errorf("commit tx: %w", err)
	}

	return ExtendRespondResult{
		DebateID: debateID,
		Status:   "pending_extension",
		Events:   []model.DebateEvent{event},
	}, nil
}
