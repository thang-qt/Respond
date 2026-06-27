package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

var ErrRevealAlreadyChosen = errors.New("reveal already chosen")

type RevealResult struct {
	DebateID uuid.UUID          `json:"debate_id"`
	Side     string             `json:"side"`
	Revealed bool               `json:"revealed"`
	User     *model.UserSummary `json:"user"`
}

func (s *Store) RevealDebateIdentity(ctx context.Context, debateID, userID uuid.UUID, reveal bool) (RevealResult, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return RevealResult{}, fmt.Errorf("begin reveal tx: %w", err)
	}
	defer tx.Rollback()

	const query = `
		SELECT status, side_a_user_id, side_b_user_id, side_a_revealed, side_b_revealed
		FROM debates
		WHERE id = $1
		FOR UPDATE
	`

	var (
		status      string
		sideAUserID uuid.UUID
		sideBUserID uuid.UUID
		sideAReveal sql.NullBool
		sideBReveal sql.NullBool
	)

	if err := tx.QueryRowContext(ctx, query, debateID).Scan(&status, &sideAUserID, &sideBUserID, &sideAReveal, &sideBReveal); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RevealResult{}, ErrNotFound
		}
		return RevealResult{}, fmt.Errorf("get debate for reveal: %w", err)
	}

	if status != "finished" {
		return RevealResult{}, ErrDebateNotFinished
	}

	var side string
	if userID == sideAUserID {
		side = "a"
		if sideAReveal.Valid {
			return RevealResult{}, ErrRevealAlreadyChosen
		}
	} else if userID == sideBUserID {
		side = "b"
		if sideBReveal.Valid {
			return RevealResult{}, ErrRevealAlreadyChosen
		}
	} else {
		return RevealResult{}, ErrDebateNotParticipant
	}

	if side == "a" {
		if _, err := tx.ExecContext(ctx, `UPDATE debates SET side_a_revealed = $2 WHERE id = $1`, debateID, reveal); err != nil {
			return RevealResult{}, fmt.Errorf("update side a reveal: %w", err)
		}
	} else {
		if _, err := tx.ExecContext(ctx, `UPDATE debates SET side_b_revealed = $2 WHERE id = $1`, debateID, reveal); err != nil {
			return RevealResult{}, fmt.Errorf("update side b reveal: %w", err)
		}
	}

	var user *model.UserSummary
	if reveal {
		var username string
		var rating int
		if err := tx.QueryRowContext(ctx, `SELECT username, rating FROM users WHERE id = $1`, userID).Scan(&username, &rating); err != nil {
			return RevealResult{}, fmt.Errorf("get user for reveal: %w", err)
		}
		user = &model.UserSummary{Username: username, Rating: rating}
	}

	if err := tx.Commit(); err != nil {
		return RevealResult{}, fmt.Errorf("commit reveal: %w", err)
	}

	return RevealResult{
		DebateID: debateID,
		Side:     side,
		Revealed: reveal,
		User:     user,
	}, nil
}
