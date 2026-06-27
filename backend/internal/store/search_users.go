package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserSearchResult struct {
	ID           uuid.UUID
	Username     string
	Bio          string
	Rating       int
	Wins         int
	Losses       int
	Draws        int
	DebatesCount int
	CreatedAt    time.Time
	SharedTags   []string
}

type SearchUsersParams struct {
	Query    string
	TagSlugs []string
	TagMode  string
	Page     int
	PerPage  int
	ViewerID *uuid.UUID
}

func (s *Store) SearchUsers(ctx context.Context, params SearchUsersParams) ([]UserSearchResult, int, error) {
	return s.listRelevantUsers(ctx, listRelevantUsersParams{
		Query:    params.Query,
		TagSlugs: params.TagSlugs,
		TagMode:  params.TagMode,
		Page:     params.Page,
		PerPage:  params.PerPage,
		ViewerID: params.ViewerID,
	})
}

type ListExploreUsersParams struct {
	Query    string
	TagSlugs []string
	TagMode  string
	Page     int
	PerPage  int
	ViewerID *uuid.UUID
}

func (s *Store) ListExploreUsers(ctx context.Context, params ListExploreUsersParams) ([]UserSearchResult, int, error) {
	return s.listRelevantUsers(ctx, listRelevantUsersParams{
		Query:    params.Query,
		TagSlugs: params.TagSlugs,
		TagMode:  params.TagMode,
		Page:     params.Page,
		PerPage:  params.PerPage,
		ViewerID: params.ViewerID,
	})
}

type listRelevantUsersParams struct {
	Query    string
	TagSlugs []string
	TagMode  string
	Page     int
	PerPage  int
	ViewerID *uuid.UUID
}

