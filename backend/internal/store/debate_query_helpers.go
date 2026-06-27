package store

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func appendDebateTagFilter(whereParts []string, args []any, tagSlugs []string, tagMode string) ([]string, []any) {
	if tagMode == "" {
		tagMode = "any"
	}
	if len(tagSlugs) == 0 {
		return whereParts, args
	}

	args = append(args, sqlStringArrayArg(tagSlugs))
	tagFilterArgPos := len(args)
	if tagMode == "all" {
		whereParts = append(whereParts, fmt.Sprintf(`(
			SELECT COUNT(DISTINCT t.slug)
			FROM debate_tags dt
			JOIN tags t ON t.id = dt.tag_id
			WHERE dt.debate_id = d.id
				AND t.slug = ANY($%d)
		) = %d`, tagFilterArgPos, len(tagSlugs)))
	} else {
		whereParts = append(whereParts, fmt.Sprintf(`EXISTS (
			SELECT 1
			FROM debate_tags dt
			JOIN tags t ON t.id = dt.tag_id
			WHERE dt.debate_id = d.id
				AND t.slug = ANY($%d)
		)`, tagFilterArgPos))
	}
	return whereParts, args
}

func appendDebateViewerVisibilityFilters(whereParts []string, args []any, viewerID *uuid.UUID) ([]string, []any) {
	if viewerID == nil {
		return append(whereParts, "d.hidden = false"), args
	}

	args = append(args, *viewerID)
	viewerFilterPos := len(args)
	whereParts = append(whereParts, fmt.Sprintf("(d.hidden = false OR d.side_a_user_id = $%d OR d.side_b_user_id = $%d)", viewerFilterPos, viewerFilterPos))
	return appendDebateViewerBlockFilter(whereParts, viewerFilterPos), args
}

func appendDebateViewerBlockFilter(whereParts []string, viewerFilterPos int) []string {
	return append(whereParts, fmt.Sprintf(`(
		d.side_a_user_id = $%d
		OR d.side_b_user_id = $%d
		OR NOT EXISTS (
			SELECT 1
			FROM user_blocks ub
			WHERE (ub.blocker_id = $%d AND (ub.blocked_id = d.side_a_user_id OR ub.blocked_id = d.side_b_user_id))
			   OR (ub.blocked_id = $%d AND (ub.blocker_id = d.side_a_user_id OR ub.blocker_id = d.side_b_user_id))
		)
	)`, viewerFilterPos, viewerFilterPos, viewerFilterPos, viewerFilterPos))
}

func joinWhereParts(whereParts []string) string {
	if len(whereParts) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(whereParts, " AND ")
}
