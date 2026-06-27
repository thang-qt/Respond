package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func (s *Store) FollowDebate(ctx context.Context, debateID, userID uuid.UUID) error {
	var (
		exists  bool
		blocked bool
	)
	if err := s.DB.QueryRowContext(ctx, `
		SELECT
			EXISTS (SELECT 1 FROM debates WHERE id = $1 AND hidden = false),
			EXISTS (
				SELECT 1
				FROM debates d
				WHERE d.id = $1
					AND d.hidden = false
					AND d.side_a_user_id <> $2
					AND (d.side_b_user_id IS NULL OR d.side_b_user_id <> $2)
					AND EXISTS (
						SELECT 1
						FROM user_blocks ub
						WHERE (ub.blocker_id = $2 AND (ub.blocked_id = d.side_a_user_id OR ub.blocked_id = d.side_b_user_id))
						   OR (ub.blocked_id = $2 AND (ub.blocker_id = d.side_a_user_id OR ub.blocker_id = d.side_b_user_id))
					)
			)
	`, debateID, userID).Scan(&exists, &blocked); err != nil {
		return fmt.Errorf("check debate follow visibility: %w", err)
	}
	if !exists {
		return ErrDebateNotFound
	}
	if blocked {
		return ErrDebateHiddenByBlock
	}

	if _, err := s.DB.ExecContext(ctx, `
		INSERT INTO follows (user_id, debate_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, debateID); err != nil {
		return fmt.Errorf("insert follow: %w", err)
	}

	return nil
}

func (s *Store) UnfollowDebate(ctx context.Context, debateID, userID uuid.UUID) error {
	var (
		exists  bool
		blocked bool
	)
	if err := s.DB.QueryRowContext(ctx, `
		SELECT
			EXISTS (SELECT 1 FROM debates WHERE id = $1 AND hidden = false),
			EXISTS (
				SELECT 1
				FROM debates d
				WHERE d.id = $1
					AND d.hidden = false
					AND d.side_a_user_id <> $2
					AND (d.side_b_user_id IS NULL OR d.side_b_user_id <> $2)
					AND EXISTS (
						SELECT 1
						FROM user_blocks ub
						WHERE (ub.blocker_id = $2 AND (ub.blocked_id = d.side_a_user_id OR ub.blocked_id = d.side_b_user_id))
						   OR (ub.blocked_id = $2 AND (ub.blocker_id = d.side_a_user_id OR ub.blocker_id = d.side_b_user_id))
					)
			)
	`, debateID, userID).Scan(&exists, &blocked); err != nil {
		return fmt.Errorf("check debate unfollow visibility: %w", err)
	}
	if !exists {
		return ErrDebateNotFound
	}
	if blocked {
		return ErrDebateHiddenByBlock
	}

	if _, err := s.DB.ExecContext(ctx, `
		DELETE FROM follows
		WHERE user_id = $1 AND debate_id = $2
	`, userID, debateID); err != nil {
		return fmt.Errorf("delete follow: %w", err)
	}

	return nil
}

func listFollowerIDsTx(ctx context.Context, tx *sql.Tx, debateID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT user_id
		FROM follows
		WHERE debate_id = $1
	`, debateID)
	if err != nil {
		return nil, fmt.Errorf("list followers: %w", err)
	}
	defer rows.Close()

	var followerIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan follower: %w", err)
		}
		followerIDs = append(followerIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate followers: %w", err)
	}

	if followerIDs == nil {
		followerIDs = []uuid.UUID{}
	}

	return followerIDs, nil
}
