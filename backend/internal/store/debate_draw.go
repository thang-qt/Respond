package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"respond/internal/model"
)

func (s *Store) ProposeDrawDebate(ctx context.Context, debateID, userID uuid.UUID) (DrawProposeResult, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return DrawProposeResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	p, err := fetchDebateForAction(ctx, tx, debateID, userID)
	if err != nil {
		return DrawProposeResult{}, err
	}
	if p.status != "active" {
		return DrawProposeResult{}, ErrDebateNotActive
	}

	// Check cooldown: draw_proposed_by must be null AND if there was a previous
	// draw by this side, it must have been at least 3 turns ago
	const checkCooldown = `
		SELECT draw_proposed_by, draw_turn_number, turn_count
		FROM debates WHERE id = $1
	`
	var (
		drawProposedBy sql.NullString
		drawTurnNumber sql.NullInt64
		turnCount      int
	)
	if err := tx.QueryRowContext(ctx, checkCooldown, debateID).Scan(
		&drawProposedBy, &drawTurnNumber, &turnCount,
	); err != nil {
		return DrawProposeResult{}, fmt.Errorf("check cooldown: %w", err)
	}

	if drawProposedBy.Valid {
		// There's already an active draw proposal
		return DrawProposeResult{}, ErrDrawCooldown
	}

	if drawTurnNumber.Valid && turnCount-int(drawTurnNumber.Int64) < 3 {
		return DrawProposeResult{}, ErrDrawCooldown
	}

	const updateDraw = `
		UPDATE debates
		SET draw_proposed_by = $2, draw_proposed_at = $3, draw_turn_number = $4
		WHERE id = $1
	`
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, updateDraw, debateID, p.userSide, now, turnCount); err != nil {
		return DrawProposeResult{}, fmt.Errorf("update draw: %w", err)
	}
	event, err := insertDebateEventTx(ctx, tx, insertDebateEventParams{
		debateID:    debateID,
		eventType:   debateEventDrawProposed,
		side:        &p.userSide,
		actorUserID: &userID,
		createdAt:   &now,
	})
	if err != nil {
		return DrawProposeResult{}, err
	}

	opponentID := p.sideAUser
	if p.userSide == "a" {
		opponentID = p.sideBUser
	}
	if opponentID != uuid.Nil {
		if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
			UserID:      opponentID,
			Type:        "draw_proposed",
			MessageKey:  "notification.draw.proposed",
			MessageVars: map[string]any{"topic": p.topic},
			DebateID:    &debateID,
		}); err != nil {
			return DrawProposeResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return DrawProposeResult{}, fmt.Errorf("commit tx: %w", err)
	}

	return DrawProposeResult{
		DebateID:   debateID,
		ProposedBy: p.userSide,
		Status:     "pending",
		Events:     []model.DebateEvent{event},
	}, nil
}

func (s *Store) RespondDrawDebate(ctx context.Context, debateID, userID uuid.UUID, accept bool) (DrawRespondResult, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return DrawRespondResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	p, err := fetchDebateForAction(ctx, tx, debateID, userID)
	if err != nil {
		return DrawRespondResult{}, err
	}
	if p.status != "active" {
		return DrawRespondResult{}, ErrDebateNotActive
	}

	// Check there is an active proposal and it's not from this user
	const checkProposal = `
		SELECT draw_proposed_by FROM debates WHERE id = $1
	`
	var drawProposedBy sql.NullString
	if err := tx.QueryRowContext(ctx, checkProposal, debateID).Scan(&drawProposedBy); err != nil {
		return DrawRespondResult{}, fmt.Errorf("check proposal: %w", err)
	}
	if !drawProposedBy.Valid {
		return DrawRespondResult{}, ErrDrawNotProposed
	}
	if drawProposedBy.String == p.userSide {
		return DrawRespondResult{}, ErrDrawSelfRespond
	}

	if accept {
		now := time.Now().UTC()
		outcome := "draw"

		const endDebate = `
			UPDATE debates
			SET status = 'finished', outcome = 'draw', winner_side = NULL,
				ended_at = $2, turn_deadline = NULL,
				draw_proposed_by = NULL, draw_proposed_at = NULL
			WHERE id = $1
		`
		if _, err := tx.ExecContext(ctx, endDebate, debateID, now); err != nil {
			return DrawRespondResult{}, fmt.Errorf("end debate: %w", err)
		}
		if err := closeAllActiveSeatStintsTx(ctx, tx, debateID, "finished", now); err != nil {
			return DrawRespondResult{}, err
		}
		event, err := insertDebateEventTx(ctx, tx, insertDebateEventParams{
			debateID:    debateID,
			eventType:   debateEventDrawAccepted,
			side:        &p.userSide,
			actorUserID: &userID,
			createdAt:   &now,
		})
		if err != nil {
			return DrawRespondResult{}, err
		}

		sideADelta, sideBDelta, err := updateRatingsDraw(ctx, tx, p.sideAUser, p.sideBUser)
		if err != nil {
			return DrawRespondResult{}, err
		}
		if err := setDebateRatingDeltas(ctx, tx, debateID, sideADelta, sideBDelta); err != nil {
			return DrawRespondResult{}, err
		}

		for _, userID := range []uuid.UUID{p.sideAUser, p.sideBUser} {
			if userID == uuid.Nil {
				continue
			}
			if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
				UserID:      userID,
				Type:        "debate_ended",
				MessageKey:  "notification.debate.ended",
				MessageVars: map[string]any{"topic": p.topic},
				DebateID:    &debateID,
			}); err != nil {
				return DrawRespondResult{}, err
			}
		}

		if err := tx.Commit(); err != nil {
			return DrawRespondResult{}, fmt.Errorf("commit tx: %w", err)
		}

		status := "finished"
		return DrawRespondResult{
			DebateID:   debateID,
			Status:     &status,
			Outcome:    &outcome,
			WinnerSide: nil,
			EndedAt:    &now,
			Events:     []model.DebateEvent{event},
		}, nil
	}

	// Declined
	const declineDraw = `
		UPDATE debates
		SET draw_proposed_by = NULL, draw_proposed_at = NULL
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, declineDraw, debateID); err != nil {
		return DrawRespondResult{}, fmt.Errorf("decline draw: %w", err)
	}
	now := time.Now().UTC()
	event, err := insertDebateEventTx(ctx, tx, insertDebateEventParams{
		debateID:    debateID,
		eventType:   debateEventDrawDeclined,
		side:        &p.userSide,
		actorUserID: &userID,
		createdAt:   &now,
	})
	if err != nil {
		return DrawRespondResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return DrawRespondResult{}, fmt.Errorf("commit tx: %w", err)
	}

	declined := "declined"
	return DrawRespondResult{
		DebateID:   debateID,
		DrawStatus: &declined,
		Events:     []model.DebateEvent{event},
	}, nil
}

// Rating update helpers
