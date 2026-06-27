package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

func (s *Store) GetReportByID(ctx context.Context, reportID uuid.UUID) (model.ReportDetail, error) {
	var (
		detail               model.ReportDetail
		details              sql.NullString
		resolution           sql.NullString
		reviewedBy           sql.NullString
		reviewedAt           sql.NullTime
		resolutionNote       sql.NullString
		reporterID           sql.NullString
		reporterName         sql.NullString
		targetAuthorID       sql.NullString
		targetAuthorUsername sql.NullString
		debateID             sql.NullString
		debateSlug           sql.NullString
		turnNumber           sql.NullInt64
	)

	err := s.DB.QueryRowContext(ctx, `
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
			r.reviewed_by_user_id,
			r.reviewed_at,
			r.resolution_note,
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
		WHERE r.id = $1
	`, reportID).Scan(
		&detail.ID,
		&detail.TargetType,
		&detail.TargetID,
		&detail.Reason,
		&details,
		&detail.Status,
		&resolution,
		&detail.TrustedReport,
		&detail.CreatedAt,
		&reviewedBy,
		&reviewedAt,
		&resolutionNote,
		&reporterID,
		&reporterName,
		&targetAuthorID,
		&targetAuthorUsername,
		&debateID,
		&debateSlug,
		&turnNumber,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ReportDetail{}, ErrReportNotFound
		}
		return model.ReportDetail{}, fmt.Errorf("get report: %w", err)
	}

	if details.Valid {
		detail.Details = &details.String
	}
	if reviewedBy.Valid {
		parsed, err := uuid.Parse(reviewedBy.String)
		if err != nil {
			return model.ReportDetail{}, fmt.Errorf("parse reviewed by: %w", err)
		}
		detail.ReviewedByUserID = &parsed
	}
	if reviewedAt.Valid {
		t := reviewedAt.Time
		detail.ReviewedAt = &t
	}
	if resolution.Valid {
		s := resolution.String
		detail.Resolution = &s
	}
	if resolutionNote.Valid {
		s := resolutionNote.String
		detail.ResolutionNote = &s
	}
	if reporterID.Valid && reporterName.Valid {
		parsed, err := uuid.Parse(reporterID.String)
		if err != nil {
			return model.ReportDetail{}, fmt.Errorf("parse reporter id: %w", err)
		}
		detail.Reporter = &model.ReportUserRef{ID: parsed, Username: reporterName.String}
	}
	if targetAuthorID.Valid && targetAuthorUsername.Valid {
		parsed, err := uuid.Parse(targetAuthorID.String)
		if err != nil {
			return model.ReportDetail{}, fmt.Errorf("parse target author id: %w", err)
		}
		detail.TargetAuthor = &model.ReportUserRef{ID: parsed, Username: targetAuthorUsername.String}
	}
	if debateID.Valid {
		parsed, err := uuid.Parse(debateID.String)
		if err != nil {
			return model.ReportDetail{}, fmt.Errorf("parse debate id: %w", err)
		}
		detail.DebateID = &parsed
	}
	if debateSlug.Valid {
		s := debateSlug.String
		detail.DebateSlug = &s
	}
	if turnNumber.Valid {
		tn := int(turnNumber.Int64)
		detail.TurnNumber = &tn
	}

	target, err := getReportTargetContent(ctx, s.DB, detail.TargetType, detail.TargetID)
	if err != nil {
		if errors.Is(err, ErrReportTargetNotFound) {
			return model.ReportDetail{}, ErrReportTargetNotFound
		}
		return model.ReportDetail{}, err
	}
	detail.Target = target

	return detail, nil
}
