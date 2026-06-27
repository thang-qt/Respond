package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

func scanDebateFeedRows(ctx context.Context, s *Store, rows *sql.Rows, locale string) ([]model.DebateFeedItem, error) {
	defer rows.Close()

	debates := make([]model.DebateFeedItem, 0)
	for rows.Next() {
		var item model.DebateFeedItem
		var contextValue, sideBAnon, outcome, winnerSide sql.NullString
		var sideAAnon string
		var sideARevealed, sideBRevealed sql.NullBool
		var turnDeadline, startedAt, endedAt sql.NullTime
		var sideARatingDelta, sideBRatingDelta sql.NullInt64
		var latestNumber sql.NullInt64
		var latestSide, latestAnon, latestContent sql.NullString
		var latestCreated sql.NullTime
		var sideAUsername string
		var sideARating int
		var sideBUsername sql.NullString
		var sideBRating sql.NullInt64
		var viewerUpvoted bool
		var isFollowing bool
		var isParticipant bool
		var promptID sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.Topic,
			&item.TimeMode,
			&item.TurnLimit,
			&contextValue,
			&item.Status,
			&sideAAnon,
			&sideBAnon,
			&sideARevealed,
			&sideBRevealed,
			&outcome,
			&winnerSide,
			&item.CurrentTurnSide,
			&item.TurnCount,
			&turnDeadline,
			&item.UpvoteCount,
			&item.CommentCount,
			&promptID,
			&item.CreatedAt,
			&startedAt,
			&endedAt,
			&sideARatingDelta,
			&sideBRatingDelta,
			&item.ModerationPending,
			&item.Hidden,
			&latestNumber,
			&latestSide,
			&latestAnon,
			&latestContent,
			&latestCreated,
			&sideAUsername,
			&sideARating,
			&sideBUsername,
			&sideBRating,
			&viewerUpvoted,
			&isFollowing,
			&isParticipant,
		); err != nil {
			return nil, fmt.Errorf("scan debate: %w", err)
		}

		item.SideA.AnonymousID = &sideAAnon
		if contextValue.Valid {
			item.Context = &contextValue.String
		}
		if sideBAnon.Valid {
			anon := sideBAnon.String
			item.SideB.AnonymousID = &anon
		}
		item.SideA.Revealed = sideARevealed.Valid && sideARevealed.Bool
		item.SideB.Revealed = sideBRevealed.Valid && sideBRevealed.Bool

		if outcome.Valid {
			item.Outcome = &outcome.String
		}
		if winnerSide.Valid {
			item.WinnerSide = &winnerSide.String
		}
		if turnDeadline.Valid {
			item.TurnDeadline = &turnDeadline.Time
		}
		if startedAt.Valid {
			item.StartedAt = &startedAt.Time
		}
		if endedAt.Valid {
			item.EndedAt = &endedAt.Time
		}
		if sideARatingDelta.Valid {
			delta := int(sideARatingDelta.Int64)
			item.SideARatingDelta = &delta
		}
		if sideBRatingDelta.Valid {
			delta := int(sideBRatingDelta.Int64)
			item.SideBRatingDelta = &delta
		}
		if latestNumber.Valid && latestSide.Valid && latestAnon.Valid && latestContent.Valid && latestCreated.Valid {
			item.LatestTurn = &model.TurnBrief{
				TurnNumber:  int(latestNumber.Int64),
				Side:        latestSide.String,
				AnonymousID: latestAnon.String,
				Content:     latestContent.String,
				CreatedAt:   latestCreated.Time,
			}
		}

		item.IsDailyPrompt = promptID.Valid
		item.SpectatorCount = 0
		item.SideA.User = selectSideUser(item.Status, item.SideA.Revealed, &model.UserSummary{
			Username: sideAUsername,
			Rating:   sideARating,
		})
		if sideBUsername.Valid && sideBRating.Valid {
			item.SideB.User = &model.UserSummary{
				Username: sideBUsername.String,
				Rating:   int(sideBRating.Int64),
			}
		}
		item.SideB.User = selectSideUser(item.Status, item.SideB.Revealed, item.SideB.User)
		item.ViewerHasUpvoted = viewerUpvoted
		item.IsFollowing = isFollowing
		item.ViewerIsParticipant = isParticipant

		debates = append(debates, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate debates: %w", err)
	}

	return attachTagsToDebates(ctx, s, debates, locale)
}

func attachTagsToDebates(ctx context.Context, s *Store, debates []model.DebateFeedItem, locale string) ([]model.DebateFeedItem, error) {
	debateIDs := make([]uuid.UUID, 0, len(debates))
	for _, debate := range debates {
		debateIDs = append(debateIDs, debate.ID)
	}
	tagsByDebate, err := s.ListTagsByDebateIDsLocalized(ctx, debateIDs, locale)
	if err != nil {
		return nil, err
	}
	for i := range debates {
		debates[i].Tags = tagsByDebate[debates[i].ID]
	}
	return debates, nil
}
