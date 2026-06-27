package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

func fetchComment(ctx context.Context, q queryer, commentID uuid.UUID, viewerID uuid.UUID) (model.Comment, error) {
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
				WHEN $2 IS NULL THEN false
				WHEN c.user_id = $2 THEN true
				ELSE false
			END AS is_author
		FROM comments c
		JOIN users u ON u.id = c.user_id
		JOIN debates d ON d.id = c.debate_id
		LEFT JOIN votes v
			ON v.target_type = 'comment'
			AND v.target_id = c.id
			AND v.user_id = $2
		WHERE c.id = $1
	`

	row := q.QueryRowContext(ctx, query, commentID, viewerID)
	return scanCommentRow(row, false)
}

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func scanComment(rows *sql.Rows, canViewHiddenContent bool) (model.Comment, error) {
	return scanCommentRow(rows, canViewHiddenContent)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCommentRow(row rowScanner, canViewHiddenContent bool) (model.Comment, error) {
	var (
		comment            model.Comment
		parentID           uuid.NullUUID
		updatedAt          sql.NullTime
		debaterSide        sql.NullString
		debaterAnonymousID sql.NullString
		isDeleted          bool
		hidden             bool
	)

	err := row.Scan(
		&comment.ID,
		&comment.DebateID,
		&parentID,
		&comment.Content,
		&comment.IsReflection,
		&isDeleted,
		&hidden,
		&comment.UpvoteCount,
		&comment.CreatedAt,
		&updatedAt,
		&comment.User.Username,
		&comment.User.Rating,
		&comment.IsDebater,
		&debaterSide,
		&debaterAnonymousID,
		&comment.ViewerHasUpvoted,
		&comment.IsAuthor,
	)
	if err != nil {
		return model.Comment{}, fmt.Errorf("scan comment: %w", err)
	}

	if parentID.Valid {
		comment.ParentID = &parentID.UUID
	}

	if updatedAt.Valid {
		comment.UpdatedAt = &updatedAt.Time
	}
	if debaterSide.Valid {
		side := debaterSide.String
		comment.DebaterSide = &side
	}
	if debaterAnonymousID.Valid {
		anon := debaterAnonymousID.String
		comment.DebaterAnonymous = &anon
	}

	if isDeleted {
		comment.Content = "[deleted]"
	} else if hidden && !canViewHiddenContent && !comment.IsAuthor {
		comment.Content = "This comment has been deleted."
	}
	comment.Hidden = hidden

	return comment, nil
}
