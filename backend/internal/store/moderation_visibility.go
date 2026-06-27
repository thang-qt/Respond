package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

func (s *Store) ListHiddenContent(ctx context.Context, params ListHiddenContentParams) ([]model.HiddenContentItem, error) {
	limit := params.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}

	args := []any{limit}
	filter := ""
	if params.TargetType != "" {
		args = append(args, params.TargetType)
		filter = "WHERE h.target_type = $2"
	}

	query := fmt.Sprintf(`
		SELECT
			h.target_type,
			h.target_id,
			h.debate_id,
			h.debate_slug,
			h.turn_number,
			h.target_author_id,
			h.target_author_username,
			h.content,
			h.hidden_at
		FROM (
			SELECT
				'debate'::text AS target_type,
				d.id AS target_id,
				d.id AS debate_id,
				d.slug AS debate_slug,
				NULL::integer AS turn_number,
				d.side_a_user_id AS target_author_id,
				u.username AS target_author_username,
				d.topic AS content,
				(
					SELECT ma.created_at
					FROM moderation_actions ma
					WHERE ma.target_type = 'debate'
						AND ma.target_id = d.id
						AND ma.action_type = 'hide_content'
					ORDER BY ma.created_at DESC
					LIMIT 1
				) AS hidden_at
			FROM debates d
			LEFT JOIN users u ON u.id = d.side_a_user_id
			WHERE d.hidden = true

			UNION ALL

			SELECT
				'turn'::text AS target_type,
				t.id AS target_id,
				t.debate_id AS debate_id,
				d.slug AS debate_slug,
				t.turn_number AS turn_number,
				t.user_id AS target_author_id,
				u.username AS target_author_username,
				t.content AS content,
				(
					SELECT ma.created_at
					FROM moderation_actions ma
					WHERE ma.target_type = 'turn'
						AND ma.target_id = t.id
						AND ma.action_type = 'hide_content'
					ORDER BY ma.created_at DESC
					LIMIT 1
				) AS hidden_at
			FROM turns t
			JOIN debates d ON d.id = t.debate_id
			LEFT JOIN users u ON u.id = t.user_id
			WHERE t.hidden = true

			UNION ALL

			SELECT
				'comment'::text AS target_type,
				c.id AS target_id,
				c.debate_id AS debate_id,
				d.slug AS debate_slug,
				NULL::integer AS turn_number,
				c.user_id AS target_author_id,
				u.username AS target_author_username,
				c.content AS content,
				(
					SELECT ma.created_at
					FROM moderation_actions ma
					WHERE ma.target_type = 'comment'
						AND ma.target_id = c.id
						AND ma.action_type = 'hide_content'
					ORDER BY ma.created_at DESC
					LIMIT 1
				) AS hidden_at
			FROM comments c
			JOIN debates d ON d.id = c.debate_id
			LEFT JOIN users u ON u.id = c.user_id
			WHERE c.hidden = true
		) h
		%s
		ORDER BY h.hidden_at DESC NULLS LAST
		LIMIT $1
	`, filter)

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list hidden content: %w", err)
	}
	defer rows.Close()

	items := make([]model.HiddenContentItem, 0)
	for rows.Next() {
		var item model.HiddenContentItem
		var debateID sql.NullString
		var debateSlug sql.NullString
		var turnNumber sql.NullInt64
		var targetAuthorID sql.NullString
		var targetAuthorUsername sql.NullString
		var hiddenAt sql.NullTime

		if err := rows.Scan(
			&item.TargetType,
			&item.TargetID,
			&debateID,
			&debateSlug,
			&turnNumber,
			&targetAuthorID,
			&targetAuthorUsername,
			&item.Content,
			&hiddenAt,
		); err != nil {
			return nil, fmt.Errorf("scan hidden content item: %w", err)
		}

		if debateID.Valid {
			parsed, err := uuid.Parse(debateID.String)
			if err != nil {
				return nil, fmt.Errorf("parse hidden debate id: %w", err)
			}
			item.DebateID = &parsed
		}
		if debateSlug.Valid {
			s := debateSlug.String
			item.DebateSlug = &s
		}
		if turnNumber.Valid {
			tn := int(turnNumber.Int64)
			item.TurnNumber = &tn
		}
		if targetAuthorID.Valid && targetAuthorUsername.Valid {
			parsed, err := uuid.Parse(targetAuthorID.String)
			if err != nil {
				return nil, fmt.Errorf("parse hidden target author id: %w", err)
			}
			item.TargetAuthor = &model.ReportUserRef{ID: parsed, Username: targetAuthorUsername.String}
		}
		if hiddenAt.Valid {
			t := hiddenAt.Time
			item.HiddenAt = &t
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate hidden content items: %w", err)
	}

	return items, nil
}

func (s *Store) RestoreHiddenTarget(ctx context.Context, reviewerID uuid.UUID, targetType string, targetID uuid.UUID, note *string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin hidden restore tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	note, err = normalizeRequiredModerationNote(note, "missing note for hidden restore")
	if err != nil {
		return err
	}

	if err := setTargetHiddenTx(ctx, tx, targetType, targetID, false); err != nil {
		return err
	}

	if err := insertModerationActionTx(ctx, tx, reviewerID, "restore_content", targetType, targetID, nil, map[string]string{"source": "hidden_queue", "resolution": "restore"}, *note); err != nil {
		return fmt.Errorf("insert hidden queue restore action: %w", err)
	}

	if err := notifyModerationTargetTx(ctx, s, tx, reviewerID, targetType, targetID, "restore", note, "create hidden restore"); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit hidden restore tx: %w", err)
	}

	return nil
}

func (s *Store) ModerateTargetVisibility(ctx context.Context, reviewerID uuid.UUID, targetType string, targetID uuid.UUID, resolution string, note *string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin direct moderation tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	note, err = normalizeRequiredModerationNote(note, "missing note for direct moderation")
	if err != nil {
		return err
	}

	actionType := ""
	hidden := false
	switch resolution {
	case "hide":
		actionType = "hide_content"
		hidden = true
	case "restore":
		actionType = "restore_content"
		hidden = false
	default:
		return fmt.Errorf("invalid resolution")
	}

	if err := setTargetHiddenTx(ctx, tx, targetType, targetID, hidden); err != nil {
		return err
	}

	if err := insertModerationActionTx(ctx, tx, reviewerID, actionType, targetType, targetID, nil, map[string]string{"source": "debate_view", "resolution": resolution}, *note); err != nil {
		return fmt.Errorf("insert direct moderation action: %w", err)
	}

	if err := notifyModerationTargetTx(ctx, s, tx, reviewerID, targetType, targetID, resolution, note, "create direct moderation"); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit direct moderation tx: %w", err)
	}

	return nil
}
