package store

import (
	"context"
	"fmt"

	"respond/internal/i18n"
	"respond/internal/model"
)

func (s *Store) SearchTags(ctx context.Context, query string, limit int) ([]model.Tag, error) {
	return s.SearchTagsLocalized(ctx, query, limit, i18n.DefaultLocale)
}

func (s *Store) SearchTagsLocalized(ctx context.Context, query string, limit int, locale string) ([]model.Tag, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	const q = `
		SELECT t.id, t.slug, t.name
		FROM tags t
		WHERE
			lower(t.name) LIKE '%' || lower($1) || '%'
			OR lower(t.slug) LIKE '%' || lower($1) || '%'
			OR similarity(lower(t.name), lower($1)) > 0.2
			OR similarity(lower(t.slug), lower($1)) > 0.2
		ORDER BY
			CASE
				WHEN lower(t.slug) = lower($1) THEN 0
				WHEN lower(t.name) = lower($1) THEN 1
				WHEN lower(t.slug) LIKE lower($1) || '%' THEN 2
				WHEN lower(t.name) LIKE lower($1) || '%' THEN 3
				ELSE 4
			END,
			GREATEST(
				similarity(lower(t.slug), lower($1)),
				similarity(lower(t.name), lower($1))
			) DESC,
			t.name ASC
		LIMIT $2
	`

	rows, err := s.DB.QueryContext(ctx, q, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search tags: %w", err)
	}
	defer rows.Close()

	tags := make([]model.Tag, 0, limit)
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.ID, &tag.Slug, &tag.Name); err != nil {
			return nil, fmt.Errorf("scan searched tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate searched tags: %w", err)
	}

	return localizeTags(locale, tags), nil
}
