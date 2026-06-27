package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

func moderationNotificationTargetTx(ctx context.Context, tx *sql.Tx, targetType string, targetID uuid.UUID) (moderationNotifTarget, error) {
	var out moderationNotifTarget

	switch targetType {
	case "debate":
		var sideAUserID string
		var sideBUserID sql.NullString
		var debateTopic sql.NullString
		err := tx.QueryRowContext(ctx, `
			SELECT d.side_a_user_id, d.side_b_user_id, d.topic
			FROM debates d
			WHERE d.id = $1
		`, targetID).Scan(&sideAUserID, &sideBUserID, &debateTopic)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return moderationNotifTarget{}, ErrReportTargetNotFound
			}
			return moderationNotifTarget{}, fmt.Errorf("load debate notification target: %w", err)
		}
		parsedSideAUserID, err := uuid.Parse(sideAUserID)
		if err != nil {
			return moderationNotifTarget{}, fmt.Errorf("parse debate side a user id: %w", err)
		}
		out.UserIDs = append(out.UserIDs, parsedSideAUserID)
		if sideBUserID.Valid {
			parsedSideBUserID, err := uuid.Parse(sideBUserID.String)
			if err != nil {
				return moderationNotifTarget{}, fmt.Errorf("parse debate side b user id: %w", err)
			}
			if parsedSideBUserID != parsedSideAUserID {
				out.UserIDs = append(out.UserIDs, parsedSideBUserID)
			}
		}
		out.DebateID = &targetID
		if debateTopic.Valid {
			topic := debateTopic.String
			out.DebateTopic = &topic
		}
	case "turn":
		var userID string
		var debateID sql.NullString
		var debateTopic sql.NullString
		var turnNumber sql.NullInt64
		err := tx.QueryRowContext(ctx, `
			SELECT t.user_id, t.debate_id, t.turn_number, d.topic
			FROM turns t
			LEFT JOIN debates d ON d.id = t.debate_id
			WHERE t.id = $1
		`, targetID).Scan(&userID, &debateID, &turnNumber, &debateTopic)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return moderationNotifTarget{}, ErrReportTargetNotFound
			}
			return moderationNotifTarget{}, fmt.Errorf("load turn notification target: %w", err)
		}
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			return moderationNotifTarget{}, fmt.Errorf("parse turn owner id: %w", err)
		}
		out.UserIDs = []uuid.UUID{parsedUserID}
		if debateID.Valid {
			parsedDebateID, err := uuid.Parse(debateID.String)
			if err != nil {
				return moderationNotifTarget{}, fmt.Errorf("parse turn debate id: %w", err)
			}
			out.DebateID = &parsedDebateID
		}
		if turnNumber.Valid {
			tn := int(turnNumber.Int64)
			out.TurnNumber = &tn
		}
		if debateTopic.Valid {
			topic := debateTopic.String
			out.DebateTopic = &topic
		}
	case "comment":
		var userID string
		var debateID sql.NullString
		var debateTopic sql.NullString
		err := tx.QueryRowContext(ctx, `
			SELECT c.user_id, c.debate_id, d.topic
			FROM comments c
			LEFT JOIN debates d ON d.id = c.debate_id
			WHERE c.id = $1
		`, targetID).Scan(&userID, &debateID, &debateTopic)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return moderationNotifTarget{}, ErrReportTargetNotFound
			}
			return moderationNotifTarget{}, fmt.Errorf("load comment notification target: %w", err)
		}
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			return moderationNotifTarget{}, fmt.Errorf("parse comment owner id: %w", err)
		}
		out.UserIDs = []uuid.UUID{parsedUserID}
		if debateID.Valid {
			parsedDebateID, err := uuid.Parse(debateID.String)
			if err != nil {
				return moderationNotifTarget{}, fmt.Errorf("parse comment debate id: %w", err)
			}
			out.DebateID = &parsedDebateID
		}
		if debateTopic.Valid {
			topic := debateTopic.String
			out.DebateTopic = &topic
		}
	default:
		return moderationNotifTarget{}, fmt.Errorf("invalid target type")
	}

	return out, nil
}

