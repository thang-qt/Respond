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

func (s *Store) CreateComment(ctx context.Context, params CreateCommentParams) (model.Comment, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.Comment{}, fmt.Errorf("begin comment tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var (
		status      string
		endedAt     sql.NullTime
		sideAUserID uuid.UUID
		sideBUserID uuid.UUID
		topic       string
	)

	const debateQuery = `
		SELECT status, ended_at, side_a_user_id, side_b_user_id, topic
		FROM debates
		WHERE id = $1
	`
	if err = tx.QueryRowContext(ctx, debateQuery, params.DebateID).Scan(
		&status,
		&endedAt,
		&sideAUserID,
		&sideBUserID,
		&topic,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Comment{}, ErrNotFound
		}
		return model.Comment{}, fmt.Errorf("get debate for comment: %w", err)
	}

	if status != "finished" {
		return model.Comment{}, ErrDebateNotFinished
	}

	if endedAt.Valid && time.Since(endedAt.Time) > 7*24*time.Hour {
		return model.Comment{}, ErrCommentThreadLocked
	}

	if params.IsReflection {
		if params.UserID != sideAUserID && params.UserID != sideBUserID {
			return model.Comment{}, ErrReflectionNotParticipant
		}

		var exists bool
		const reflectionQuery = `
			SELECT EXISTS(
				SELECT 1
				FROM comments
				WHERE debate_id = $1
					AND user_id = $2
					AND is_reflection = true
			)
		`
		if err = tx.QueryRowContext(ctx, reflectionQuery, params.DebateID, params.UserID).Scan(&exists); err != nil {
			return model.Comment{}, fmt.Errorf("check reflection: %w", err)
		}
		if exists {
			return model.Comment{}, ErrReflectionExists
		}
	}

	if params.ParentID != nil {
		var (
			parentParentID uuid.NullUUID
			hidden         bool
			isDeleted      bool
			parentUserID   uuid.UUID
			isReflection   bool
		)
		const parentQuery = `
			SELECT parent_id, hidden, is_deleted, user_id, is_reflection
			FROM comments
			WHERE id = $1 AND debate_id = $2
		`
		if err = tx.QueryRowContext(ctx, parentQuery, *params.ParentID, params.DebateID).Scan(
			&parentParentID,
			&hidden,
			&isDeleted,
			&parentUserID,
			&isReflection,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return model.Comment{}, ErrCommentParentNotFound
			}
			return model.Comment{}, fmt.Errorf("get parent comment: %w", err)
		}

		if hidden || isDeleted {
			return model.Comment{}, ErrCommentParentNotFound
		}

		if parentParentID.Valid {
			return model.Comment{}, ErrCommentNestedReply
		}

		if isReflection && parentUserID != params.UserID {
			if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
				UserID:      parentUserID,
				Type:        "comment_on_reflection",
				MessageKey:  "notification.comment.reply",
				MessageVars: map[string]any{"topic": topic},
				DebateID:    &params.DebateID,
			}); err != nil {
				return model.Comment{}, err
			}
		}
	}

	const insertQuery = `
		INSERT INTO comments (debate_id, parent_id, user_id, content, is_reflection)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var commentID uuid.UUID
	if err = tx.QueryRowContext(ctx, insertQuery, params.DebateID, params.ParentID, params.UserID, params.Content, params.IsReflection).Scan(&commentID); err != nil {
		return model.Comment{}, fmt.Errorf("insert comment: %w", err)
	}

	const updateCount = `
		UPDATE debates
		SET comment_count = comment_count + 1
		WHERE id = $1
	`
	if _, err = tx.ExecContext(ctx, updateCount, params.DebateID); err != nil {
		return model.Comment{}, fmt.Errorf("increment comment count: %w", err)
	}

	comment, err := fetchComment(ctx, tx, commentID, params.UserID)
	if err != nil {
		return model.Comment{}, err
	}
	comment.Replies = []model.Comment{}

	if err = tx.Commit(); err != nil {
		return model.Comment{}, fmt.Errorf("commit comment: %w", err)
	}

	return comment, nil
}

type UpdateCommentParams struct {
	DebateID  uuid.UUID
	CommentID uuid.UUID
	UserID    uuid.UUID
	Content   string
}
