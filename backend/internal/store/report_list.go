package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

func (s *Store) ListReports(ctx context.Context, params ListReportsParams) ([]model.Report, int, error) {
	pagination := normalizePagination(params.Page, params.PerPage, 20, 50)
	perPage := pagination.PerPage

	args := []any{}
	where := "WHERE 1=1"

	if params.Status != "" && params.Status != "all" {
		args = append(args, params.Status)
		where += fmt.Sprintf(" AND r.status = $%d::report_status", len(args))
	}
	if params.TargetType != "" {
		args = append(args, params.TargetType)
		where += fmt.Sprintf(" AND r.target_type = $%d::report_target", len(args))
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM reports r %s`, where)
	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count reports: %w", err)
	}

	args = append(args, perPage, pagination.Offset)
	query := fmt.Sprintf(`
		SELECT
			r.id,
			r.target_type,
			r.target_id,
			r.reason,
			r.details,
			r.status,
			r.resolution,
			r.trusted_report,
			r.created_at,
			reporter.id,
			reporter.username,
			target_author.id,
			target_author.username,
			d.id,
			d.slug,
			t.turn_number
		FROM reports r
		JOIN users reporter ON reporter.id = r.reporter_id
		LEFT JOIN debates rd ON r.target_type = 'debate'::report_target AND rd.id = r.target_id
		LEFT JOIN turns t ON r.target_type = 'turn'::report_target AND t.id = r.target_id
		LEFT JOIN comments c ON r.target_type = 'comment'::report_target AND c.id = r.target_id
		LEFT JOIN users target_author ON target_author.id = COALESCE(t.user_id, c.user_id, rd.side_a_user_id)
		LEFT JOIN debates d ON d.id = COALESCE(t.debate_id, c.debate_id, rd.id)
		%s
		ORDER BY r.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)-1, len(args))

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	items := make([]model.Report, 0)
	for rows.Next() {
		var item model.Report
		var details sql.NullString
		var resolution sql.NullString
		var reporterID sql.NullString
		var reporterUsername sql.NullString
		var targetAuthorID sql.NullString
		var targetAuthorUsername sql.NullString
		var debateID sql.NullString
		var debateSlug sql.NullString
		var turnNumber sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.TargetType,
			&item.TargetID,
			&item.Reason,
			&details,
			&item.Status,
			&resolution,
			&item.TrustedReport,
			&item.CreatedAt,
			&reporterID,
			&reporterUsername,
			&targetAuthorID,
			&targetAuthorUsername,
			&debateID,
			&debateSlug,
			&turnNumber,
		); err != nil {
			return nil, 0, fmt.Errorf("scan report: %w", err)
		}
		if details.Valid {
			item.Details = &details.String
		}
		if resolution.Valid {
			s := resolution.String
			item.Resolution = &s
		}
		if reporterID.Valid && reporterUsername.Valid {
			parsed, err := uuid.Parse(reporterID.String)
			if err != nil {
				return nil, 0, fmt.Errorf("parse reporter id: %w", err)
			}
			item.Reporter = &model.ReportUserRef{ID: parsed, Username: reporterUsername.String}
		}
		if targetAuthorID.Valid && targetAuthorUsername.Valid {
			parsed, err := uuid.Parse(targetAuthorID.String)
			if err != nil {
				return nil, 0, fmt.Errorf("parse target author id: %w", err)
			}
			item.TargetAuthor = &model.ReportUserRef{ID: parsed, Username: targetAuthorUsername.String}
		}
		if debateID.Valid {
			parsed, err := uuid.Parse(debateID.String)
			if err != nil {
				return nil, 0, fmt.Errorf("parse debate id: %w", err)
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
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate reports: %w", err)
	}

	return items, total, nil
}
