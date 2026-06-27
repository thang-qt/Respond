package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/i18n"
	"respond/internal/model"
)

func (s *Store) GetDebateByID(ctx context.Context, id uuid.UUID, canViewHiddenContent bool) (model.DebateDetail, error) {
	return s.GetDebateByIDLocalized(ctx, id, canViewHiddenContent, i18n.DefaultLocale)
}

func (s *Store) GetDebateByIDLocalized(ctx context.Context, id uuid.UUID, canViewHiddenContent bool, locale string) (model.DebateDetail, error) {
	const query = `
		SELECT d.id, d.slug, d.topic, d.time_mode, d.turn_limit, d.context, d.status,
			d.side_a_anonymous_id, d.side_b_anonymous_id, d.side_a_revealed, d.side_b_revealed,
			d.outcome, d.winner_side, d.current_turn_side, d.turn_count, d.turn_deadline,
			d.draw_proposed_by, d.open_side,
			d.extension_deadline, d.extension_a_accepted, d.extension_b_accepted,
			d.upvote_count, d.comment_count, d.prompt_id, d.created_at, d.started_at, d.ended_at,
			d.side_a_rating_delta, d.side_b_rating_delta, d.moderation_pending, d.hidden,
			(d.invited_user_id IS NOT NULL) AS is_challenge,
			d.challenge_identity_visible,
			ui.username,
			ua.username, ua.rating,
			ub.username, ub.rating
		FROM debates d
		LEFT JOIN users ui ON ui.id = d.invited_user_id
		JOIN users ua ON ua.id = d.side_a_user_id
		LEFT JOIN users ub ON ub.id = d.side_b_user_id
		WHERE d.id = $1
	`

	var (
		detail                   model.DebateDetail
		context                  sql.NullString
		sideAAnon                string
		sideBAnon                sql.NullString
		sideARevealed            sql.NullBool
		sideBRevealed            sql.NullBool
		outcome                  sql.NullString
		winnerSide               sql.NullString
		turnDeadline             sql.NullTime
		drawProposedBy           sql.NullString
		openSide                 sql.NullString
		extDeadline              sql.NullTime
		extAAccepted             sql.NullBool
		extBAccepted             sql.NullBool
		promptID                 sql.NullString
		startedAt                sql.NullTime
		endedAt                  sql.NullTime
		sideARatingDelta         sql.NullInt64
		sideBRatingDelta         sql.NullInt64
		sideAUsername            string
		sideARating              int
		isChallenge              bool
		challengeIdentityVisible bool
		invitedUsername          sql.NullString
		sideBUsername            sql.NullString
		sideBRating              sql.NullInt64
	)

	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&detail.ID,
		&detail.Slug,
		&detail.Topic,
		&detail.TimeMode,
		&detail.TurnLimit,
		&context,
		&detail.Status,
		&sideAAnon,
		&sideBAnon,
		&sideARevealed,
		&sideBRevealed,
		&outcome,
		&winnerSide,
		&detail.CurrentTurnSide,
		&detail.TurnCount,
		&turnDeadline,
		&drawProposedBy,
		&openSide,
		&extDeadline,
		&extAAccepted,
		&extBAccepted,
		&detail.UpvoteCount,
		&detail.CommentCount,
		&promptID,
		&detail.CreatedAt,
		&startedAt,
		&endedAt,
		&sideARatingDelta,
		&sideBRatingDelta,
		&detail.ModerationPending,
		&detail.Hidden,
		&isChallenge,
		&challengeIdentityVisible,
		&invitedUsername,
		&sideAUsername,
		&sideARating,
		&sideBUsername,
		&sideBRating,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.DebateDetail{}, ErrNotFound
		}
		return model.DebateDetail{}, fmt.Errorf("get debate: %w", err)
	}

	detail.SideA.AnonymousID = &sideAAnon
	detail.IsChallenge = isChallenge
	detail.ChallengeIdentityVisible = challengeIdentityVisible
	if invitedUsername.Valid && challengeIdentityVisible {
		detail.InvitedUsername = &invitedUsername.String
	}
	if context.Valid {
		detail.Context = &context.String
	}
	if sideBAnon.Valid {
		anon := sideBAnon.String
		detail.SideB.AnonymousID = &anon
	}
	detail.SideA.Revealed = sideARevealed.Valid && sideARevealed.Bool
	detail.SideB.Revealed = sideBRevealed.Valid && sideBRevealed.Bool

	if outcome.Valid {
		detail.Outcome = &outcome.String
	}
	if winnerSide.Valid {
		detail.WinnerSide = &winnerSide.String
	}
	if turnDeadline.Valid {
		detail.TurnDeadline = &turnDeadline.Time
	}
	if drawProposedBy.Valid {
		detail.DrawProposedBy = &drawProposedBy.String
	}
	if openSide.Valid {
		detail.OpenSide = &openSide.String
	}
	if extDeadline.Valid {
		detail.ExtensionDeadline = &extDeadline.Time
	}
	if extAAccepted.Valid {
		detail.ExtensionAAccepted = &extAAccepted.Bool
	}
	if extBAccepted.Valid {
		detail.ExtensionBAccepted = &extBAccepted.Bool
	}
	if startedAt.Valid {
		detail.StartedAt = &startedAt.Time
	}
	if endedAt.Valid {
		detail.EndedAt = &endedAt.Time
	}
	if sideARatingDelta.Valid {
		delta := int(sideARatingDelta.Int64)
		detail.SideARatingDelta = &delta
	}
	if sideBRatingDelta.Valid {
		delta := int(sideBRatingDelta.Int64)
		detail.SideBRatingDelta = &delta
	}

	detail.IsDailyPrompt = promptID.Valid
	detail.SpectatorCount = 0

	detail.SideA.User = selectSideUser(detail.Status, detail.SideA.Revealed, &model.UserSummary{
		Username: sideAUsername,
		Rating:   sideARating,
	})
	if detail.Status == "waiting" && detail.IsChallenge && detail.ChallengeIdentityVisible {
		detail.SideA.User = &model.UserSummary{
			Username: sideAUsername,
			Rating:   sideARating,
		}
	}
	if sideBUsername.Valid && sideBRating.Valid {
		detail.SideB.User = &model.UserSummary{
			Username: sideBUsername.String,
			Rating:   int(sideBRating.Int64),
		}
	}
	detail.SideB.User = selectSideUser(detail.Status, detail.SideB.Revealed, detail.SideB.User)

	turns, err := s.listDebateTurns(ctx, id, canViewHiddenContent)
	if err != nil {
		return model.DebateDetail{}, err
	}
	if detail.Hidden && !canViewHiddenContent {
		for i := range turns {
			if turns[i].IsSystem {
				continue
			}
			turns[i].Hidden = true
			turns[i].Content = "This content has been flagged for review."
		}
	}
	detail.Turns = turns

	events, err := s.listDebateEvents(ctx, id)
	if err != nil {
		return model.DebateDetail{}, err
	}
	detail.Timeline = buildDebateTimeline(turns, events)

	stints, err := s.listDebateSeatStints(ctx, id)
	if err != nil {
		return model.DebateDetail{}, err
	}
	detail.ParticipantHistory = stints

	tags, err := s.ListTagsByDebateIDLocalized(ctx, id, locale)
	if err != nil {
		return model.DebateDetail{}, err
	}
	detail.Tags = tags

	return detail, nil
}

