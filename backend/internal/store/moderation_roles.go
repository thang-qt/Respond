package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

func (s *Store) UpdateUserRoleWithAudit(ctx context.Context, actorID, targetUserID uuid.UUID, role model.UserRole) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin role update tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx, `
		UPDATE users
		SET role = $2::user_role,
			updated_at = now()
		WHERE id = $1
	`, targetUserID, string(role))
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user role rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	if err := insertModerationActionTx(
		ctx,
		tx,
		actorID,
		"change_user_role",
		"user",
		targetUserID,
		nil,
		map[string]string{"role": string(role)},
		"",
	); err != nil {
		return fmt.Errorf("insert role change action: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit role update tx: %w", err)
	}

	return nil
}

type moderationNotifTarget struct {
	UserIDs     []uuid.UUID
	DebateID    *uuid.UUID
	DebateTopic *string
	TurnNumber  *int
}
