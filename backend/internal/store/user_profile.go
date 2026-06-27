package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"respond/internal/model"
)

func (s *Store) GetUserProfileByUsername(ctx context.Context, username string) (UserProfile, uuid.UUID, error) {
	const query = `
		SELECT u.id, u.username, u.bio, u.rating, u.wins, u.losses, u.draws, u.created_at,
			(
				SELECT COUNT(*)
				FROM debates d
				WHERE d.status = 'finished'
					AND (d.side_a_user_id = u.id OR d.side_b_user_id = u.id)
			) AS debates_count
		FROM users u
		WHERE LOWER(u.username) = LOWER($1)
	`

	var (
		profile UserProfile
		userID  uuid.UUID
	)

	err := s.DB.QueryRowContext(ctx, query, username).Scan(
		&userID,
		&profile.Username,
		&profile.Bio,
		&profile.Rating,
		&profile.Wins,
		&profile.Losses,
		&profile.Draws,
		&profile.CreatedAt,
		&profile.DebatesCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserProfile{}, uuid.Nil, ErrNotFound
		}
		return UserProfile{}, uuid.Nil, fmt.Errorf("get user profile by username: %w", err)
	}

	return profile, userID, nil
}

type ListUserDebatesParams struct {
	Username string
	ViewerID *uuid.UUID
	Page     int
	PerPage  int
	Locale   string
}

func (s *Store) ListUserDebatesByUsername(ctx context.Context, params ListUserDebatesParams) ([]model.DebateFeedItem, int, error) {
	pagination := normalizePagination(params.Page, params.PerPage, 20, 50)
	perPage := pagination.PerPage

	profile, targetUserID, err := s.GetUserProfileByUsername(ctx, params.Username)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, 0, ErrNotFound
		}
		return nil, 0, fmt.Errorf("list user debates get profile: %w", err)
	}
	_ = profile

	isOwner := params.ViewerID != nil && *params.ViewerID == targetUserID

	if params.ViewerID != nil && !isOwner {
		blocked, err := s.IsEitherUserBlocked(ctx, *params.ViewerID, targetUserID)
		if err != nil {
			return nil, 0, fmt.Errorf("list user debates blocked check: %w", err)
		}
		if blocked {
			return nil, 0, ErrUserHiddenByBlock
		}
	}

	args := make([]any, 0, 6)
	args = append(args, targetUserID)
	whereParts := []string{"(d.side_a_user_id = $1 OR d.side_b_user_id = $1)"}
	whereParts = append(whereParts, "NOT (d.status = 'waiting' AND d.invited_user_id IS NOT NULL)")
	whereParts = append(whereParts, "d.status <> 'expired'")
	if !isOwner {
		whereParts = append(whereParts, "d.hidden = false")
	}
	if !isOwner {
		whereParts = append(whereParts, "d.status = 'finished'")
		whereParts = append(whereParts, "((d.side_a_user_id = $1 AND d.side_a_revealed = true) OR (d.side_b_user_id = $1 AND d.side_b_revealed = true))")
	}
	if params.ViewerID != nil {
		args = append(args, *params.ViewerID)
		viewerFilterPos := len(args)
		whereParts = append(whereParts, fmt.Sprintf(`(
			d.side_a_user_id = $%d
			OR d.side_b_user_id = $%d
			OR NOT EXISTS (
				SELECT 1
				FROM user_blocks ub
				WHERE (ub.blocker_id = $%d AND (ub.blocked_id = d.side_a_user_id OR ub.blocked_id = d.side_b_user_id))
				   OR (ub.blocked_id = $%d AND (ub.blocker_id = d.side_a_user_id OR ub.blocker_id = d.side_b_user_id))
			)
		)`, viewerFilterPos, viewerFilterPos, viewerFilterPos, viewerFilterPos))
	}
	whereSQL := "WHERE " + strings.Join(whereParts, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM debates d %s`, whereSQL)
	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count user debates: %w", err)
	}

	var viewerArg any
	if params.ViewerID != nil {
		viewerArg = *params.ViewerID
	}

	args = append(args, viewerArg, perPage, pagination.Offset)
	query := fmt.Sprintf(`
		SELECT d.id, d.slug, d.topic, d.time_mode, d.turn_limit, d.context, d.status,
			d.side_a_anonymous_id, d.side_b_anonymous_id, d.side_a_revealed, d.side_b_revealed,
			d.outcome, d.winner_side, d.current_turn_side, d.turn_count, d.turn_deadline,
			d.upvote_count, d.comment_count, d.prompt_id, d.created_at, d.started_at, d.ended_at,
			NULL::integer AS side_a_rating_delta, NULL::integer AS side_b_rating_delta,
			d.moderation_pending, d.hidden,
			lt.turn_number, lt.side, lt.anonymous_id, lt.content, lt.created_at,
			ua.username, ua.rating,
			ub.username, ub.rating,
			CASE WHEN v.id IS NULL THEN false ELSE true END AS viewer_has_upvoted,
			EXISTS (
				SELECT 1 FROM follows f
				WHERE f.user_id = $%d AND f.debate_id = d.id
			) AS is_following,
			CASE
				WHEN $%d IS NULL THEN false
				WHEN d.side_a_user_id = $%d OR d.side_b_user_id = $%d THEN true
				ELSE false
			END AS viewer_is_participant
		FROM debates d
		JOIN users ua ON ua.id = d.side_a_user_id
		LEFT JOIN users ub ON ub.id = d.side_b_user_id
		LEFT JOIN votes v
			ON v.target_type = 'debate'
			AND v.target_id = d.id
			AND v.user_id = $%d
		LEFT JOIN LATERAL (
			SELECT
				t.turn_number,
				t.side,
				t.anonymous_id,
				CASE
					WHEN d.hidden THEN 'This content has been flagged for review.'
					WHEN t.hidden THEN 'This content has been flagged for review.'
					WHEN char_length(t.content) > 360 THEN left(convert_from(convert_to(t.content, 'UTF8'), 'UTF8'), 360) || '...'
					ELSE t.content
				END AS content,
				t.created_at
			FROM turns t
			WHERE t.debate_id = d.id AND t.is_system = false
			ORDER BY t.turn_number DESC
			LIMIT 1
		) lt ON true
		%s
		ORDER BY d.created_at DESC
		LIMIT $%d OFFSET $%d
	`,
		len(args)-2,
		len(args)-2,
		len(args)-2,
		len(args)-2,
		len(args)-2,
		whereSQL,
		len(args)-1,
		len(args),
	)

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list user debates: %w", err)
	}
	defer rows.Close()

	debates, err := scanDebateFeedRows(ctx, s, rows, params.Locale)
	if err != nil {
		return nil, 0, err
	}

	return debates, total, nil
}

type CreateUserParams struct {
	Email           string
	Username        string
	PasswordHash    string
	InvitedByUserID *uuid.UUID
}
