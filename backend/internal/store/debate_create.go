package store

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"respond/internal/model"
)

type CreateDebateParams struct {
	Topic                    string
	TagIDs                   []uuid.UUID
	TimeMode                 string
	TurnLimit                int
	Context                  *string
	OpeningTurn              string
	OpeningTurnAIAssisted    bool
	OpeningTurnAINote        *string
	UserID                   uuid.UUID
	PromptID                 *uuid.UUID
	InvitedUserID            *uuid.UUID
	ChallengeExpiresAt       *time.Time
	ChallengeIdentityVisible bool
}

func (s *Store) CountActiveDebatesForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM debates
		WHERE status IN ('waiting', 'active', 'pending_extension', 'waiting_replacement')
			AND (side_a_user_id = $1 OR side_b_user_id = $1)
	`
	var count int
	if err := s.DB.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active debates: %w", err)
	}
	return count, nil
}

func (s *Store) CreateDebate(ctx context.Context, params CreateDebateParams) (model.DebateDetail, error) {
	anonID, err := generateAnonymousID("A")
	if err != nil {
		return model.DebateDetail{}, fmt.Errorf("generate anonymous id: %w", err)
	}

	var promptArg any
	if params.PromptID != nil {
		promptArg = *params.PromptID
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.DebateDetail{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const insertDebate = `
		INSERT INTO debates (
			slug, topic, time_mode, turn_limit, context,
			status, side_a_user_id, side_a_anonymous_id,
			current_turn_side, turn_count, prompt_id,
			invited_user_id, challenge_expires_at, challenge_identity_visible
		)
		VALUES ($1, $2, $3, $4, $5, 'waiting', $6, $7, 'b', 1, $8, $9, $10, $11)
		RETURNING id, created_at
	`

	var id uuid.UUID
	var createdAt time.Time
	baseSlug := slugify(params.Topic)
	for attempt := 0; attempt < 3; attempt++ {
		slug, err := ensureUniqueDebateSlug(ctx, tx, baseSlug)
		if err != nil {
			return model.DebateDetail{}, fmt.Errorf("generate debate slug: %w", err)
		}
		if err := tx.QueryRowContext(
			ctx,
			insertDebate,
			slug,
			params.Topic,
			params.TimeMode,
			params.TurnLimit,
			params.Context,
			params.UserID,
			anonID,
			promptArg,
			params.InvitedUserID,
			params.ChallengeExpiresAt,
			params.ChallengeIdentityVisible,
		).Scan(&id, &createdAt); err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				continue
			}
			return model.DebateDetail{}, fmt.Errorf("insert debate: %w", err)
		}
		break
	}
	if id == uuid.Nil {
		return model.DebateDetail{}, fmt.Errorf("insert debate: slug collision")
	}

	const insertTurn = `
		INSERT INTO turns (debate_id, turn_number, side, user_id, anonymous_id, content, ai_assisted, ai_note)
		VALUES ($1, 1, 'a', $2, $3, $4, $5, $6)
	`

	if _, err := tx.ExecContext(ctx, insertTurn, id, params.UserID, anonID, params.OpeningTurn, params.OpeningTurnAIAssisted, params.OpeningTurnAINote); err != nil {
		return model.DebateDetail{}, fmt.Errorf("insert opening turn: %w", err)
	}

	const insertDebateTag = `
		INSERT INTO debate_tags (debate_id, tag_id)
		VALUES ($1, $2)
	`
	for _, tagID := range params.TagIDs {
		if _, err := tx.ExecContext(ctx, insertDebateTag, id, tagID); err != nil {
			return model.DebateDetail{}, fmt.Errorf("insert debate tag: %w", err)
		}
	}

	if _, err := createSeatStintTx(ctx, tx, id, "a", params.UserID, anonID, createdAt); err != nil {
		return model.DebateDetail{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.DebateDetail{}, fmt.Errorf("commit tx: %w", err)
	}

	return s.GetDebateByID(ctx, id, false)
}

func generateAnonymousID(prefix string) (string, error) {
	max := big.NewInt(10000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s#%04d", prefix, n.Int64()), nil
}
