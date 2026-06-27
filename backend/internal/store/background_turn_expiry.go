package store

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

func (s *Store) ProcessTurnExpiry(ctx context.Context, logger *slog.Logger) {
	// turn_deadline - (window * 0.25) < now AND turn_deadline > now AND turn_nudge_sent = false
	// Equivalently: now is in the last 25% of the window before the deadline
	const q = `
		SELECT id, topic, current_turn_side, side_a_user_id, side_b_user_id
		FROM debates
		WHERE status = 'active'
		  AND turn_deadline IS NOT NULL
		  AND turn_nudge_sent = false
		  AND turn_deadline > now()
		  AND turn_deadline - (CASE time_mode
				WHEN 'marathon' THEN INTERVAL '1 day 18 hours'
				WHEN 'standard' THEN INTERVAL '12 hours'
				WHEN 'rapid'    THEN INTERVAL '3 hours'
				WHEN 'blitz'    THEN INTERVAL '30 minutes'
				ELSE INTERVAL '12 hours'
			END) < now()
	`

	rows, err := s.DB.QueryContext(ctx, q)
	if err != nil {
		logger.Error("ProcessTurnExpiry: query failed", "error", err)
		return
	}
	defer rows.Close()

	var candidates []turnExpiryCandidate
	for rows.Next() {
		var c turnExpiryCandidate
		if err := rows.Scan(&c.debateID, &c.topic, &c.currentTurnSide, &c.sideAUser, &c.sideBUser); err != nil {
			logger.Error("ProcessTurnExpiry: scan failed", "error", err)
			continue
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		logger.Error("ProcessTurnExpiry: rows iteration failed", "error", err)
		return
	}

	for _, c := range candidates {
		if err := s.processOneTurnExpiry(ctx, c); err != nil {
			logger.Error("ProcessTurnExpiry: failed to process debate",
				"debate_id", c.debateID, "error", err)
			continue
		}
		logger.Info("ProcessTurnExpiry: nudge sent", "debate_id", c.debateID)
	}
}

func (s *Store) processOneTurnExpiry(ctx context.Context, c turnExpiryCandidate) error {
	// Determine who holds the current turn
	var currentUserID uuid.UUID
	if c.currentTurnSide == "a" {
		currentUserID = c.sideAUser
	} else {
		currentUserID = c.sideBUser
	}
	if currentUserID == uuid.Nil {
		return nil
	}

	// Send notification
	message := fmt.Sprintf("Your turn is expiring soon in \"%s\"", c.topic)
	if err := s.CreateNotification(ctx, CreateNotificationParams{
		UserID:   currentUserID,
		Type:     "turn_expiring",
		Message:  message,
		DebateID: &c.debateID,
	}); err != nil {
		return err
	}

	// Mark nudge as sent
	_, err := s.DB.ExecContext(ctx,
		`UPDATE debates SET turn_nudge_sent = true WHERE id = $1`, c.debateID)
	return err
}

// expirationCandidate holds a waiting debate that has been waiting too long.
type expirationCandidate struct {
	debateID  uuid.UUID
	topic     string
	sideAUser uuid.UUID
}

// ProcessExpirations finds debates in 'waiting' status for 14+ days and expires them.
