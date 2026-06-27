package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"respond/internal/model"
)

func (s *Store) ConcedeDebate(ctx context.Context, debateID, userID uuid.UUID) (DebateEndResult, error) {
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

	winnerSide := otherSide(p.userSide)
	now := time.Now().UTC()
	outcome := "concession"

	const updateDebate = `
		UPDATE debates
		SET status = 'finished', outcome = $2, winner_side = $3, ended_at = $4, turn_deadline = NULL
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateDebate, debateID, outcome, winnerSide, now); err != nil {
		return DebateEndResult{}, fmt.Errorf("update debate: %w", err)
	}
	if err := closeAllActiveSeatStintsTx(ctx, tx, debateID, "finished", now); err != nil {
		return DebateEndResult{}, err
	}
	event, err := insertDebateEventTx(ctx, tx, insertDebateEventParams{
		debateID:    debateID,
		eventType:   debateEventConceded,
		side:        &p.userSide,
		actorUserID: &userID,
		payload: map[string]any{
			"winner_side": winnerSide,
		},
		createdAt: &now,
	})
	if err != nil {
		return DebateEndResult{}, err
	}

	// Update ELO ratings
	var winnerUserID, loserUserID uuid.UUID
	if winnerSide == "a" {
		winnerUserID = p.sideAUser
		loserUserID = p.sideBUser
	} else {
		winnerUserID = p.sideBUser
		loserUserID = p.sideAUser
	}

	winnerDelta, loserDelta, err := updateRatingsWin(ctx, tx, winnerUserID, loserUserID)
	if err != nil {
		return DebateEndResult{}, err
	}

	sideADelta := loserDelta
	sideBDelta := winnerDelta
	if winnerSide == "a" {
		sideADelta = winnerDelta
		sideBDelta = loserDelta
	}
	if err := setDebateRatingDeltas(ctx, tx, debateID, sideADelta, sideBDelta); err != nil {
		return DebateEndResult{}, err
	}

	message := fmt.Sprintf("Debate ended in \"%s\"", p.topic)
	for _, userID := range []uuid.UUID{p.sideAUser, p.sideBUser} {
		if userID == uuid.Nil {
			continue
		}
		if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
			UserID:   userID,
			Type:     "debate_ended",
			Message:  message,
			DebateID: &debateID,
		}); err != nil {
			return DebateEndResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return DebateEndResult{}, fmt.Errorf("commit tx: %w", err)
	}

	return DebateEndResult{
		DebateID:   debateID,
		Status:     "finished",
		Outcome:    &outcome,
		WinnerSide: &winnerSide,
		EndedAt:    &now,
		Events:     []model.DebateEvent{event},
	}, nil
}
