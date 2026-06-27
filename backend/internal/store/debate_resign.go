package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"respond/internal/model"
)

func (s *Store) ResignDebate(ctx context.Context, debateID, userID uuid.UUID) (DebateEndResult, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return DebateEndResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	p, err := fetchDebateForAction(ctx, tx, debateID, userID)
	if err != nil {
		return DebateEndResult{}, err
	}
	if p.status != "active" {
		return DebateEndResult{}, ErrDebateNotActive
	}

	openSide := p.userSide
	now := time.Now().UTC()
	replacementDeadline := now.Add(7 * 24 * time.Hour)

	const updateDebate = `
		UPDATE debates
		SET status = 'waiting_replacement',
			open_side = $2,
			resigned_user_id = $3,
			turn_deadline = $4,
			current_turn_side = $2
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateDebate, debateID, openSide, userID, replacementDeadline); err != nil {
		return DebateEndResult{}, fmt.Errorf("update debate: %w", err)
	}
	if _, err := closeActiveSeatStintTx(ctx, tx, debateID, openSide, "resigned", now); err != nil {
		return DebateEndResult{}, err
	}
	resignerAnon := p.sideAAnon
	if openSide == "b" {
		resignerAnon = p.sideBAnon
	}
	event, err := insertDebateEventTx(ctx, tx, insertDebateEventParams{
		debateID:    debateID,
		eventType:   debateEventSeatOpened,
		side:        &openSide,
		actorUserID: &userID,
		payload: map[string]any{
			"anonymous_id": resignerAnon,
		},
		createdAt: &now,
	})
	if err != nil {
		return DebateEndResult{}, err
	}

	// Apply resign penalty (25% of normal loss)
	var opponentID uuid.UUID
	if p.userSide == "a" {
		opponentID = p.sideBUser
	} else {
		opponentID = p.sideAUser
	}
	if _, err := updateRatingResign(ctx, tx, userID, opponentID); err != nil {
		return DebateEndResult{}, err
	}

	followers, err := listFollowerIDsTx(ctx, tx, debateID)
	if err != nil {
		return DebateEndResult{}, err
	}

	if len(followers) > 0 {
		message := fmt.Sprintf("A seat opened in \"%s\"", p.topic)
		for _, followerID := range followers {
			if followerID == p.sideAUser || followerID == p.sideBUser || followerID == userID {
				continue
			}
			if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
				UserID:   followerID,
				Type:     "seat_open",
				Message:  message,
				DebateID: &debateID,
			}); err != nil {
				return DebateEndResult{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return DebateEndResult{}, fmt.Errorf("commit tx: %w", err)
	}

	status := "waiting_replacement"
	return DebateEndResult{
		DebateID:   debateID,
		Status:     status,
		Outcome:    nil,
		WinnerSide: &openSide, // reuse as open_side in response
		Events:     []model.DebateEvent{event},
	}, nil
}
