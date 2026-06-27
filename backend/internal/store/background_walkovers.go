package store

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"respond/internal/model"
)

type BackgroundDebateUpdate struct {
	DebateID   uuid.UUID
	Event      model.DebateEvent
	Outcome    *string
	WinnerSide *string
	EndedAt    *time.Time
}

// walkoverCandidate holds the data needed to process a single walkover.
type walkoverCandidate struct {
	debateID        uuid.UUID
	topic           string
	sideAUser       uuid.UUID
	sideBUser       uuid.UUID
	currentTurnSide string
}

// ProcessWalkovers finds active debates where the turn holder has exceeded
// twice the turn window (deadline + one full window) and ends them as walkovers.
func (s *Store) ProcessWalkovers(ctx context.Context, logger *slog.Logger) []BackgroundDebateUpdate {
	// The turn_deadline column stores when the turn was originally due.
	// A walkover triggers when now > turn_deadline + turn_window (2x total grace).
	// We use a CASE expression matching the time_mode to compute the grace window.
	const q = `
		SELECT id, topic, side_a_user_id, side_b_user_id, current_turn_side
		FROM debates
		WHERE status = 'active'
		  AND turn_deadline IS NOT NULL
		  AND turn_deadline + (CASE time_mode
				WHEN 'marathon' THEN INTERVAL '7 days'
				WHEN 'standard' THEN INTERVAL '48 hours'
				WHEN 'rapid'    THEN INTERVAL '12 hours'
				WHEN 'blitz'    THEN INTERVAL '2 hours'
				ELSE INTERVAL '48 hours'
			END) < now()
	`

	rows, err := s.DB.QueryContext(ctx, q)
	if err != nil {
		logger.Error("ProcessWalkovers: query failed", "error", err)
		return nil
	}
	defer rows.Close()

	var candidates []walkoverCandidate
	for rows.Next() {
		var c walkoverCandidate
		if err := rows.Scan(&c.debateID, &c.topic, &c.sideAUser, &c.sideBUser, &c.currentTurnSide); err != nil {
			logger.Error("ProcessWalkovers: scan failed", "error", err)
			continue
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		logger.Error("ProcessWalkovers: rows iteration failed", "error", err)
		return nil
	}

	updates := make([]BackgroundDebateUpdate, 0, len(candidates))
	for _, c := range candidates {
		update, err := s.processOneWalkover(ctx, c)
		if err != nil {
			logger.Error("ProcessWalkovers: failed to process debate",
				"debate_id", c.debateID, "error", err)
			continue
		}
		if update != nil {
			updates = append(updates, *update)
		}
		logger.Info("ProcessWalkovers: debate ended by walkover", "debate_id", c.debateID)
	}

	return updates
}

func (s *Store) processOneWalkover(ctx context.Context, c walkoverCandidate) (*BackgroundDebateUpdate, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Re-check under lock to avoid race conditions
	var status string
	if err := tx.QueryRowContext(ctx,
		`SELECT status FROM debates WHERE id = $1 FOR UPDATE`, c.debateID,
	).Scan(&status); err != nil {
		return nil, fmt.Errorf("re-fetch: %w", err)
	}
	if status != "active" {
		return nil, nil // already processed
	}

	// The current turn holder is the one who failed to respond
	defaultedSide := c.currentTurnSide
	winnerSide := otherSide(defaultedSide)
	now := time.Now().UTC()

	const updateDebate = `
		UPDATE debates
		SET status = 'finished', outcome = 'walkover', winner_side = $2,
			ended_at = $3, turn_deadline = NULL
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
		eventType: debateEventWalkover,
		side:      &defaultedSide,
		payload: map[string]any{
			"winner_side": winnerSide,
		},
		createdAt: &now,
	})
	if err != nil {
		return nil, err
	}

	// Determine winner/loser user IDs
	var winnerUserID, loserUserID uuid.UUID
	if winnerSide == "a" {
		winnerUserID = c.sideAUser
		loserUserID = c.sideBUser
	} else {
		winnerUserID = c.sideBUser
		loserUserID = c.sideAUser
	}

	winnerDelta, loserDelta, err := updateRatingsWin(ctx, tx, winnerUserID, loserUserID)
	if err != nil {
		return nil, err
	}
	sideADelta := loserDelta
	sideBDelta := winnerDelta
	if winnerSide == "a" {
		sideADelta = winnerDelta
		sideBDelta = loserDelta
	}
	if err := setDebateRatingDeltas(ctx, tx, c.debateID, sideADelta, sideBDelta); err != nil {
		return nil, err
	}

	// Notify both participants
	message := fmt.Sprintf("Debate ended in \"%s\" — walkover", c.topic)
	for _, uid := range []uuid.UUID{c.sideAUser, c.sideBUser} {
		if uid == uuid.Nil {
			continue
		}
		if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
			UserID:   uid,
			Type:     "debate_ended",
			Message:  message,
			DebateID: &c.debateID,
		}); err != nil {
			return nil, err
		}
	}

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

// turnExpiryCandidate holds a debate whose turn is about to expire (75% elapsed).
type turnExpiryCandidate struct {
	debateID        uuid.UUID
	topic           string
	currentTurnSide string
	sideAUser       uuid.UUID
	sideBUser       uuid.UUID
}

// ProcessTurnExpiry finds active debates where 75% of the turn window has elapsed
// but the deadline hasn't passed yet, and sends a nudge notification.
