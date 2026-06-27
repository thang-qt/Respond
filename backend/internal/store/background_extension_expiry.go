package store

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

func (s *Store) ProcessExtensionExpiry(ctx context.Context, logger *slog.Logger) []BackgroundDebateUpdate {
	const q = `
		SELECT id, topic, side_a_user_id, side_b_user_id
		FROM debates
		WHERE status = 'pending_extension'
		  AND extension_deadline IS NOT NULL
		  AND extension_deadline < now()
	`

	rows, err := s.DB.QueryContext(ctx, q)
	if err != nil {
		logger.Error("ProcessExtensionExpiry: query failed", "error", err)
		return nil
	}
	defer rows.Close()

	var candidates []extensionExpiryCandidate
	for rows.Next() {
		var c extensionExpiryCandidate
		if err := rows.Scan(&c.debateID, &c.topic, &c.sideAUser, &c.sideBUser); err != nil {
			logger.Error("ProcessExtensionExpiry: scan failed", "error", err)
			continue
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		logger.Error("ProcessExtensionExpiry: rows iteration failed", "error", err)
		return nil
	}

	updates := make([]BackgroundDebateUpdate, 0, len(candidates))
	for _, c := range candidates {
		update, err := s.processOneExtensionExpiry(ctx, c)
		if err != nil {
			logger.Error("ProcessExtensionExpiry: failed to process debate",
				"debate_id", c.debateID, "error", err)
			continue
		}
		if update != nil {
			updates = append(updates, *update)
		}
		logger.Info("ProcessExtensionExpiry: debate ended as draw", "debate_id", c.debateID)
	}

	return updates
}

func (s *Store) processOneExtensionExpiry(ctx context.Context, c extensionExpiryCandidate) (*BackgroundDebateUpdate, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Re-check under lock
	var status string
	if err := tx.QueryRowContext(ctx,
		`SELECT status FROM debates WHERE id = $1 FOR UPDATE`, c.debateID,
	).Scan(&status); err != nil {
		return nil, fmt.Errorf("re-fetch: %w", err)
	}
	if status != "pending_extension" {
		return nil, nil
	}

	now := time.Now().UTC()
	const updateDebate = `
		UPDATE debates
		SET status = 'finished', outcome = 'draw', winner_side = NULL,
			ended_at = $2, extension_deadline = NULL, turn_deadline = NULL
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateDebate, c.debateID, now); err != nil {
		return nil, fmt.Errorf("update debate: %w", err)
	}
	if err := closeAllActiveSeatStintsTx(ctx, tx, c.debateID, "finished", now); err != nil {
		return nil, err
	}
	event, err := insertDebateEventTx(ctx, tx, insertDebateEventParams{
		debateID:  c.debateID,
		eventType: debateEventExtendExpired,
		createdAt: &now,
	})
	if err != nil {
		return nil, err
	}

	// Update ELO for both as draw
	if c.sideBUser.Valid {
		sideBID, err := uuid.Parse(c.sideBUser.String)
		if err != nil {
			return nil, fmt.Errorf("parse side_b_user_id: %w", err)
		}
		sideADelta, sideBDelta, err := updateRatingsDraw(ctx, tx, c.sideAUser, sideBID)
		if err != nil {
			return nil, err
		}
		if err := setDebateRatingDeltas(ctx, tx, c.debateID, sideADelta, sideBDelta); err != nil {
			return nil, err
		}
	}

	// Notify both participants
	message := fmt.Sprintf("Extension expired in \"%s\". The debate ended in a draw.", c.topic)
	if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
		UserID:   c.sideAUser,
		Type:     "debate_ended",
		Message:  message,
		DebateID: &c.debateID,
	}); err != nil {
		return nil, err
	}
	if c.sideBUser.Valid {
		sideBID, _ := uuid.Parse(c.sideBUser.String)
		if sideBID != uuid.Nil {
			if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
				UserID:   sideBID,
				Type:     "debate_ended",
				Message:  message,
				DebateID: &c.debateID,
			}); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	outcome := "draw"
	return &BackgroundDebateUpdate{
		DebateID:   c.debateID,
		Event:      event,
		Outcome:    &outcome,
		WinnerSide: nil,
		EndedAt:    &now,
	}, nil
}
