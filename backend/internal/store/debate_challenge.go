package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidChallengeBox    = errors.New("invalid challenge box")
	ErrInvalidChallengeStatus = errors.New("invalid challenge status")
)

type RespondChallengeResult struct {
	DebateID     uuid.UUID  `json:"debate_id"`
	Accepted     bool       `json:"accepted"`
	Status       string     `json:"status"`
	Side         *string    `json:"side,omitempty"`
	AnonymousID  *string    `json:"anonymous_id,omitempty"`
	TurnDeadline *time.Time `json:"turn_deadline,omitempty"`
}

type RechallengeTarget struct {
	InvitedUserID            uuid.UUID
	ChallengeIdentityVisible bool
}

func (s *Store) ResolveRechallengeInvitedUser(ctx context.Context, sourceDebateID, challengerID uuid.UUID) (RechallengeTarget, error) {
	const query = `
		SELECT status, side_a_user_id, side_b_user_id, side_a_revealed, side_b_revealed
		FROM debates
		WHERE id = $1
	`

	var (
		status   string
		sideAID  uuid.UUID
		sideBRaw sql.NullString
		sideARev sql.NullBool
		sideBRev sql.NullBool
	)

	if err := s.DB.QueryRowContext(ctx, query, sourceDebateID).Scan(&status, &sideAID, &sideBRaw, &sideARev, &sideBRev); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RechallengeTarget{}, ErrNotFound
		}
		return RechallengeTarget{}, fmt.Errorf("resolve rechallenge source debate: %w", err)
	}

	if status != "finished" {
		return RechallengeTarget{}, ErrDebateNotFinished
	}

	if !sideBRaw.Valid {
		return RechallengeTarget{}, ErrDebateNotFinished
	}

	sideBID, err := uuid.Parse(sideBRaw.String)
	if err != nil {
		return RechallengeTarget{}, fmt.Errorf("parse side b user id: %w", err)
	}

	var invitedUserID uuid.UUID
	switch challengerID {
	case sideAID:
		invitedUserID = sideBID
	case sideBID:
		invitedUserID = sideAID
	default:
		return RechallengeTarget{}, ErrDebateNotParticipant
	}

	blocked, err := s.IsEitherUserBlocked(ctx, challengerID, invitedUserID)
	if err != nil {
		return RechallengeTarget{}, err
	}
	if blocked {
		return RechallengeTarget{}, ErrDebateUserBlocked
	}

	identityVisible := sideARev.Valid && sideARev.Bool && sideBRev.Valid && sideBRev.Bool

	return RechallengeTarget{
		InvitedUserID:            invitedUserID,
		ChallengeIdentityVisible: identityVisible,
	}, nil
}

func (s *Store) IsDebateVisibleToViewer(ctx context.Context, debateID uuid.UUID, viewerID *uuid.UUID) (bool, error) {
	const query = `
		SELECT status, side_a_user_id, invited_user_id
		FROM debates
		WHERE id = $1
	`

	var (
		status    string
		sideAUser uuid.UUID
		invitedID sql.NullString
	)

	if err := s.DB.QueryRowContext(ctx, query, debateID).Scan(&status, &sideAUser, &invitedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrNotFound
		}
		return false, fmt.Errorf("is debate visible to viewer: %w", err)
	}

	if !invitedID.Valid {
		return true, nil
	}

	if status != "waiting" && status != "expired" {
		return true, nil
	}

	if viewerID == nil {
		return false, nil
	}

	if sideAUser == *viewerID {
		return true, nil
	}

	return invitedID.String == viewerID.String(), nil
}

