package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Store) ResolveReport(ctx context.Context, params ResolveReportParams) (ResolveReportParams, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return ResolveReportParams{}, fmt.Errorf("begin resolve tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var status string
	err = tx.QueryRowContext(ctx, `
		SELECT target_type, target_id, status
		FROM reports
		WHERE id = $1
		FOR UPDATE
	`, params.ReportID).Scan(&params.TargetType, &params.TargetID, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ResolveReportParams{}, ErrReportNotFound
		}
		return ResolveReportParams{}, fmt.Errorf("load report for resolve: %w", err)
	}
	if status != "open" {
		return ResolveReportParams{}, ErrReportAlreadyClosed
	}

	params.ReviewedAt = time.Now().UTC()
	if params.Note != nil {
		note := strings.TrimSpace(*params.Note)
		if note == "" {
			params.Note = nil
		} else {
			params.Note = &note
		}
	}
	if params.Note != nil && len([]rune(*params.Note)) > 500 {
		return ResolveReportParams{}, fmt.Errorf("invalid note length")
	}
	if (params.Resolution == "hide" || params.Resolution == "restore") && params.Note == nil {
		return ResolveReportParams{}, fmt.Errorf("missing note for resolution")
	}

	var actionType string
	switch params.Resolution {
	case "dismiss":
		actionType = "dismiss_report"
		params.FinalStatus = "dismissed"
	case "hide":
		actionType = "hide_content"
		params.FinalStatus = "actioned"
		if err := setTargetHiddenTx(ctx, tx, params.TargetType, params.TargetID, true); err != nil {
			return ResolveReportParams{}, err
		}
	case "restore":
		actionType = "restore_content"
		params.FinalStatus = "actioned"
		if err := setTargetHiddenTx(ctx, tx, params.TargetType, params.TargetID, false); err != nil {
			return ResolveReportParams{}, err
		}
	default:
		return ResolveReportParams{}, fmt.Errorf("invalid resolution")
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE reports
		SET status = $2::report_status,
			resolution = $3,
			reviewed_by_user_id = $4,
			reviewed_at = $5,
			resolution_note = $6
		WHERE id = $1
	`, params.ReportID, params.FinalStatus, params.Resolution, params.ReviewerID, params.ReviewedAt, params.Note); err != nil {
		return ResolveReportParams{}, fmt.Errorf("update report resolution: %w", err)
	}

	reason := ""
	if params.Note != nil {
		reason = *params.Note
	}
	if err := insertModerationActionTx(ctx, tx, params.ReviewerID, actionType, params.TargetType, params.TargetID, &params.ReportID, map[string]string{"resolution": params.Resolution}, reason); err != nil {
		return ResolveReportParams{}, fmt.Errorf("insert moderation action: %w", err)
	}

	if params.Resolution == "hide" || params.Resolution == "restore" {
		notif, err := moderationNotificationTargetTx(ctx, tx, params.TargetType, params.TargetID)
		if err != nil {
			return ResolveReportParams{}, err
		}
		notifType := "content_restored"
		if params.Resolution == "hide" {
			notifType = "content_hidden"
		}
		message := moderationNotificationMessage(params.Resolution, params.TargetType, notif.TurnNumber, notif.DebateTopic, params.Note)
		for _, userID := range notif.UserIDs {
			if userID == uuid.Nil || userID == params.ReviewerID {
				continue
			}
			if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
				UserID:     userID,
				Type:       notifType,
				Message:    message,
				DebateID:   notif.DebateID,
				TurnNumber: notif.TurnNumber,
			}); err != nil {
				return ResolveReportParams{}, fmt.Errorf("create moderation notification: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return ResolveReportParams{}, fmt.Errorf("commit resolve report tx: %w", err)
	}

	return params, nil
}
