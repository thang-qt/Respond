package store

import (
	"context"
	"fmt"

	"respond/internal/model"
)

func (s *Store) ListExplore(ctx context.Context, params ListExploreParams) ([]model.DebateFeedItem, int, error) {
	pagination := normalizePagination(params.Page, params.PerPage, 20, 50)
	perPage := pagination.PerPage

	sortBy := params.Sort
	if sortBy == "" {
		sortBy = "hot"
	}

	var whereParts []string
	args := make([]any, 0, 4)
	whereParts = append(whereParts, "NOT (d.status = 'waiting' AND d.invited_user_id IS NOT NULL)")
	whereParts = append(whereParts, "d.status <> 'expired'")

	whereParts, args = appendDebateTagFilter(whereParts, args, params.TagSlugs, params.TagMode)

	whereParts, args = appendDebateViewerVisibilityFilters(whereParts, args, params.ViewerID)
	whereSQL := joinWhereParts(whereParts)

	orderSQL := ""
	switch sortBy {
	case "hot":
		orderSQL = debateTrendingOrderSQL()
	case "rising":
		orderSQL = debateRisingOrderSQL()
	case "new":
		orderSQL = "ORDER BY d.created_at DESC"
	case "random":
		orderSQL = "ORDER BY RANDOM()"
	default:
		return nil, 0, ErrInvalidExploreSort
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM debates d %s`, whereSQL)
	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count explore debates: %w", err)
	}

	var viewerArg any
	if params.ViewerID != nil {
		viewerArg = *params.ViewerID
	}

	offset := pagination.Offset
	if sortBy == "random" {
		offset = 0
	}

	args = append(args, viewerArg, perPage, offset)
	query := debateFeedSelectSQL("", whereSQL, orderSQL, len(args)-2, len(args)-1, len(args))

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list explore debates: %w", err)
	}
	defer rows.Close()

	debates, err := scanDebateFeedRows(ctx, s, rows, params.Locale)
	if err != nil {
		return nil, 0, err
	}

	return debates, total, nil
}