func isReportRateLimited(ctx context.Context, tx *sql.Tx, reporterID uuid.UUID) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM reports
		WHERE reporter_id = $1
			AND created_at >= now() - interval '24 hours'
	`, reporterID).Scan(&count); err != nil {
		return false, err
	}
	return count >= reportLimitPer24h, nil
}

func isTrustedReporterTx(ctx context.Context, tx *sql.Tx, reporterID uuid.UUID) (bool, error) {
	var trusted bool
	err := tx.QueryRowContext(ctx, `
		SELECT email_verified
			AND created_at <= (now() - interval '180 days')
			AND rating BETWEEN $2 AND $3
		FROM users
		WHERE id = $1
	`, reporterID, trustedMinRating, trustedMaxRating).Scan(&trusted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrNotFound
		}
		return false, err
	}
	return trusted, nil
}

func trustedReportCountTx(ctx context.Context, tx *sql.Tx, targetType string, targetID uuid.UUID) (int, error) {
	var count int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT reporter_id)
		FROM reports
		WHERE target_type = $1::report_target
			AND target_id = $2
			AND trusted_report = true
			AND created_at >= now() - interval '24 hours'
	`, targetType, targetID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func targetExistsTx(ctx context.Context, tx *sql.Tx, targetType string, targetID uuid.UUID) (bool, error) {
	spec, ok := reportTargetSpecFor(targetType)
	if !ok {
		return false, fmt.Errorf("invalid target type")
	}

	var exists bool
	if err := tx.QueryRowContext(ctx, fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)`, spec.table), targetID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check report target exists: %w", err)
	}
	return exists, nil
}

func isSelfReportTargetTx(ctx context.Context, tx *sql.Tx, reporterID uuid.UUID, targetType string, targetID uuid.UUID) (bool, error) {
	var query string
	switch targetType {
	case "debate":
		query = `
			SELECT EXISTS(
				SELECT 1
				FROM debates
				WHERE id = $1
					AND ($2 = side_a_user_id OR $2 = side_b_user_id)
			)
		`
	case "turn":
		query = `
			SELECT EXISTS(
				SELECT 1
				FROM turns
				WHERE id = $1
					AND user_id = $2
			)
		`
	case "comment":
		query = `
			SELECT EXISTS(
				SELECT 1
				FROM comments
				WHERE id = $1
					AND user_id = $2
			)
		`
	default:
		return false, fmt.Errorf("invalid target type")
	}

	var isSelf bool
	if err := tx.QueryRowContext(ctx, query, targetID, reporterID).Scan(&isSelf); err != nil {
		return false, fmt.Errorf("check self report target: %w", err)
	}

	return isSelf, nil
}

func setTargetHiddenTx(ctx context.Context, tx *sql.Tx, targetType string, targetID uuid.UUID, hidden bool) error {
	spec, ok := reportTargetSpecFor(targetType)
	if !ok {
		return fmt.Errorf("invalid target type")
	}

	res, err := tx.ExecContext(ctx, fmt.Sprintf(`UPDATE %s SET hidden = $2, moderation_pending = $2 WHERE id = $1`, spec.table), targetID, hidden)
	if err != nil {
		return fmt.Errorf("update target hidden: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update target hidden rows: %w", err)
	}
	if rows == 0 {
		return ErrReportTargetNotFound
	}
	return nil
}

func insertModerationActionTx(
	ctx context.Context,
	tx *sql.Tx,
	actorID uuid.UUID,
	actionType string,
	targetType string,
	targetID uuid.UUID,
	reportID *uuid.UUID,
	payload any,
	reason string,
) error {
	var payloadJSON []byte
	if payload == nil {
		payloadJSON = []byte("{}")
	} else {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal moderation payload: %w", err)
		}
		payloadJSON = encoded
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO moderation_actions (actor_user_id, action_type, target_type, target_id, report_id, payload_json, reason)
		VALUES ($1, $2::moderation_action_type, $3, $4, $5, $6::jsonb, NULLIF($7, ''))
	`, actorID, actionType, targetType, targetID, reportID, string(payloadJSON), reason); err != nil {
		return err
	}

	return nil
}

func getReportTargetContent(ctx context.Context, db *sql.DB, targetType string, targetID uuid.UUID) (model.ReportTarget, error) {
	spec, ok := reportTargetSpecFor(targetType)
	if !ok {
		return model.ReportTarget{}, fmt.Errorf("invalid target type")
	}

	var target model.ReportTarget
	err := db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT hidden, %s
		FROM %s
		WHERE id = $1
	`, spec.contentColumn, spec.table), targetID).Scan(&target.Hidden, &target.Content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ReportTarget{}, ErrReportTargetNotFound
		}
		return model.ReportTarget{}, fmt.Errorf("get report target content: %w", err)
	}

	return target, nil
}
