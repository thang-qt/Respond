package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

func (s *Store) ProcessExpirations(ctx context.Context, logger *slog.Logger) {
	const q = `
		SELECT id, topic, side_a_user_id
		FROM debates
		WHERE status = 'waiting'
		  AND invited_user_id IS NULL
		  AND created_at + INTERVAL '14 days' < now()
	`

	rows, err := s.DB.QueryContext(ctx, q)
	if err != nil {
		logger.Error("ProcessExpirations: query failed", "error", err)
		return
	}
	defer rows.Close()

	var candidates []expirationCandidate
	for rows.Next() {
		var c expirationCandidate
		if err := rows.Scan(&c.debateID, &c.topic, &c.sideAUser); err != nil {
			logger.Error("ProcessExpirations: scan failed", "error", err)
			continue
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		logger.Error("ProcessExpirations: rows iteration failed", "error", err)
		return
	}

	for _, c := range candidates {
		if err := s.processOneExpiration(ctx, c); err != nil {
			logger.Error("ProcessExpirations: failed to process debate",
				"debate_id", c.debateID, "error", err)
			continue
		}
		logger.Info("ProcessExpirations: debate expired", "debate_id", c.debateID)
	}
}

// ProcessChallengeExpirations expires pending invite-only challenges after their deadline.
func (s *Store) ProcessChallengeExpirations(ctx context.Context, logger *slog.Logger) {
	const q = `
		SELECT id, topic, side_a_user_id
		FROM debates
		WHERE status = 'waiting'
		  AND invited_user_id IS NOT NULL
		  AND COALESCE(challenge_expires_at, created_at + INTERVAL '7 days') < now()
	`

	rows, err := s.DB.QueryContext(ctx, q)
	if err != nil {
		logger.Error("ProcessChallengeExpirations: query failed", "error", err)
		return
	}
	defer rows.Close()

	var candidates []expirationCandidate
	for rows.Next() {
		var c expirationCandidate
		if err := rows.Scan(&c.debateID, &c.topic, &c.sideAUser); err != nil {
			logger.Error("ProcessChallengeExpirations: scan failed", "error", err)
			continue
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		logger.Error("ProcessChallengeExpirations: rows iteration failed", "error", err)
		return
	}

	for _, c := range candidates {
		if err := s.processOneChallengeExpiration(ctx, c); err != nil {
			logger.Error("ProcessChallengeExpirations: failed to process debate", "debate_id", c.debateID, "error", err)
			continue
		}
		logger.Info("ProcessChallengeExpirations: challenge expired", "debate_id", c.debateID)
	}
}

func (s *Store) processOneChallengeExpiration(ctx context.Context, c expirationCandidate) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var status string
	var invitedUser sql.NullString
	if err := tx.QueryRowContext(ctx,
		`SELECT status, invited_user_id FROM debates WHERE id = $1 FOR UPDATE`, c.debateID,
	).Scan(&status, &invitedUser); err != nil {
		return fmt.Errorf("re-fetch challenge: %w", err)
	}
	if status != "waiting" || !invitedUser.Valid {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `UPDATE debates SET status = 'expired' WHERE id = $1`, c.debateID); err != nil {
		return fmt.Errorf("expire challenge debate: %w", err)
	}

	if invitedUser.Valid {
		if _, err := tx.ExecContext(ctx, `
			UPDATE notifications
			SET is_read = true
			WHERE user_id = $1
			  AND debate_id = $2
			  AND type = 'challenge_received'
			  AND is_read = false
		`, invitedUser.String, c.debateID); err != nil {
			return fmt.Errorf("mark challenge notification read: %w", err)
		}
	}

	message := fmt.Sprintf("Your challenge for \"%s\" expired.", c.topic)
	if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
		UserID:   c.sideAUser,
		Type:     "challenge_expired",
		Message:  message,
		DebateID: &c.debateID,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) processOneExpiration(ctx context.Context, c expirationCandidate) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Re-check under lock
	var status string
	if err := tx.QueryRowContext(ctx,
		`SELECT status FROM debates WHERE id = $1 FOR UPDATE`, c.debateID,
	).Scan(&status); err != nil {
		return fmt.Errorf("re-fetch: %w", err)
	}
	if status != "waiting" {
		return nil
	}

	const updateDebate = `
		UPDATE debates SET status = 'expired' WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateDebate, c.debateID); err != nil {
		return fmt.Errorf("update debate: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE notifications
		SET is_read = true
		WHERE debate_id = $1
		  AND type = 'debate_invited'
		  AND is_read = false
	`, c.debateID); err != nil {
		return fmt.Errorf("mark invite notifications read: %w", err)
	}

	// Notify creator
	message := fmt.Sprintf("Your debate \"%s\" expired. No one joined.", c.topic)
	if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
		UserID:   c.sideAUser,
		Type:     "debate_ended",
		Message:  message,
		DebateID: &c.debateID,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

// replacementExpiryCandidate holds a debate waiting for replacement that's timed out.
type replacementExpiryCandidate struct {
	debateID  uuid.UUID
	topic     string
	sideAUser uuid.UUID
	sideBUser uuid.UUID
	openSide  string
}

// ProcessReplacementExpiry finds debates in 'waiting_replacement' where
// turn_deadline (set to 7 days after resign) has passed, and ends them
// as walkovers for the remaining debater.
