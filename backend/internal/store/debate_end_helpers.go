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
	ErrDebateNotParticipant = errors.New("not a participant")
	ErrDrawCooldown         = errors.New("draw proposed too recently")
	ErrDrawNotProposed      = errors.New("no active draw proposal")
	ErrDrawSelfRespond      = errors.New("cannot respond to own draw proposal")
)

type DebateEndResult struct {
	DebateID   uuid.UUID           `json:"debate_id"`
	Status     string              `json:"status"`
	Outcome    *string             `json:"outcome"`
	WinnerSide *string             `json:"winner_side"`
	EndedAt    *time.Time          `json:"ended_at,omitempty"`
	Events     []model.DebateEvent `json:"-"`
}

type DrawProposeResult struct {
	DebateID   uuid.UUID           `json:"debate_id"`
	ProposedBy string              `json:"proposed_by"`
	Status     string              `json:"status"`
	Events     []model.DebateEvent `json:"-"`
}

type DrawRespondResult struct {
	DebateID   uuid.UUID           `json:"debate_id"`
	DrawStatus *string             `json:"draw_status,omitempty"`
	Status     *string             `json:"status,omitempty"`
	Outcome    *string             `json:"outcome,omitempty"`
	WinnerSide *string             `json:"winner_side"`
	EndedAt    *time.Time          `json:"ended_at,omitempty"`
	Events     []model.DebateEvent `json:"-"`
}

// debateParticipant holds the info needed to determine a user's side in a debate.
type debateParticipant struct {
	status    string
	sideAUser uuid.UUID
	sideBUser uuid.UUID
	sideAAnon string
	sideBAnon string
	userSide  string
	topic     string
}

// fetchDebateForAction fetches a debate in a transaction and validates the user is a participant.
func fetchDebateForAction(ctx context.Context, tx *sql.Tx, debateID, userID uuid.UUID) (debateParticipant, error) {
	const q = `
		SELECT status, side_a_user_id, side_b_user_id, side_a_anonymous_id, side_b_anonymous_id, topic
		FROM debates
		WHERE id = $1
		FOR UPDATE
	`
	var (
		status    string
		sideAUser uuid.UUID
		sideBUser sql.NullString
		sideAAnon string
		sideBAnon sql.NullString
		topic     string
	)
	if err := tx.QueryRowContext(ctx, q, debateID).Scan(&status, &sideAUser, &sideBUser, &sideAAnon, &sideBAnon, &topic); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return debateParticipant{}, ErrNotFound
		}
		return debateParticipant{}, fmt.Errorf("fetch debate: %w", err)
	}

	var userSide string
	var sideBUUID uuid.UUID
	switch {
	case sideAUser == userID:
		userSide = "a"
	case sideBUser.Valid:
		parsed, err := uuid.Parse(sideBUser.String)
		if err != nil {
			return debateParticipant{}, fmt.Errorf("parse side_b_user_id: %w", err)
		}
		sideBUUID = parsed
		if parsed == userID {
			userSide = "b"
		}
	}

	if userSide == "" {
		return debateParticipant{}, ErrDebateNotParticipant
	}

	return debateParticipant{
		status:    status,
		sideAUser: sideAUser,
		sideBUser: sideBUUID,
		sideAAnon: sideAAnon,
		sideBAnon: sideBAnon.String,
		userSide:  userSide,
		topic:     topic,
	}, nil
}

func otherSide(side string) string {
	if side == "a" {
		return "b"
	}
	return "a"
}
