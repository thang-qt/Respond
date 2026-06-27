package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"respond/internal/model"
)

func (s *Store) CreateReport(ctx context.Context, params CreateReportParams) (model.Report, bool, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.Report{}, false, fmt.Errorf("begin report tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	rateLimited, err := isReportRateLimited(ctx, tx, params.ReporterID)
	if err != nil {
		return model.Report{}, false, fmt.Errorf("check report rate: %w", err)
	}
	if rateLimited {
		return model.Report{}, false, ErrReportRateLimited
	}

	exists, err := targetExistsTx(ctx, tx, params.TargetType, params.TargetID)
	if err != nil {
		return model.Report{}, false, err
	}
	if !exists {
		return model.Report{}, false, ErrReportTargetNotFound
	}

	isSelfReport, err := isSelfReportTargetTx(ctx, tx, params.ReporterID, params.TargetType, params.TargetID)
	if err != nil {
		return model.Report{}, false, err
	}
	if isSelfReport {
		return model.Report{}, false, ErrReportSelfNotAllowed
	}

	trusted, err := isTrustedReporterTx(ctx, tx, params.ReporterID)
	if err != nil {
		return model.Report{}, false, fmt.Errorf("check trusted reporter: %w", err)
	}

	report := model.Report{}
	var details sql.NullString
	if params.Details != nil {
		details.String = *params.Details
		details.Valid = true
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO reports (reporter_id, target_type, target_id, reason, details, trusted_report)
		VALUES ($1, $2::report_target, $3, $4::report_reason, $5, $6)
		RETURNING id, target_type, target_id, reason, details, status, trusted_report, created_at
	`, params.ReporterID, params.TargetType, params.TargetID, params.Reason, details, trusted).Scan(
		&report.ID,
		&report.TargetType,
		&report.TargetID,
		&report.Reason,
		&details,
		&report.Status,
		&report.TrustedReport,
		&report.CreatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" && pqErr.Constraint == "idx_reports_unique_reporter_target" {
				return model.Report{}, false, ErrReportDuplicate
			}
		}
		return model.Report{}, false, fmt.Errorf("insert report: %w", err)
	}
	if details.Valid {
		report.Details = &details.String
	}

	autoHidden := false
	if trusted {
		threshold := trustedTurnTrigger
		switch params.TargetType {
		case "debate":
			threshold = trustedDebateTrigger
		case "comment":
			threshold = trustedCommentTrigger
		}

		trustedCount, err := trustedReportCountTx(ctx, tx, params.TargetType, params.TargetID)
		if err != nil {
			return model.Report{}, false, fmt.Errorf("count trusted reports: %w", err)
		}

		if trustedCount >= threshold {
			if err := setTargetHiddenTx(ctx, tx, params.TargetType, params.TargetID, true); err != nil {
				return model.Report{}, false, err
			}
			autoHidden = true
			if err := insertModerationActionTx(ctx, tx, systemUserID, "hide_content", params.TargetType, params.TargetID, &report.ID, nil, "auto-hide threshold reached"); err != nil {
				return model.Report{}, false, fmt.Errorf("insert auto-hide action: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return model.Report{}, false, fmt.Errorf("commit report tx: %w", err)
	}

	return report, autoHidden, nil
}
