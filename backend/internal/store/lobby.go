package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"respond/internal/model"
)

var ErrLobbyTagLimit = errors.New("lobby entry may have at most 15 tags")

// ChallengeLobbyEntry is a user's public standing-challenge signal.
type ChallengeLobbyEntry struct {
	Username  string     `json:"username"`
	Bio       string     `json:"bio"`
	Rating    int        `json:"rating"`
	Wins      int        `json:"wins"`
	Losses    int        `json:"losses"`
	Draws     int        `json:"draws"`
	BioNote   string     `json:"bio_note"`
	Tags      []model.Tag `json:"tags"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// UpsertLobbyEntry creates or replaces the authenticated user's lobby entry.
// tagIDs must contain 0–15 valid tag UUIDs.
func (s *Store) UpsertLobbyEntry(ctx context.Context, userID uuid.UUID, bioNote string, tagIDs []uuid.UUID) error {
	if len(tagIDs) > 15 {
		return ErrLobbyTagLimit
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin upsert lobby tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO challenge_lobby_entries (user_id, bio_note, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		ON CONFLICT (user_id) DO UPDATE
		  SET bio_note   = EXCLUDED.bio_note,
		      updated_at = EXCLUDED.updated_at
	`, userID, bioNote, now); err != nil {
		return fmt.Errorf("upsert lobby entry: %w", err)
	}

	// Replace tags: delete all and re-insert.
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM challenge_lobby_entry_tags WHERE user_id = $1
	`, userID); err != nil {
		return fmt.Errorf("delete lobby entry tags: %w", err)
	}

	for _, tagID := range tagIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO challenge_lobby_entry_tags (user_id, tag_id)
			VALUES ($1, $2)
		`, userID, tagID); err != nil {
			return fmt.Errorf("insert lobby entry tag: %w", err)
		}
	}

	return tx.Commit()
}

// DeleteLobbyEntry removes the user's lobby entry (no-op if not present).
func (s *Store) DeleteLobbyEntry(ctx context.Context, userID uuid.UUID) error {
	if _, err := s.DB.ExecContext(ctx, `
		DELETE FROM challenge_lobby_entries WHERE user_id = $1
	`, userID); err != nil {
		return fmt.Errorf("delete lobby entry: %w", err)
	}
	return nil
}

// GetMyLobbyEntry returns the authenticated user's own lobby entry.
func (s *Store) GetMyLobbyEntry(ctx context.Context, userID uuid.UUID) (*ChallengeLobbyEntry, error) {
	const q = `
		SELECT u.username, u.bio, u.rating, u.wins, u.losses, u.draws,
		       e.bio_note, e.created_at, e.updated_at
		FROM challenge_lobby_entries e
		JOIN users u ON u.id = e.user_id
		WHERE e.user_id = $1
	`
	var entry ChallengeLobbyEntry
	if err := s.DB.QueryRowContext(ctx, q, userID).Scan(
		&entry.Username, &entry.Bio, &entry.Rating,
		&entry.Wins, &entry.Losses, &entry.Draws,
		&entry.BioNote, &entry.CreatedAt, &entry.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get my lobby entry: %w", err)
	}
	tags, err := s.listLobbyEntryTags(ctx, userID)
	if err != nil {
		return nil, err
	}
	entry.Tags = tags
	return &entry, nil
}

// GetUserLobbyEntry returns any user's lobby entry by username,
// with block enforcement. Returns ErrNotFound if no entry exists or blocked.
func (s *Store) GetUserLobbyEntry(ctx context.Context, viewerID *uuid.UUID, targetUsername string) (*ChallengeLobbyEntry, error) {
	const q = `
		SELECT u.id, u.username, u.bio, u.rating, u.wins, u.losses, u.draws,
		       e.bio_note, e.created_at, e.updated_at
		FROM challenge_lobby_entries e
		JOIN users u ON u.id = e.user_id
		WHERE LOWER(u.username) = LOWER($1)
	`
	var (
		targetID uuid.UUID
		entry    ChallengeLobbyEntry
	)
	if err := s.DB.QueryRowContext(ctx, q, targetUsername).Scan(
		&targetID,
		&entry.Username, &entry.Bio, &entry.Rating,
		&entry.Wins, &entry.Losses, &entry.Draws,
		&entry.BioNote, &entry.CreatedAt, &entry.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user lobby entry: %w", err)
	}

	// Block check: if viewer is logged in, verify no block relationship.
	if viewerID != nil && *viewerID != targetID {
		blocked, err := s.IsEitherUserBlocked(ctx, *viewerID, targetID)
		if err != nil {
			return nil, err
		}
		if blocked {
			return nil, ErrNotFound // treat as not found — don't expose block
		}
	}

	tags, err := s.listLobbyEntryTags(ctx, targetID)
	if err != nil {
		return nil, err
	}
	entry.Tags = tags
	return &entry, nil
}

// ListLobbyEntriesParams holds filtering options for the lobby browse endpoint.
type ListLobbyEntriesParams struct {
	ViewerID *uuid.UUID
	TagSlugs []string
	TagMode  string // "any" | "all"
	Page     int
	PerPage  int
}