func (s *Store) RespondChallenge(ctx context.Context, debateID, userID uuid.UUID, accept bool) (RespondChallengeResult, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return RespondChallengeResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const fetchQuery = `
		SELECT status, topic, side_a_user_id, side_b_user_id, invited_user_id, challenge_expires_at, time_mode
		FROM debates
		WHERE id = $1
		FOR UPDATE
	`

	var (
		status    string
		topic     string
		sideAUser uuid.UUID
		sideBUser sql.NullString
		invitedID sql.NullString
		expiresAt sql.NullTime
		timeMode  string
	)

	if err := tx.QueryRowContext(ctx, fetchQuery, debateID).Scan(
		&status,
		&topic,
		&sideAUser,
		&sideBUser,
		&invitedID,
		&expiresAt,
		&timeMode,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RespondChallengeResult{}, ErrNotFound
		}
		return RespondChallengeResult{}, fmt.Errorf("fetch challenge debate: %w", err)
	}

	if !invitedID.Valid {
		return RespondChallengeResult{}, ErrDebateChallengeOnly
	}

	if invitedID.String != userID.String() {
		return RespondChallengeResult{}, ErrDebateChallengeNotInvited
	}

	if status != "waiting" {
		if sideBUser.Valid {
			return RespondChallengeResult{}, ErrDebateChallengeResponded
		}
		if status == "expired" {
			return RespondChallengeResult{}, ErrDebateChallengeExpired
		}
		return RespondChallengeResult{}, ErrDebateNotWaiting
	}

	now := time.Now().UTC()
	if expiresAt.Valid && now.After(expiresAt.Time) {
		if _, err := tx.ExecContext(ctx, `UPDATE debates SET status = 'expired' WHERE id = $1`, debateID); err != nil {
			return RespondChallengeResult{}, fmt.Errorf("expire stale challenge: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE notifications
			SET is_read = true
			WHERE user_id = $1
			  AND debate_id = $2
			  AND type = 'challenge_received'
			  AND is_read = false
		`, userID, debateID); err != nil {
			return RespondChallengeResult{}, fmt.Errorf("mark challenge notification read: %w", err)
		}
		message := fmt.Sprintf("Your challenge for \"%s\" expired.", topic)
		if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
			UserID:   sideAUser,
			Type:     "challenge_expired",
			Message:  message,
			DebateID: &debateID,
		}); err != nil {
			return RespondChallengeResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return RespondChallengeResult{}, fmt.Errorf("commit tx: %w", err)
		}
		return RespondChallengeResult{}, ErrDebateChallengeExpired
	}

	if !accept {
		if _, err := tx.ExecContext(ctx, `UPDATE debates SET status = 'expired' WHERE id = $1`, debateID); err != nil {
			return RespondChallengeResult{}, fmt.Errorf("decline challenge: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE notifications
			SET is_read = true
			WHERE user_id = $1
			  AND debate_id = $2
			  AND type = 'challenge_received'
			  AND is_read = false
		`, userID, debateID); err != nil {
			return RespondChallengeResult{}, fmt.Errorf("mark challenge notification read: %w", err)
		}
		message := fmt.Sprintf("Your challenge for \"%s\" was declined.", topic)
		if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
			UserID:   sideAUser,
			Type:     "challenge_declined",
			Message:  message,
			DebateID: &debateID,
		}); err != nil {
			return RespondChallengeResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return RespondChallengeResult{}, fmt.Errorf("commit tx: %w", err)
		}
		return RespondChallengeResult{
			DebateID: debateID,
			Accepted: false,
			Status:   "expired",
		}, nil
	}

	blocked, err := isEitherUserBlockedTx(ctx, tx, userID, sideAUser)
	if err != nil {
		return RespondChallengeResult{}, err
	}
	if blocked {
		return RespondChallengeResult{}, ErrDebateUserBlocked
	}

	anonID, err := generateAnonymousID("B")
	if err != nil {
		return RespondChallengeResult{}, fmt.Errorf("generate anonymous id: %w", err)
	}

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
	if _, err := tx.ExecContext(ctx, updateQuery, debateID, userID, anonID, now, deadline); err != nil {
		return RespondChallengeResult{}, fmt.Errorf("accept challenge update: %w", err)
	}

	if _, err := createSeatStintTx(ctx, tx, debateID, "b", userID, anonID, now); err != nil {
		return RespondChallengeResult{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE notifications
		SET is_read = true
		WHERE user_id = $1
		  AND debate_id = $2
		  AND type = 'challenge_received'
		  AND is_read = false
	`, userID, debateID); err != nil {
		return RespondChallengeResult{}, fmt.Errorf("mark challenge notification read: %w", err)
	}

	message := fmt.Sprintf("Your challenge for \"%s\" was accepted.", topic)
	if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
		UserID:   sideAUser,
		Type:     "challenge_accepted",
		Message:  message,
		DebateID: &debateID,
	}); err != nil {
		return RespondChallengeResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return RespondChallengeResult{}, fmt.Errorf("commit tx: %w", err)
	}

	side := "b"
	return RespondChallengeResult{
		DebateID:     debateID,
		Accepted:     true,
		Status:       "active",
		Side:         &side,
		AnonymousID:  &anonID,
		TurnDeadline: &deadline,
	}, nil
}

type ListChallengesParams struct {
	UserID  uuid.UUID
	Box     string
	Status  string
	Page    int
	PerPage int
}

type ChallengeListItem struct {
	DebateID           uuid.UUID  `json:"debate_id"`
	DebateSlug         string     `json:"debate_slug"`
	Topic              string     `json:"topic"`
	TimeMode           string     `json:"time_mode"`
	TurnLimit          int        `json:"turn_limit"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	ChallengeExpiresAt *time.Time `json:"challenge_expires_at"`
	Challenger         string     `json:"challenger_username"`
	InvitedUser        string     `json:"invited_username"`
}

func (s *Store) ListChallenges(ctx context.Context, params ListChallengesParams) ([]ChallengeListItem, int, error) {
	pagination := normalizePagination(params.Page, params.PerPage, 20, 50)
	perPage := pagination.PerPage

	box := strings.TrimSpace(params.Box)
	if box == "" {
		box = "inbox"
	}
	status := strings.TrimSpace(params.Status)
	if status == "" {
		status = "pending"
	}

	whereParts := []string{"d.invited_user_id IS NOT NULL"}
	args := []any{params.UserID}

	switch box {
	case "inbox":
		whereParts = append(whereParts, "d.invited_user_id = $1")
	case "outbox":
		whereParts = append(whereParts, "d.side_a_user_id = $1")
	default:
		return nil, 0, ErrInvalidChallengeBox
	}

	switch status {
	case "pending":
		whereParts = append(whereParts, "d.status = 'waiting'")
	case "accepted":
		whereParts = append(whereParts, "d.side_b_user_id IS NOT NULL")
	case "declined":
		whereParts = append(whereParts, "d.status = 'expired'", "d.side_b_user_id IS NULL")
	case "expired":
		whereParts = append(whereParts, "d.status = 'expired'", "d.side_b_user_id IS NULL")
	case "all":
	default:
		return nil, 0, ErrInvalidChallengeStatus
	}

	whereSQL := "WHERE " + strings.Join(whereParts, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM debates d %s`, whereSQL)
	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count challenges: %w", err)
	}

	args = append(args, perPage, pagination.Offset)
	query := fmt.Sprintf(`
		SELECT d.id, d.slug, d.topic, d.time_mode, d.turn_limit, d.status, d.created_at, d.challenge_expires_at,
			challenger.username, invited.username
		FROM debates d
		JOIN users challenger ON challenger.id = d.side_a_user_id
		JOIN users invited ON invited.id = d.invited_user_id
		%s
		ORDER BY d.created_at DESC
		LIMIT $2 OFFSET $3
	`, whereSQL)

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list challenges: %w", err)
	}
	defer rows.Close()

	items := make([]ChallengeListItem, 0, perPage)
	for rows.Next() {
		var (
			item      ChallengeListItem
			expiresAt sql.NullTime
		)
		if err := rows.Scan(
			&item.DebateID,
			&item.DebateSlug,
			&item.Topic,
			&item.TimeMode,
			&item.TurnLimit,
			&item.Status,
			&item.CreatedAt,
			&expiresAt,
			&item.Challenger,
			&item.InvitedUser,
		); err != nil {
			return nil, 0, fmt.Errorf("scan challenge: %w", err)
		}
		if expiresAt.Valid {
			item.ChallengeExpiresAt = &expiresAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate challenges: %w", err)
	}

	return items, total, nil
}
