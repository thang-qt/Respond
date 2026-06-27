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
	ErrUserBlockSelf       = errors.New("cannot block yourself")
	ErrUserAlreadyBlocked  = errors.New("user already blocked")
	ErrUserNotBlocked      = errors.New("user not blocked")
	ErrDebateHiddenByBlock = errors.New("debate hidden by block")
	ErrUserHiddenByBlock   = errors.New("user hidden by block")
)

type BlockedUser struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Store) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	if blockerID == blockedID {
		return ErrUserBlockSelf
	}

	res, err := s.DB.ExecContext(ctx, `
		INSERT INTO user_blocks (blocker_id, blocked_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, blockerID, blockedID)
	if err != nil {
		return fmt.Errorf("block user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("block user rows affected: %w", err)
	}
	if rows == 0 {
		return ErrUserAlreadyBlocked
	}

	return nil
}

func (s *Store) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	res, err := s.DB.ExecContext(ctx, `
		DELETE FROM user_blocks
		WHERE blocker_id = $1 AND blocked_id = $2
	`, blockerID, blockedID)
	if err != nil {
		return fmt.Errorf("unblock user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("unblock user rows affected: %w", err)
	}
	if rows == 0 {
		return ErrUserNotBlocked
	}

	return nil
}

func (s *Store) IsEitherUserBlocked(ctx context.Context, userAID, userBID uuid.UUID) (bool, error) {
	return isEitherUserBlocked(ctx, s.DB, userAID, userBID)
}

func isEitherUserBlockedTx(ctx context.Context, tx *sql.Tx, userAID, userBID uuid.UUID) (bool, error) {
	return isEitherUserBlocked(ctx, tx, userAID, userBID)
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func isEitherUserBlocked(ctx context.Context, q queryRower, userAID, userBID uuid.UUID) (bool, error) {
	if userAID == userBID {
		return false, nil
	}

	var blocked bool
	err := q.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_blocks
			WHERE (blocker_id = $1 AND blocked_id = $2)
			   OR (blocker_id = $2 AND blocked_id = $1)
		)
	`, userAID, userBID).Scan(&blocked)
	if err != nil {
		return false, fmt.Errorf("check users blocked: %w", err)
	}

	return blocked, nil
}

func (s *Store) ListBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]BlockedUser, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT ub.blocked_id, u.username, ub.created_at
		FROM user_blocks ub
		JOIN users u ON u.id = ub.blocked_id
		WHERE ub.blocker_id = $1
		ORDER BY ub.created_at DESC, ub.blocked_id ASC
	`, blockerID)
	if err != nil {
		return nil, fmt.Errorf("list blocked users: %w", err)
	}
	defer rows.Close()

	result := make([]BlockedUser, 0)
	for rows.Next() {
		var item BlockedUser
		if err := rows.Scan(&item.UserID, &item.Username, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan blocked user: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate blocked users: %w", err)
	}

	return result, nil
}

func (s *Store) IsBlockedBy(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error) {
	var exists bool
	err := s.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_blocks
			WHERE blocker_id = $1 AND blocked_id = $2
		)
	`, blockerID, blockedID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("is blocked by: %w", err)
	}
	return exists, nil
}

func (s *Store) IsDebateBlockedForViewer(ctx context.Context, debateID, viewerID uuid.UUID) (bool, error) {
	var blocked bool
	err := s.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM debates d
			WHERE d.id = $1
				AND d.side_a_user_id <> $2
				AND (d.side_b_user_id IS NULL OR d.side_b_user_id <> $2)
				AND EXISTS (
					SELECT 1
					FROM user_blocks ub
					WHERE (ub.blocker_id = $2 AND (ub.blocked_id = d.side_a_user_id OR ub.blocked_id = d.side_b_user_id))
					   OR (ub.blocked_id = $2 AND (ub.blocker_id = d.side_a_user_id OR ub.blocker_id = d.side_b_user_id))
				)
		)
	`, debateID, viewerID).Scan(&blocked)
	if err != nil {
		return false, fmt.Errorf("is debate blocked for viewer: %w", err)
	}
	return blocked, nil
}
