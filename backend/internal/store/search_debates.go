package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

var ErrInvalidSearchSort = errors.New("invalid search sort")

type SearchDebatesParams struct {
	Query    string
	Sort     string
	TagSlugs []string
	TagMode  string
	Page     int
	PerPage  int
	ViewerID *uuid.UUID
	Locale   string
}

func (s *Store) SearchDebates(ctx context.Context, params SearchDebatesParams) ([]model.DebateFeedItem, int, error) {
	pagination := normalizePagination(params.Page, params.PerPage, 20, 50)
	perPage := pagination.PerPage

	sortBy := params.Sort
	if sortBy == "" {
		sortBy = "relevance"
	}

	args := make([]any, 0, 6)
	args = append(args, params.Query)

	whereParts := []string{
		"d.status IN ('waiting', 'active', 'pending_extension', 'waiting_replacement', 'finished')",
		"d.hidden = false",
		"NOT (d.status = 'waiting' AND d.invited_user_id IS NOT NULL)",
		"s.document @@ websearch_to_tsquery('english', $1)",
	}
	whereParts, args = appendDebateTagFilter(whereParts, args, params.TagSlugs, params.TagMode)
	if params.ViewerID != nil {
		args = append(args, *params.ViewerID)
		whereParts = appendDebateViewerBlockFilter(whereParts, len(args))
	}
	whereSQL := joinWhereParts(whereParts)

	orderSQL := ""
	switch sortBy {
	case "relevance":
		orderSQL = "ORDER BY ts_rank_cd(s.document, websearch_to_tsquery('english', $1)) DESC, d.created_at DESC"
	case "new":
		orderSQL = "ORDER BY d.created_at DESC"
	default:
		return nil, 0, ErrInvalidSearchSort
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM debates d
		JOIN debate_search_documents s ON s.debate_id = d.id
		%s
	`, whereSQL)

	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count search debates: %w", err)
	}

	var viewerArg any
	if params.ViewerID != nil {
		viewerArg = *params.ViewerID
	}

	args = append(args, viewerArg, perPage, pagination.Offset)
	query := debateFeedSelectSQL("JOIN debate_search_documents s ON s.debate_id = d.id", whereSQL, orderSQL, len(args)-2, len(args)-1, len(args))

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search debates: %w", err)
	}
	defer rows.Close()

	debates, err := scanDebateFeedRows(ctx, s, rows, params.Locale)
	if err != nil {
		return nil, 0, err
	}

	return debates, total, nil
}
