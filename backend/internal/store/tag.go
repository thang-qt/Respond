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

func localizeTags(locale string, tags []model.Tag) []model.Tag {
	for i := range tags {
		tags[i].Name = i18n.LocalizeTagName(locale, tags[i].Slug, tags[i].Name)
	}
	return tags
}

func (s *Store) ListTags(ctx context.Context) ([]model.Tag, error) {
	return s.ListTagsLocalized(ctx, i18n.DefaultLocale)
}

func (s *Store) ListTagsLocalized(ctx context.Context, locale string) ([]model.Tag, error) {
	const query = `
		SELECT t.id, t.slug, t.name
		FROM tags t
		ORDER BY
			CASE WHEN t.display_order IS NULL THEN 1 ELSE 0 END,
			t.display_order ASC,
			t.name ASC
	`

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.ID, &tag.Slug, &tag.Name); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tag.Name = i18n.LocalizeTagName(locale, tag.Slug, tag.Name)
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}

	return tags, nil
}

func (s *Store) GetTagByID(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	const query = `
		SELECT id, slug, name
		FROM tags
		WHERE id = $1
	`

	var tag model.Tag
	if err := s.DB.QueryRowContext(ctx, query, id).Scan(&tag.ID, &tag.Slug, &tag.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Tag{}, ErrNotFound
		}
		return model.Tag{}, fmt.Errorf("get tag: %w", err)
	}

	return tag, nil
}

func (s *Store) GetTagBySlug(ctx context.Context, slug string) (model.Tag, error) {
	return s.GetTagBySlugLocalized(ctx, slug, i18n.DefaultLocale)
}

func (s *Store) GetTagBySlugLocalized(ctx context.Context, slug, locale string) (model.Tag, error) {
	const query = `
		SELECT id, slug, name
		FROM tags
		WHERE slug = $1
	`

	var tag model.Tag
	if err := s.DB.QueryRowContext(ctx, query, slug).Scan(&tag.ID, &tag.Slug, &tag.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Tag{}, ErrNotFound
		}
		return model.Tag{}, fmt.Errorf("get tag by slug: %w", err)
	}

	tag.Name = i18n.LocalizeTagName(locale, tag.Slug, tag.Name)
	return tag, nil
}

func (s *Store) ListTagsByDebateID(ctx context.Context, debateID uuid.UUID) ([]model.Tag, error) {
	return s.ListTagsByDebateIDLocalized(ctx, debateID, i18n.DefaultLocale)
}

func (s *Store) ListTagsByDebateIDLocalized(ctx context.Context, debateID uuid.UUID, locale string) ([]model.Tag, error) {
	const query = `
		SELECT t.id, t.slug, t.name
		FROM debate_tags dt
		JOIN tags t ON t.id = dt.tag_id
		WHERE dt.debate_id = $1
		ORDER BY
			CASE WHEN t.display_order IS NULL THEN 1 ELSE 0 END,
			t.display_order ASC,
			t.name ASC
	`

	rows, err := s.DB.QueryContext(ctx, query, debateID)
	if err != nil {
		return nil, fmt.Errorf("list debate tags: %w", err)
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.ID, &tag.Slug, &tag.Name); err != nil {
			return nil, fmt.Errorf("scan debate tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate debate tags: %w", err)
	}

	return localizeTags(locale, tags), nil
}

func (s *Store) ListTagsByDebateIDs(ctx context.Context, debateIDs []uuid.UUID) (map[uuid.UUID][]model.Tag, error) {
	return s.ListTagsByDebateIDsLocalized(ctx, debateIDs, i18n.DefaultLocale)
}

