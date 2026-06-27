package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

func (s *Store) listCommentReplies(ctx context.Context, parentID uuid.UUID, viewerArg any, canViewHiddenContent bool) ([]model.Comment, error) {
	const query = `
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
		WHERE c.parent_id = $1
			AND (
				$2::uuid IS NULL
				OR NOT EXISTS (
					SELECT 1
					FROM user_blocks ub
					WHERE (ub.blocker_id = $2::uuid AND ub.blocked_id = c.user_id)
					   OR (ub.blocked_id = $2::uuid AND ub.blocker_id = c.user_id)
				)
			)
		ORDER BY c.created_at ASC
	`

	rows, err := s.DB.QueryContext(ctx, query, parentID, viewerArg)
	if err != nil {
		return nil, fmt.Errorf("list comment replies: %w", err)
	}
	defer rows.Close()

	var replies []model.Comment
	for rows.Next() {
		reply, err := scanComment(rows, canViewHiddenContent)
		if err != nil {
			return nil, err
		}
		replies = append(replies, reply)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate replies: %w", err)
	}

	if replies == nil {
		replies = []model.Comment{}
	}

	return replies, nil
}
