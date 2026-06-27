package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDebateNotWaiting          = errors.New("debate not in waiting status")
	ErrDebateOwnDebate           = errors.New("cannot join own debate")
	ErrDebateFull                = errors.New("debate already has two sides")
	ErrDebateUserBlocked         = errors.New("debate users are blocked")
	ErrDebateChallengeOnly       = errors.New("debate is challenge only")
	ErrDebateChallengeNotInvited = errors.New("not invited to challenge")
	ErrDebateChallengeExpired    = errors.New("challenge expired")
	ErrDebateChallengeResponded  = errors.New("challenge already responded")
)

// turnWindow returns the turn duration for a given time mode.
func turnWindow(timeMode string) time.Duration {
	switch timeMode {
	case "marathon":
		return 7 * 24 * time.Hour
	case "standard":
		return 48 * time.Hour
	case "rapid":
		return 12 * time.Hour
	case "blitz":
		return 2 * time.Hour
	default:
		return 48 * time.Hour
	}
}

type JoinDebateResult struct {
	DebateID     uuid.UUID  `json:"debate_id"`
	Side         string     `json:"side"`
	AnonymousID  string     `json:"anonymous_id"`
	Status       string     `json:"status"`
	TurnDeadline *time.Time `json:"turn_deadline"`
}

func (s *Store) JoinDebate(ctx context.Context, debateID, userID uuid.UUID) (JoinDebateResult, error) {
	// Fetch debate state in a single query
	const fetchQuery = `
		SELECT status, side_a_user_id, side_b_user_id, time_mode, topic,
			invited_user_id
		FROM debates
		WHERE id = $1
		FOR UPDATE
	`

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return JoinDebateResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var (
		status    string
		sideAUser uuid.UUID
		sideBUser sql.NullString
		timeMode  string
		topic     string
		invitedID sql.NullString
	)

	if err := tx.QueryRowContext(ctx, fetchQuery, debateID).Scan(
		&status, &sideAUser, &sideBUser, &timeMode, &topic, &invitedID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return JoinDebateResult{}, ErrNotFound
		}
		return JoinDebateResult{}, fmt.Errorf("fetch debate: %w", err)
	}

	if status != "waiting" {
		return JoinDebateResult{}, ErrDebateNotWaiting
	}

	if sideAUser == userID {
		return JoinDebateResult{}, ErrDebateOwnDebate
	}

	if invitedID.Valid {
		return JoinDebateResult{}, ErrDebateChallengeOnly
	}

	if sideBUser.Valid {
		return JoinDebateResult{}, ErrDebateFull
	}

	blocked, err := isEitherUserBlockedTx(ctx, tx, userID, sideAUser)
	if err != nil {
		return JoinDebateResult{}, err
	}
	if blocked {
		return JoinDebateResult{}, ErrDebateUserBlocked
	}

	anonID, err := generateAnonymousID("B")
	if err != nil {
		return JoinDebateResult{}, fmt.Errorf("generate anonymous id: %w", err)
	}

	now := time.Now().UTC()
	deadline := now.Add(turnWindow(timeMode))

	const updateQuery = `
		UPDATE debates
		SET status = 'active',
			side_b_user_id = $2,
			side_b_anonymous_id = $3,
			started_at = $4,
			turn_deadline = $5,
			current_turn_side = 'b'
		WHERE id = $1
	`

	if _, err := tx.ExecContext(ctx, updateQuery,
		debateID, userID, anonID, now, deadline,
	); err != nil {
		return JoinDebateResult{}, fmt.Errorf("update debate: %w", err)
	}

	if _, err := createSeatStintTx(ctx, tx, debateID, "b", userID, anonID, now); err != nil {
		return JoinDebateResult{}, err
	}

	if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
		UserID:      sideAUser,
		Type:        "debate_joined",
		MessageKey:  "notification.debate.joined",
		MessageVars: map[string]any{"topic": topic},
		DebateID:    &debateID,
	}); err != nil {
		return JoinDebateResult{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE notifications
		SET is_read = true
		WHERE debate_id = $1
		  AND type = 'debate_invited'
		  AND is_read = false
	`, debateID); err != nil {
		return JoinDebateResult{}, fmt.Errorf("mark invite notifications read: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return JoinDebateResult{}, fmt.Errorf("commit tx: %w", err)
	}

	return JoinDebateResult{
		DebateID:     debateID,
		Side:         "b",
		AnonymousID:  anonID,
		Status:       "active",
		TurnDeadline: &deadline,
	}, nil
}
