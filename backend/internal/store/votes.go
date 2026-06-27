package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrCommentNotFound = errors.New("comment not found")
var ErrDebateNotFound = errors.New("debate not found")

type ToggleCommentVoteResult struct {
	CommentID   uuid.UUID
	Voted       bool
	UpvoteCount int
}

type ToggleDebateVoteResult struct {
	DebateID    uuid.UUID
	Voted       bool
	UpvoteCount int
}

func (s *Store) ToggleDebateVote(ctx context.Context, debateID, userID uuid.UUID) (ToggleDebateVoteResult, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return ToggleDebateVoteResult{}, fmt.Errorf("begin vote tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var (
		exists  bool
		blocked bool
	)
	if err := tx.QueryRowContext(ctx, `
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
		return ToggleDebateVoteResult{}, fmt.Errorf("check debate vote visibility: %w", err)
	}
	if !exists {
		return ToggleDebateVoteResult{}, ErrDebateNotFound
	}
	if blocked {
		return ToggleDebateVoteResult{}, ErrDebateHiddenByBlock
	}

	var insertedID uuid.UUID
	insertErr := tx.QueryRowContext(ctx, `
		INSERT INTO votes (user_id, target_type, target_id)
		VALUES ($1, 'debate', $2)
		ON CONFLICT DO NOTHING
		RETURNING id
	`, userID, debateID).Scan(&insertedID)

	var voted bool
	var upvoteCount int
	switch {
	case insertErr == nil:
		voted = true
		if err := tx.QueryRowContext(ctx, `
			UPDATE debates
			SET upvote_count = upvote_count + 1
			WHERE id = $1
			RETURNING upvote_count
		`, debateID).Scan(&upvoteCount); err != nil {
			return ToggleDebateVoteResult{}, fmt.Errorf("increment debate upvotes: %w", err)
		}
	case errors.Is(insertErr, sql.ErrNoRows):
		voted = false
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM votes
			WHERE user_id = $1 AND target_type = 'debate' AND target_id = $2
		`, userID, debateID); err != nil {
			return ToggleDebateVoteResult{}, fmt.Errorf("delete debate vote: %w", err)
		}
		if err := tx.QueryRowContext(ctx, `
			UPDATE debates
			SET upvote_count = GREATEST(upvote_count - 1, 0)
			WHERE id = $1
			RETURNING upvote_count
		`, debateID).Scan(&upvoteCount); err != nil {
			return ToggleDebateVoteResult{}, fmt.Errorf("decrement debate upvotes: %w", err)
		}
	default:
		return ToggleDebateVoteResult{}, fmt.Errorf("insert debate vote: %w", insertErr)
	}

	if err := tx.Commit(); err != nil {
		return ToggleDebateVoteResult{}, fmt.Errorf("commit debate vote: %w", err)
	}

	return ToggleDebateVoteResult{
		DebateID:    debateID,
		Voted:       voted,
		UpvoteCount: upvoteCount,
	}, nil
}

func (s *Store) ToggleCommentVote(ctx context.Context, commentID, userID uuid.UUID) (ToggleCommentVoteResult, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return ToggleCommentVoteResult{}, fmt.Errorf("begin vote tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var (
		exists  bool
		blocked bool
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT
			EXISTS (
				SELECT 1
				FROM comments c
				JOIN debates d ON d.id = c.debate_id
				WHERE c.id = $1
					AND c.is_deleted = false
					AND c.hidden = false
					AND d.hidden = false
			),
			EXISTS (
				SELECT 1
				FROM comments c
				JOIN debates d ON d.id = c.debate_id
				WHERE c.id = $1
					AND c.is_deleted = false
					AND c.hidden = false
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
	`, commentID, userID).Scan(&exists, &blocked); err != nil {
		return ToggleCommentVoteResult{}, fmt.Errorf("check comment vote visibility: %w", err)
	}
	if !exists {
		return ToggleCommentVoteResult{}, ErrCommentNotFound
	}
	if blocked {
		return ToggleCommentVoteResult{}, ErrDebateHiddenByBlock
	}

	var insertedID uuid.UUID
	insertErr := tx.QueryRowContext(ctx, `
		INSERT INTO votes (user_id, target_type, target_id)
		VALUES ($1, 'comment', $2)
		ON CONFLICT DO NOTHING
		RETURNING id
	`, userID, commentID).Scan(&insertedID)

	var voted bool
	var upvoteCount int
	switch {
	case insertErr == nil:
		voted = true
		if err := tx.QueryRowContext(ctx, `
			UPDATE comments
			SET upvote_count = upvote_count + 1
			WHERE id = $1
			RETURNING upvote_count
		`, commentID).Scan(&upvoteCount); err != nil {
			return ToggleCommentVoteResult{}, fmt.Errorf("increment comment upvotes: %w", err)
		}
	case errors.Is(insertErr, sql.ErrNoRows):
		voted = false
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM votes
			WHERE user_id = $1 AND target_type = 'comment' AND target_id = $2
		`, userID, commentID); err != nil {
			return ToggleCommentVoteResult{}, fmt.Errorf("delete vote: %w", err)
		}
		if err := tx.QueryRowContext(ctx, `
			UPDATE comments
			SET upvote_count = GREATEST(upvote_count - 1, 0)
			WHERE id = $1
			RETURNING upvote_count
		`, commentID).Scan(&upvoteCount); err != nil {
			return ToggleCommentVoteResult{}, fmt.Errorf("decrement comment upvotes: %w", err)
		}
	default:
		return ToggleCommentVoteResult{}, fmt.Errorf("insert vote: %w", insertErr)
	}

	if err := tx.Commit(); err != nil {
		return ToggleCommentVoteResult{}, fmt.Errorf("commit vote: %w", err)
	}

	return ToggleCommentVoteResult{
		CommentID:   commentID,
		Voted:       voted,
		UpvoteCount: upvoteCount,
	}, nil
}
