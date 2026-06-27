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

func (s *Store) UpdateComment(ctx context.Context, params UpdateCommentParams) (model.Comment, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.Comment{}, fmt.Errorf("begin comment edit tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var (
		authorID  uuid.UUID
		createdAt time.Time
		isDeleted bool
	)
	err = tx.QueryRowContext(ctx, `
		SELECT user_id, created_at, is_deleted
		FROM comments
		WHERE id = $1 AND debate_id = $2
	`, params.CommentID, params.DebateID).Scan(&authorID, &createdAt, &isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Comment{}, ErrCommentNotFound
		}
		return model.Comment{}, fmt.Errorf("get comment for edit: %w", err)
	}
	if isDeleted {
		return model.Comment{}, ErrCommentNotFound
	}
	if authorID != params.UserID {
		return model.Comment{}, ErrCommentNotAuthor
	}
	if time.Since(createdAt) > 5*time.Minute {
		return model.Comment{}, ErrCommentEditExpired
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE comments
		SET content = $1, updated_at = $2
		WHERE id = $3
	`, params.Content, time.Now().UTC(), params.CommentID); err != nil {
		return model.Comment{}, fmt.Errorf("update comment: %w", err)
	}

	comment, err := fetchComment(ctx, tx, params.CommentID, params.UserID)
	if err != nil {
		return model.Comment{}, err
	}
	comment.Replies = []model.Comment{}

	if err := tx.Commit(); err != nil {
		return model.Comment{}, fmt.Errorf("commit comment edit: %w", err)
	}

	return comment, nil
}

type DeleteCommentParams struct {
	DebateID  uuid.UUID
	CommentID uuid.UUID
	UserID    uuid.UUID
}

func (s *Store) DeleteComment(ctx context.Context, params DeleteCommentParams) error {
	var (
		authorID  uuid.UUID
		isDeleted bool
	)
	err := s.DB.QueryRowContext(ctx, `
		SELECT user_id, is_deleted
		FROM comments
		WHERE id = $1 AND debate_id = $2
	`, params.CommentID, params.DebateID).Scan(&authorID, &isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCommentNotFound
		}
		return fmt.Errorf("get comment for delete: %w", err)
	}
	if isDeleted {
		return ErrCommentNotFound
	}
	if authorID != params.UserID {
		return ErrCommentNotAuthor
	}

	if _, err := s.DB.ExecContext(ctx, `
		UPDATE comments
		SET is_deleted = true, content = '', updated_at = $1
		WHERE id = $2
	`, time.Now().UTC(), params.CommentID); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	return nil
}
