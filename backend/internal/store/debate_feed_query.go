package store

import "fmt"

func debateFeedSelectSQL(extraJoinSQL, whereSQL, orderSQL string, viewerArgPos, limitArgPos, offsetArgPos int) string {
	return fmt.Sprintf(`
		SELECT d.id, d.slug, d.topic, d.time_mode, d.turn_limit, d.context, d.status,
			d.side_a_anonymous_id, d.side_b_anonymous_id, d.side_a_revealed, d.side_b_revealed,
			d.outcome, d.winner_side, d.current_turn_side, d.turn_count, d.turn_deadline,
			d.upvote_count, d.comment_count, d.prompt_id, d.created_at, d.started_at, d.ended_at,
			d.side_a_rating_delta, d.side_b_rating_delta, d.moderation_pending, d.hidden,
			lt.turn_number, lt.side, lt.anonymous_id, lt.content, lt.created_at,
			ua.username, ua.rating,
			ub.username, ub.rating,
			CASE WHEN v.id IS NULL THEN false ELSE true END AS viewer_has_upvoted,
			EXISTS (
				SELECT 1 FROM follows f
				WHERE f.user_id = $%d AND f.debate_id = d.id
			) AS is_following,
			CASE
				WHEN $%d IS NULL THEN false
				WHEN d.side_a_user_id = $%d OR d.side_b_user_id = $%d THEN true
				ELSE false
			END AS viewer_is_participant
		FROM debates d
		%s
		JOIN users ua ON ua.id = d.side_a_user_id
		LEFT JOIN users ub ON ub.id = d.side_b_user_id
		LEFT JOIN votes v
			ON v.target_type = 'debate'
			AND v.target_id = d.id
			AND v.user_id = $%d
		LEFT JOIN LATERAL (
			SELECT
				t.turn_number,
				t.side,
				t.anonymous_id,
				CASE
					WHEN d.hidden THEN 'This content has been flagged for review.'
					WHEN t.hidden THEN 'This content has been flagged for review.'
					WHEN char_length(t.content) > 360 THEN left(convert_from(convert_to(t.content, 'UTF8'), 'UTF8'), 360) || '...'
					ELSE t.content
				END AS content,
				t.created_at
			FROM turns t
			WHERE t.debate_id = d.id AND t.is_system = false
			ORDER BY t.turn_number DESC
			LIMIT 1
		) lt ON true
		%s
		%s
		LIMIT $%d OFFSET $%d
	`,
		viewerArgPos,
		viewerArgPos,
		viewerArgPos,
		viewerArgPos,
		extraJoinSQL,
		viewerArgPos,
		whereSQL,
		orderSQL,
		limitArgPos,
		offsetArgPos,
	)
}
