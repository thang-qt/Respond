package store

import (
	"context"
	"fmt"

	"respond/internal/model"
)

func (s *Store) ListDebates(ctx context.Context, params ListDebatesParams) ([]model.DebateFeedItem, int, error) {
	pagination := normalizePagination(params.Page, params.PerPage, 20, 50)
	perPage := pagination.PerPage

	feed := params.Feed
	if feed == "" {
		feed = "trending"
	}

	var whereParts []string
	args := make([]any, 0, 4)
	whereParts = append(whereParts, "NOT (d.status = 'waiting' AND d.invited_user_id IS NOT NULL)")
	whereParts = append(whereParts, "d.status <> 'expired'")

	whereParts, args = appendDebateTagFilter(whereParts, args, params.TagSlugs, params.TagMode)

	switch feed {
	case "trending":
		whereParts = append(whereParts, "d.status IN ('active', 'pending_extension', 'finished')")
	case "new":
		// No extra filter.
	case "live":
		whereParts = append(whereParts, "d.status = 'active'")
	case "needs_challenger":
		whereParts = append(whereParts, "d.status = 'waiting'")
	case "following":
		if params.ViewerID == nil {
			return nil, 0, ErrInvalidFeed
		}
		whereParts = append(whereParts, fmt.Sprintf("EXISTS (SELECT 1 FROM follows f WHERE f.user_id = $%d AND f.debate_id = d.id)", len(args)+1))
		args = append(args, *params.ViewerID)
	case "following_tags":
		if params.ViewerID == nil {
			return nil, 0, ErrInvalidFeed
		}
		whereParts = append(whereParts, "d.status IN ('active', 'pending_extension', 'finished')")
		whereParts = append(whereParts, fmt.Sprintf(`EXISTS (
			SELECT 1
			FROM debate_tags dt
			JOIN user_tag_follows utf ON utf.tag_id = dt.tag_id
			WHERE utf.user_id = $%d
				AND dt.debate_id = d.id
		)`, len(args)+1))
		args = append(args, *params.ViewerID)
	default:
		return nil, 0, ErrInvalidFeed
	}

	whereParts, args = appendDebateViewerVisibilityFilters(whereParts, args, params.ViewerID)
	whereSQL := joinWhereParts(whereParts)

	orderSQL := ""
	switch feed {
	case "trending":
		orderSQL = debateTrendingOrderSQL()
	case "new":
		orderSQL = "ORDER BY d.created_at DESC"
	case "live":
		orderSQL = fmt.Sprintf("ORDER BY %s DESC", debateLastActivitySQL())
	case "needs_challenger":
		orderSQL = "ORDER BY d.created_at DESC"
	case "following":
		orderSQL = fmt.Sprintf("ORDER BY (SELECT created_at FROM follows f WHERE f.user_id = $%d AND f.debate_id = d.id) DESC", len(args))
	case "following_tags":
		orderSQL = debateTrendingOrderSQL()
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM debates d %s`, whereSQL)
	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count debates: %w", err)
	}

	var viewerArg any
	if params.ViewerID != nil {
		viewerArg = *params.ViewerID
	}

	args = append(args, viewerArg, perPage, pagination.Offset)
	query := debateFeedSelectSQL("", whereSQL, orderSQL, len(args)-2, len(args)-1, len(args))

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list debates: %w", err)
	}
	defer rows.Close()

	debates, err := scanDebateFeedRows(ctx, s, rows, params.Locale)
	if err != nil {
		return nil, 0, err
	}

	return debates, total, nil
}
