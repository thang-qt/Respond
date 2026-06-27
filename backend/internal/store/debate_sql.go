package store

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type ListDebatesParams struct {
	Feed     string
	TagSlugs []string
	TagMode  string
	Page     int
	PerPage  int
	ViewerID *uuid.UUID
	Locale   string
}

var ErrInvalidFeed = errors.New("invalid feed")
var ErrInvalidTagMode = errors.New("invalid tag mode")
var ErrInvalidExploreSort = errors.New("invalid explore sort")

type ListExploreParams struct {
	Sort     string
	TagSlugs []string
	TagMode  string
	Page     int
	PerPage  int
	ViewerID *uuid.UUID
	Locale   string
}

func debateLastActivitySQL() string {
	return "COALESCE((SELECT MAX(created_at) FROM turns t WHERE t.debate_id = d.id AND t.is_system = false), d.started_at, d.created_at)"
}

func debateTrendingOrderSQL() string {
	lastActivity := debateLastActivitySQL()
	return fmt.Sprintf(`ORDER BY
		(
			(
				d.upvote_count::double precision
				+ (
					SELECT COUNT(*)::double precision * 4.0
					FROM votes v2
					WHERE v2.target_type = 'debate'
						AND v2.target_id = d.id
						AND v2.created_at >= NOW() - INTERVAL '48 hours'
				)
				+ (
					SELECT COUNT(*)::double precision * 1.5
					FROM comments c2
					WHERE c2.debate_id = d.id
						AND c2.is_deleted = false
						AND c2.hidden = false
						AND c2.created_at >= NOW() - INTERVAL '48 hours'
				)
			)
			/ POWER((EXTRACT(EPOCH FROM (NOW() - %s)) / 3600.0) + 2.0, 1.5)
		) DESC,
		%s DESC,
		d.created_at DESC`, lastActivity, lastActivity)
}

func debateRisingOrderSQL() string {
	lastActivity := debateLastActivitySQL()
	return fmt.Sprintf(`ORDER BY
		(
			(
				SELECT COUNT(*)::double precision * 4.0
				FROM votes v2
				WHERE v2.target_type = 'debate'
					AND v2.target_id = d.id
					AND v2.created_at >= NOW() - INTERVAL '24 hours'
			)
			+ (
				SELECT COUNT(*)::double precision * 2.0
				FROM comments c2
				WHERE c2.debate_id = d.id
					AND c2.is_deleted = false
					AND c2.hidden = false
					AND c2.created_at >= NOW() - INTERVAL '24 hours'
			)
			+ (
				SELECT COUNT(*)::double precision
				FROM turns t2
				WHERE t2.debate_id = d.id
					AND t2.is_system = false
					AND t2.hidden = false
					AND t2.created_at >= NOW() - INTERVAL '24 hours'
			)
		) DESC,
		%s DESC,
		d.created_at DESC`, lastActivity)
}