func (s *Store) GetDebateIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	const query = `SELECT id FROM debates WHERE slug = $1`
	var id uuid.UUID
	if err := s.DB.QueryRowContext(ctx, query, slug).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, fmt.Errorf("get debate id by slug: %w", err)
	}
	return id, nil
}

func (s *Store) GetDebateBySlug(ctx context.Context, slug string, canViewHiddenContent bool) (model.DebateDetail, error) {
	id, err := s.GetDebateIDBySlug(ctx, slug)
	if err != nil {
		return model.DebateDetail{}, err
	}
	return s.GetDebateByID(ctx, id, canViewHiddenContent)
}

func (s *Store) GetDebateBySlugLocalized(ctx context.Context, slug string, canViewHiddenContent bool, locale string) (model.DebateDetail, error) {
	id, err := s.GetDebateIDBySlug(ctx, slug)
	if err != nil {
		return model.DebateDetail{}, err
	}
	return s.GetDebateByIDLocalized(ctx, id, canViewHiddenContent, locale)
}

func (s *Store) GetDebateStatus(ctx context.Context, id uuid.UUID) (string, bool, error) {
	const query = `SELECT status, hidden FROM debates WHERE id = $1`
	var status string
	var hidden bool
	if err := s.DB.QueryRowContext(ctx, query, id).Scan(&status, &hidden); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, ErrNotFound
		}
		return "", false, fmt.Errorf("get debate status: %w", err)
	}
	return status, hidden, nil
}

func (s *Store) GetDebateViewer(ctx context.Context, debateID, userID uuid.UUID) (model.DebateViewer, error) {
	const query = `
		SELECT
			CASE
				WHEN d.side_a_user_id = $2 THEN true
				WHEN d.side_b_user_id = $2 THEN true
				ELSE false
			END AS is_participant,
			CASE
				WHEN d.side_a_user_id = $2 THEN 'a'
				WHEN d.side_b_user_id = $2 THEN 'b'
				ELSE NULL
			END AS side,
			d.side_a_revealed,
			d.side_b_revealed,
			EXISTS (
				SELECT 1 FROM votes v
				WHERE v.user_id = $2 AND v.target_type = 'debate' AND v.target_id = $1
			) AS has_upvoted,
			EXISTS (
				SELECT 1 FROM follows f
				WHERE f.user_id = $2 AND f.debate_id = $1
			) AS is_following
		FROM debates d
		WHERE d.id = $1
	`

	var (
		viewer      model.DebateViewer
		side        sql.NullString
		sideAReveal sql.NullBool
		sideBReveal sql.NullBool
	)

	err := s.DB.QueryRowContext(ctx, query, debateID, userID).Scan(
		&viewer.IsParticipant,
		&side,
		&sideAReveal,
		&sideBReveal,
		&viewer.HasUpvoted,
		&viewer.IsFollowing,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.DebateViewer{}, ErrNotFound
		}
		return model.DebateViewer{}, fmt.Errorf("get debate viewer: %w", err)
	}

	if side.Valid {
		viewer.Side = &side.String
		if side.String == "a" {
			if sideAReveal.Valid {
				value := sideAReveal.Bool
				viewer.RevealChoice = &value
			}
		} else if side.String == "b" {
			if sideBReveal.Valid {
				value := sideBReveal.Bool
				viewer.RevealChoice = &value
			}
		}
	}

	return viewer, nil
}