// ListLobbyEntries returns the paginated, block-filtered lobby.
func (s *Store) ListLobbyEntries(ctx context.Context, params ListLobbyEntriesParams) ([]ChallengeLobbyEntry, int, error) {
	page := params.Page
	if page < 1 {
		page = 1
	}
	perPage := params.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 50 {
		perPage = 50
	}

	args := []any{}
	argN := 1
	nextArg := func(v any) int {
		args = append(args, v)
		n := argN
		argN++
		return n
	}

	whereParts := []string{"1=1"}

	// Block filter (if viewer is authenticated)
	if params.ViewerID != nil {
		vn := nextArg(params.ViewerID)
		whereParts = append(whereParts, fmt.Sprintf(`
			NOT EXISTS (
				SELECT 1 FROM user_blocks ub
				WHERE (ub.blocker_id = $%d AND ub.blocked_id = e.user_id)
				   OR (ub.blocker_id = e.user_id AND ub.blocked_id = $%d)
			)`, vn, vn))
	}

	// Tag filter
	if len(params.TagSlugs) > 0 {
		tagMode := strings.ToLower(strings.TrimSpace(params.TagMode))
		tn := nextArg(sqlStringArrayArg(params.TagSlugs))
		if tagMode == "all" {
			countN := nextArg(len(params.TagSlugs))
			whereParts = append(whereParts, fmt.Sprintf(`
				(SELECT COUNT(DISTINCT t.slug)
				 FROM challenge_lobby_entry_tags clt
				 JOIN tags t ON t.id = clt.tag_id
				 WHERE clt.user_id = e.user_id AND t.slug = ANY($%d)
				) = $%d`, tn, countN))
		} else {
			whereParts = append(whereParts, fmt.Sprintf(`
				EXISTS (
					SELECT 1 FROM challenge_lobby_entry_tags clt
					JOIN tags t ON t.id = clt.tag_id
					WHERE clt.user_id = e.user_id AND t.slug = ANY($%d)
				)`, tn))
		}
	}

	whereSQL := "WHERE " + strings.Join(whereParts, " AND ")

	countQ := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM challenge_lobby_entries e
		JOIN users u ON u.id = e.user_id
		%s`, whereSQL)
	var total int
	if err := s.DB.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count lobby entries: %w", err)
	}

	limitN := nextArg(perPage)
	offsetN := nextArg((page - 1) * perPage)
	query := fmt.Sprintf(`
		SELECT e.user_id, u.username, u.bio, u.rating, u.wins, u.losses, u.draws,
		       e.bio_note, e.created_at, e.updated_at
		FROM challenge_lobby_entries e
		JOIN users u ON u.id = e.user_id
		%s
		ORDER BY e.updated_at DESC
		LIMIT $%d OFFSET $%d`, whereSQL, limitN, offsetN)

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list lobby entries: %w", err)
	}
	defer rows.Close()

	var entries []ChallengeLobbyEntry
	var userIDs []uuid.UUID
	for rows.Next() {
		var (
			entry  ChallengeLobbyEntry
			userID uuid.UUID
		)
		if err := rows.Scan(
			&userID,
			&entry.Username, &entry.Bio, &entry.Rating,
			&entry.Wins, &entry.Losses, &entry.Draws,
			&entry.BioNote, &entry.CreatedAt, &entry.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan lobby entry: %w", err)
		}
		entry.Tags = []model.Tag{}
		entries = append(entries, entry)
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate lobby entries: %w", err)
	}

	// Batch-load tags for all entries.
	if len(userIDs) > 0 {
		tagsByUser, err := s.listLobbyEntryTagsBatch(ctx, userIDs)
		if err != nil {
			return nil, 0, err
		}
		for i, uid := range userIDs {
			if t, ok := tagsByUser[uid]; ok {
				entries[i].Tags = t
			}
		}
	}

	return entries, total, nil
}

// listLobbyEntryTags returns tags for a single lobby entry user.
func (s *Store) listLobbyEntryTags(ctx context.Context, userID uuid.UUID) ([]model.Tag, error) {
	const q = `
		SELECT t.id, t.slug, t.name
		FROM challenge_lobby_entry_tags clt
		JOIN tags t ON t.id = clt.tag_id
		WHERE clt.user_id = $1
		ORDER BY t.name ASC
	`
	rows, err := s.DB.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list lobby entry tags: %w", err)
	}
	defer rows.Close()
	tags := []model.Tag{}
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.ID, &tag.Slug, &tag.Name); err != nil {
			return nil, fmt.Errorf("scan lobby entry tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// listLobbyEntryTagsBatch batch-loads tags for multiple lobby entry users.
func (s *Store) listLobbyEntryTagsBatch(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]model.Tag, error) {
	const q = `
		SELECT clt.user_id, t.id, t.slug, t.name
		FROM challenge_lobby_entry_tags clt
		JOIN tags t ON t.id = clt.tag_id
		WHERE clt.user_id = ANY($1::uuid[])
		ORDER BY clt.user_id, t.name ASC
	`
	rows, err := s.DB.QueryContext(ctx, q, uuidArrayArg(userIDs))
	if err != nil {
		return nil, fmt.Errorf("batch list lobby entry tags: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]model.Tag)
	for rows.Next() {
		var (
			userID uuid.UUID
			tag    model.Tag
		)
		if err := rows.Scan(&userID, &tag.ID, &tag.Slug, &tag.Name); err != nil {
			return nil, fmt.Errorf("scan batch lobby entry tag: %w", err)
		}
		result[userID] = append(result[userID], tag)
	}
	return result, rows.Err()
}