func (s *Store) ListTagsByDebateIDsLocalized(ctx context.Context, debateIDs []uuid.UUID, locale string) (map[uuid.UUID][]model.Tag, error) {
	if len(debateIDs) == 0 {
		return map[uuid.UUID][]model.Tag{}, nil
	}

	const query = `
		SELECT dt.debate_id, t.id, t.slug, t.name
		FROM debate_tags dt
		JOIN tags t ON t.id = dt.tag_id
		WHERE dt.debate_id = ANY($1::uuid[])
		ORDER BY
			dt.debate_id,
			CASE WHEN t.display_order IS NULL THEN 1 ELSE 0 END,
			t.display_order ASC,
			t.name ASC
	`

	rows, err := s.DB.QueryContext(ctx, query, uuidArrayArg(debateIDs))
	if err != nil {
		return nil, fmt.Errorf("list debate tags by ids: %w", err)
	}
	defer rows.Close()

	tagsByDebate := make(map[uuid.UUID][]model.Tag, len(debateIDs))
	for _, debateID := range debateIDs {
		tagsByDebate[debateID] = []model.Tag{}
	}

	for rows.Next() {
		var (
			debateID uuid.UUID
			tag      model.Tag
		)
		if err := rows.Scan(&debateID, &tag.ID, &tag.Slug, &tag.Name); err != nil {
			return nil, fmt.Errorf("scan debate tag by ids: %w", err)
		}
		tag.Name = i18n.LocalizeTagName(locale, tag.Slug, tag.Name)
		tagsByDebate[debateID] = append(tagsByDebate[debateID], tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate debate tags by ids: %w", err)
	}

	return tagsByDebate, nil
}

func (s *Store) CountTagsByIDs(ctx context.Context, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	const query = `SELECT COUNT(*) FROM tags WHERE id = ANY($1::uuid[])`
	var count int
	if err := s.DB.QueryRowContext(ctx, query, uuidArrayArg(ids)).Scan(&count); err != nil {
		return 0, fmt.Errorf("count tags by ids: %w", err)
	}
	return count, nil
}

func (s *Store) CountTagsBySlugs(ctx context.Context, slugs []string) (int, error) {
	if len(slugs) == 0 {
		return 0, nil
	}

	const query = `SELECT COUNT(*) FROM tags WHERE slug = ANY($1)`
	var count int
	if err := s.DB.QueryRowContext(ctx, query, sqlStringArrayArg(slugs)).Scan(&count); err != nil {
		return 0, fmt.Errorf("count tags by slugs: %w", err)
	}
	return count, nil
}

func (s *Store) ListUserTagFollows(ctx context.Context, userID uuid.UUID) ([]model.Tag, error) {
	return s.ListUserTagFollowsLocalized(ctx, userID, i18n.DefaultLocale)
}

func (s *Store) ListUserTagFollowsLocalized(ctx context.Context, userID uuid.UUID, locale string) ([]model.Tag, error) {
	const query = `
		SELECT t.id, t.slug, t.name
		FROM user_tag_follows utf
		JOIN tags t ON t.id = utf.tag_id
		WHERE utf.user_id = $1
		ORDER BY
			CASE WHEN t.display_order IS NULL THEN 1 ELSE 0 END,
			t.display_order ASC,
			t.name ASC
	`

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list user tag follows: %w", err)
	}
	defer rows.Close()

	follows := []model.Tag{}
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.ID, &tag.Slug, &tag.Name); err != nil {
			return nil, fmt.Errorf("scan user tag follow: %w", err)
		}
		follows = append(follows, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user tag follows: %w", err)
	}

	return localizeTags(locale, follows), nil
}

func (s *Store) ReplaceUserTagFollows(ctx context.Context, userID uuid.UUID, tagIDs []uuid.UUID) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin replace user tag follows tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM user_tag_follows
		WHERE user_id = $1
	`, userID); err != nil {
		return fmt.Errorf("delete user tag follows: %w", err)
	}

	for _, tagID := range tagIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO user_tag_follows (user_id, tag_id)
			VALUES ($1, $2)
		`, userID, tagID); err != nil {
			return fmt.Errorf("insert user tag follows: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit replace user tag follows tx: %w", err)
	}

	return nil
}
