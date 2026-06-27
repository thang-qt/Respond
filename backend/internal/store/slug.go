package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/gosimple/slug"
)

const maxSlugLength = 80

func slugify(value string) string {
	slug.MaxLength = maxSlugLength
	return strings.Trim(slug.Make(value), "-")
}

func ensureUniqueDebateSlug(ctx context.Context, tx *sql.Tx, baseSlug string) (string, error) {
	if baseSlug == "" {
		baseSlug = "debate"
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT slug
		FROM debates
		WHERE slug = $1 OR slug LIKE $1 || '-%'
	`, baseSlug)
	if err != nil {
		return "", fmt.Errorf("query debate slugs: %w", err)
	}
	defer rows.Close()

	usedBase := false
	maxSuffix := 1
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return "", fmt.Errorf("scan debate slug: %w", err)
		}
		if slug == baseSlug {
			usedBase = true
			continue
		}
		if strings.HasPrefix(slug, baseSlug+"-") {
			suffix := strings.TrimPrefix(slug, baseSlug+"-")
			if num, err := strconv.Atoi(suffix); err == nil {
				if num >= maxSuffix {
					maxSuffix = num + 1
				}
			}
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate debate slugs: %w", err)
	}

	if !usedBase {
		return baseSlug, nil
	}
	return fmt.Sprintf("%s-%d", baseSlug, maxSuffix), nil
}