func (s *Store) listRelevantUsers(ctx context.Context, params listRelevantUsersParams) ([]UserSearchResult, int, error) {
	pagination := normalizePagination(params.Page, params.PerPage, 20, 50)
	perPage := pagination.PerPage

	query := strings.TrimSpace(params.Query)
	tagMode := params.TagMode
	if tagMode == "" {
		tagMode = "any"
	}

	args := make([]any, 0, 4)
	viewerPos := 0
	if params.ViewerID != nil {
		args = append(args, *params.ViewerID)
		viewerPos = len(args)
	}

	queryPos := 0
	if query != "" {
		args = append(args, query)
		queryPos = len(args)
	}

	tagPos := 0
	if len(params.TagSlugs) > 0 {
		args = append(args, sqlStringArrayArg(params.TagSlugs))
		tagPos = len(args)
	}

	viewerTagPoolSQL := "viewer_tag_pool as (select null::uuid as tag_id where false)"
	if viewerPos > 0 {
		viewerTagPoolSQL = fmt.Sprintf(`viewer_tag_pool as (
			select utf.tag_id
			from user_tag_follows utf
			where utf.user_id = $%d
			union
			select dt.tag_id
			from debates d
			join debate_tags dt on dt.debate_id = d.id
			where d.created_at >= now() - interval '180 days'
				and (
					d.side_a_user_id = $%d
					or d.side_b_user_id = $%d
					or exists (
						select 1
						from comments c
						where c.debate_id = d.id
							and c.user_id = $%d
					)
					or exists (
						select 1
						from votes v
						where v.target_type = 'debate'
							and v.target_id = d.id
							and v.user_id = $%d
					)
				)
		)`, viewerPos, viewerPos, viewerPos, viewerPos, viewerPos)
	}

	withSQL := `
		with
		` + viewerTagPoolSQL + `,
		candidate_user_tags as (
			select d.side_a_user_id as user_id, dt.tag_id
			from debates d
			join debate_tags dt on dt.debate_id = d.id
			where d.hidden = false
			union all
			select d.side_b_user_id as user_id, dt.tag_id
			from debates d
			join debate_tags dt on dt.debate_id = d.id
			where d.hidden = false and d.side_b_user_id is not null
		),
		candidate_activity as (
			select
				u.id as user_id,
				count(*) filter (
					where d.created_at >= now() - interval '30 days'
						and d.hidden = false
				) as debates_30d,
				max(d.created_at) filter (where d.hidden = false) as last_debate_at,
				count(*) filter (
					where d.status = 'finished'
						and d.hidden = false
				) as debates_count
			from users u
			left join debates d
				on d.side_a_user_id = u.id
				or d.side_b_user_id = u.id
			group by u.id
		),
		candidate_overlap as (
			select cut.user_id, count(distinct cut.tag_id) as shared_tag_count
			from candidate_user_tags cut
			join viewer_tag_pool vtp on vtp.tag_id = cut.tag_id
			group by cut.user_id
		)
	`

	whereParts := []string{"1=1"}
	if viewerPos > 0 {
		whereParts = append(whereParts, fmt.Sprintf("u.id <> $%d", viewerPos))
		whereParts = append(whereParts, fmt.Sprintf(`not exists (
			select 1
			from user_blocks ub
			where (ub.blocker_id = $%d and ub.blocked_id = u.id)
			   or (ub.blocked_id = $%d and ub.blocker_id = u.id)
		)`, viewerPos, viewerPos))
	}
	if queryPos > 0 {
		whereParts = append(whereParts, fmt.Sprintf(`(
			lower(u.username) like '%%' || lower($%d) || '%%'
			or similarity(lower(u.username), lower($%d)) > 0.2
		)`, queryPos, queryPos))
	}
	if tagPos > 0 {
		switch tagMode {
		case "any":
			whereParts = append(whereParts, fmt.Sprintf(`exists (
				select 1
				from candidate_user_tags cut
				join tags t on t.id = cut.tag_id
				where cut.user_id = u.id
					and t.slug = any($%d)
			)`, tagPos))
		case "all":
			whereParts = append(whereParts, fmt.Sprintf(`(
				select count(distinct t.slug)
				from candidate_user_tags cut
				join tags t on t.id = cut.tag_id
				where cut.user_id = u.id
					and t.slug = any($%d)
			) = %d`, tagPos, len(params.TagSlugs)))
		default:
			return nil, 0, ErrInvalidTagMode
		}
	}

	whereSQL := "where " + strings.Join(whereParts, " and ")

	countQuery := withSQL + `
		select count(*)
		from users u
		left join candidate_activity a on a.user_id = u.id
		left join candidate_overlap o on o.user_id = u.id
		` + whereSQL

	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count searched users: %w", err)
	}

	orderBy := `
		order by (
			coalesce(o.shared_tag_count, 0) * 5
			+ least(coalesce(a.debates_30d, 0), 10)
			+ case
				when a.last_debate_at is not null and a.last_debate_at >= now() - interval '7 days' then 3
				else 0
			end
		) desc,
		u.rating desc,
		u.created_at desc`
	if queryPos > 0 {
		orderBy = fmt.Sprintf(`
		order by (
			coalesce(o.shared_tag_count, 0) * 5
			+ least(coalesce(a.debates_30d, 0), 10)
			+ case
				when a.last_debate_at is not null and a.last_debate_at >= now() - interval '7 days' then 3
				else 0
			end
			+ case
				when lower(u.username) = lower($%d) then 12
				when lower(u.username) like lower($%d) || '%%' then 8
				when lower(u.username) like '%%' || lower($%d) || '%%' then 4
				else 0
			end
			+ similarity(lower(u.username), lower($%d))
		) desc,
		u.rating desc,
		u.created_at desc`, queryPos, queryPos, queryPos, queryPos)
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	dataQuery := withSQL + `
		select
			u.id,
			u.username,
			u.bio,
			u.rating,
			u.wins,
			u.losses,
			u.draws,
			coalesce(a.debates_count, 0) as debates_count,
			u.created_at,
			coalesce((
				select array(
					select distinct t.slug
					from candidate_user_tags cut
					join viewer_tag_pool vtp on vtp.tag_id = cut.tag_id
					join tags t on t.id = cut.tag_id
					where cut.user_id = u.id
					order by t.slug
					limit 3
				)
			), '{}'::text[]) as shared_tags
		from users u
		left join candidate_activity a on a.user_id = u.id
		left join candidate_overlap o on o.user_id = u.id
		` + whereSQL + `
		` + orderBy + `
	` + fmt.Sprintf(" limit $%d offset $%d", limitPos, offsetPos)

	dataArgs := append(args, perPage, pagination.Offset)
	rows, err := s.DB.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("search users: %w", err)
	}
	defer rows.Close()

	users := make([]UserSearchResult, 0, perPage)
	for rows.Next() {
		var user UserSearchResult
		var sharedTags []string
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Bio,
			&user.Rating,
			&user.Wins,
			&user.Losses,
			&user.Draws,
			&user.DebatesCount,
			&user.CreatedAt,
			pq.Array(&sharedTags),
		); err != nil {
			return nil, 0, fmt.Errorf("scan searched user: %w", err)
		}
		user.SharedTags = sharedTags
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate searched users: %w", err)
	}

	return users, total, nil
}
