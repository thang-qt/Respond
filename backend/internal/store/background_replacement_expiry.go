package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

func (s *Store) ProcessReplacementExpiry(ctx context.Context, logger *slog.Logger) []BackgroundDebateUpdate {
	const q = `
		SELECT id, topic, side_a_user_id, COALESCE(side_b_user_id, '00000000-0000-0000-0000-000000000000'),
			open_side
		FROM debates
		WHERE status = 'waiting_replacement'
		  AND turn_deadline IS NOT NULL
		  AND turn_deadline < now()
	`

	rows, err := s.DB.QueryContext(ctx, q)
	if err != nil {
		logger.Error("ProcessReplacementExpiry: query failed", "error", err)
		return nil
	}
	defer rows.Close()

	var candidates []replacementExpiryCandidate
	for rows.Next() {
		var c replacementExpiryCandidate
		if err := rows.Scan(&c.debateID, &c.topic, &c.sideAUser, &c.sideBUser, &c.openSide); err != nil {
			logger.Error("ProcessReplacementExpiry: scan failed", "error", err)
			continue
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		logger.Error("ProcessReplacementExpiry: rows iteration failed", "error", err)
		return nil
	}

	updates := make([]BackgroundDebateUpdate, 0, len(candidates))
	for _, c := range candidates {
		update, err := s.processOneReplacementExpiry(ctx, c)
		if err != nil {
			logger.Error("ProcessReplacementExpiry: failed to process debate",
				"debate_id", c.debateID, "error", err)
			continue
		}
		if update != nil {
			updates = append(updates, *update)
		}
		logger.Info("ProcessReplacementExpiry: debate ended by walkover", "debate_id", c.debateID)
	}

	return updates
}

func (s *Store) processOneReplacementExpiry(ctx context.Context, c replacementExpiryCandidate) (*BackgroundDebateUpdate, error) {
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
	if status != "waiting_replacement" {
		return nil, nil
	}

	// The remaining debater (non-open side) wins by walkover
	winnerSide := otherSide(c.openSide)
	now := time.Now().UTC()

	const updateDebate = `
		UPDATE debates
		SET status = 'finished', outcome = 'walkover', winner_side = $2,
			ended_at = $3, turn_deadline = NULL, open_side = NULL
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateDebate, c.debateID, winnerSide, now); err != nil {
		return nil, fmt.Errorf("update debate: %w", err)
	}
	if err := closeAllActiveSeatStintsTx(ctx, tx, c.debateID, "walkover", now); err != nil {
		return nil, err
	}
	event, err := insertDebateEventTx(ctx, tx, insertDebateEventParams{
		debateID:  c.debateID,
		eventType: debateEventReplaceExpired,
		side:      &c.openSide,
		payload: map[string]any{
			"winner_side": winnerSide,
		},
		createdAt: &now,
	})
	if err != nil {
		return nil, err
	}

	// Determine winner user to notify (the remaining debater)
	var winnerUserID uuid.UUID
	if winnerSide == "a" {
		winnerUserID = c.sideAUser
	} else {
		winnerUserID = c.sideBUser
	}

	if winnerUserID != uuid.Nil {
		message := fmt.Sprintf("No replacement found for \"%s\". You win by walkover.", c.topic)
		if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
			UserID:   winnerUserID,
			Type:     "debate_ended",
			Message:  message,
			DebateID: &c.debateID,
		}); err != nil {
			return nil, err
		}
	}

	// Note: No ELO update here because the resign penalty was already applied.
	// The remaining debater gets no ELO change.

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	outcome := "walkover"
	return &BackgroundDebateUpdate{
		DebateID:   c.debateID,
		Event:      event,
		Outcome:    &outcome,
		WinnerSide: &winnerSide,
		EndedAt:    &now,
	}, nil
}

// extensionExpiryCandidate holds a debate whose extension deadline has passed.
type extensionExpiryCandidate struct {
	debateID  uuid.UUID
	topic     string
	sideAUser uuid.UUID
	sideBUser sql.NullString
}

// ProcessExtensionExpiry finds debates in 'pending_extension' where the
// extension_deadline has passed and ends them as draws.
