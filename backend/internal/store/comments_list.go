package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

type ListCommentsParams struct {
	DebateID             uuid.UUID
	Sort                 string
	Page                 int
	PerPage              int
	ViewerID             *uuid.UUID
	CanViewHiddenContent bool
}

type CreateCommentParams struct {
	DebateID     uuid.UUID
	UserID       uuid.UUID
	Content      string
	ParentID     *uuid.UUID
	IsReflection bool
}

var ErrInvalidSort = errors.New("invalid sort")
var ErrDebateNotFinished = errors.New("debate not finished")
var ErrCommentThreadLocked = errors.New("comment thread locked")
var ErrCommentParentNotFound = errors.New("comment parent not found")
var ErrCommentNestedReply = errors.New("comment nested reply")
var ErrReflectionExists = errors.New("reflection exists")
var ErrReflectionNotParticipant = errors.New("reflection not participant")
var ErrCommentNotAuthor = errors.New("comment not author")
var ErrCommentEditExpired = errors.New("comment edit expired")

func (s *Store) ListDebateComments(ctx context.Context, params ListCommentsParams) ([]model.Comment, int, error) {
	sort := params.Sort
	if sort == "" {
		sort = "newest"
	}
	if sort != "newest" && sort != "top" {
		return nil, 0, ErrInvalidSort
	}

	pagination := normalizePagination(params.Page, params.PerPage, 20, 50)
	perPage := pagination.PerPage

	var viewerArg any
	if params.ViewerID != nil {
		viewerArg = *params.ViewerID
	}

	const countQuery = `
		SELECT COUNT(*)
		FROM comments
		WHERE debate_id = $1
			AND parent_id IS NULL
			AND (
				$2::uuid IS NULL
				OR NOT EXISTS (
					SELECT 1
					FROM user_blocks ub
					WHERE (ub.blocker_id = $2::uuid AND ub.blocked_id = comments.user_id)
					   OR (ub.blocked_id = $2::uuid AND ub.blocker_id = comments.user_id)
				)
			)
	`
	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, params.DebateID, viewerArg).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count comments: %w", err)
	}

	orderSQL := "ORDER BY c.is_reflection DESC, c.created_at DESC"
	if sort == "top" {
		orderSQL = "ORDER BY c.is_reflection DESC, c.upvote_count DESC, c.created_at DESC"
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.debate_id, c.parent_id, c.content, c.is_reflection, c.is_deleted, c.hidden,
			c.upvote_count, c.created_at, c.updated_at,
			CASE
				WHEN c.user_id = d.side_a_user_id AND d.side_a_revealed IS NOT TRUE THEN d.side_a_anonymous_id
				WHEN c.user_id = d.side_b_user_id AND d.side_b_revealed IS NOT TRUE THEN d.side_b_anonymous_id
				ELSE u.username
			END AS username,
			u.rating,
			CASE
				WHEN c.user_id = d.side_a_user_id THEN true
				WHEN c.user_id = d.side_b_user_id THEN true
				ELSE false
			END AS is_debater,
			CASE
				WHEN c.user_id = d.side_a_user_id THEN 'a'
				WHEN c.user_id = d.side_b_user_id THEN 'b'
				ELSE NULL
			END AS debater_side,
			CASE
				WHEN c.user_id = d.side_a_user_id AND d.side_a_revealed IS NOT TRUE THEN d.side_a_anonymous_id
				WHEN c.user_id = d.side_b_user_id AND d.side_b_revealed IS NOT TRUE THEN d.side_b_anonymous_id
				ELSE NULL
			END AS debater_anonymous_id,
			CASE WHEN v.id IS NULL THEN false ELSE true END AS viewer_has_upvoted,
			CASE
				WHEN $2::uuid IS NULL THEN false
				WHEN c.user_id = $2::uuid THEN true
				ELSE false
			END AS is_author
		FROM comments c
		JOIN users u ON u.id = c.user_id
		JOIN debates d ON d.id = c.debate_id
		LEFT JOIN votes v
			ON v.target_type = 'comment'
			AND v.target_id = c.id
			AND v.user_id = $2::uuid
		WHERE c.debate_id = $1
			AND c.parent_id IS NULL
			AND (
				$2::uuid IS NULL
				OR NOT EXISTS (
					SELECT 1
					FROM user_blocks ub
					WHERE (ub.blocker_id = $2::uuid AND ub.blocked_id = c.user_id)
					   OR (ub.blocked_id = $2::uuid AND ub.blocker_id = c.user_id)
				)
			)
		%s
		LIMIT $3 OFFSET $4
	`, orderSQL)

	args := []any{params.DebateID, viewerArg, perPage, pagination.Offset}
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		comment, err := scanComment(rows, params.CanViewHiddenContent)
		if err != nil {
			return nil, 0, err
		}

		replies, err := s.listCommentReplies(ctx, comment.ID, viewerArg, params.CanViewHiddenContent)
		if err != nil {
			return nil, 0, err
		}
		if replies == nil {
			replies = []model.Comment{}
		}
		comment.Replies = replies
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate comments: %w", err)
	}

	return comments, total, nil
}
