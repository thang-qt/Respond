package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"respond/internal/model"
)

var (
	ErrUserEnforcementActionNotFound    = errors.New("user enforcement action not found")
	ErrUserEnforcementActionInvalid     = errors.New("user enforcement action invalid")
	ErrUserEnforcementCapabilityInvalid = errors.New("user enforcement capability invalid")
	ErrUserEnforcementNoteRequired      = errors.New("user enforcement note required")
	ErrUserEnforcementRevokeInvalid     = errors.New("user enforcement revoke invalid")
	ErrUserSuspended                    = errors.New("user suspended")
	ErrUserBanned                       = errors.New("user banned")
	ErrUserRestricted                   = errors.New("user restricted")
)

type CreateUserEnforcementActionParams struct {
	ActorUserID  uuid.UUID
	TargetUserID uuid.UUID
	ActionType   model.UserEnforcementActionType
	Capabilities []model.UserCapability
	ExpiresAt    *time.Time
	Note         string
}

type RevokeUserEnforcementActionParams struct {
	ActorUserID  uuid.UUID
	TargetUserID uuid.UUID
	ActionID     uuid.UUID
	Note         string
}

type ListUserEnforcementActionsParams struct {
	TargetUserID uuid.UUID
	Status       string
	Page         int
	PerPage      int
}

func (s *Store) CreateUserEnforcementAction(ctx context.Context, params CreateUserEnforcementActionParams) (model.UserEnforcementAction, error) {
	note := strings.TrimSpace(params.Note)
	if note == "" || len([]rune(note)) > 500 {
		return model.UserEnforcementAction{}, ErrUserEnforcementNoteRequired
	}

	if err := validateUserEnforcementActionInput(params.ActionType, params.Capabilities, params.ExpiresAt); err != nil {
		return model.UserEnforcementAction{}, err
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.UserEnforcementAction{}, fmt.Errorf("begin user enforcement tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := ensureUserExistsTx(ctx, tx, params.TargetUserID); err != nil {
		return model.UserEnforcementAction{}, err
	}

	capabilityStrings := make([]string, 0, len(params.Capabilities))
	for _, capability := range params.Capabilities {
		capabilityStrings = append(capabilityStrings, string(capability))
	}

	action := model.UserEnforcementAction{}
	var capabilities []string
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO user_enforcement_actions (target_user_id, actor_user_id, action_type, capabilities, expires_at, note)
		VALUES ($1, $2, $3::user_enforcement_action_type, $4::user_capability[], $5, $6)
		RETURNING id, target_user_id, actor_user_id, action_type, capabilities, expires_at, revoked_at, note, created_at
	`, params.TargetUserID, params.ActorUserID, string(params.ActionType), pq.Array(capabilityStrings), params.ExpiresAt, note).Scan(
		&action.ID,
		&action.TargetUserID,
		&action.ActorUserID,
		&action.ActionType,
		pq.Array(&capabilities),
		&action.ExpiresAt,
		&action.RevokedAt,
		&action.Note,
		&action.CreatedAt,
	); err != nil {
		return model.UserEnforcementAction{}, fmt.Errorf("insert user enforcement action: %w", err)
	}
	action.Capabilities = parseUserCapabilities(capabilities)

	if params.ActionType == model.UserEnforcementActionSuspension || params.ActionType == model.UserEnforcementActionBan {
		nextStatus := model.UserAccountStatusSuspended
		if params.ActionType == model.UserEnforcementActionBan {
			nextStatus = model.UserAccountStatusBanned
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE users
			SET account_status = $2::account_status,
				updated_at = now()
			WHERE id = $1
		`, params.TargetUserID, string(nextStatus)); err != nil {
			return model.UserEnforcementAction{}, fmt.Errorf("update account status: %w", err)
		}
	}

	auditActionType := moderationAuditActionForUserEnforcement(params.ActionType)
	if err := insertModerationActionTx(
		ctx,
		tx,
		params.ActorUserID,
		auditActionType,
		"user",
		params.TargetUserID,
		nil,
		map[string]any{
			"action_id":    action.ID.String(),
			"action":       action.ActionType,
			"capabilities": capabilityStrings,
			"expires_at":   action.ExpiresAt,
		},
		note,
	); err != nil {
		return model.UserEnforcementAction{}, fmt.Errorf("insert user enforcement moderation action: %w", err)
	}

	if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
		UserID:  params.TargetUserID,
		Type:    "account_enforcement",
		Message: userEnforcementNotificationMessage(action.ActionType, action.Capabilities, action.ExpiresAt, note),
	}); err != nil {
		return model.UserEnforcementAction{}, fmt.Errorf("create account enforcement notification: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.UserEnforcementAction{}, fmt.Errorf("commit user enforcement tx: %w", err)
	}

	return action, nil
}

func (s *Store) RevokeUserEnforcementAction(ctx context.Context, params RevokeUserEnforcementActionParams) (model.UserEnforcementAction, error) {
	note := strings.TrimSpace(params.Note)
	if note == "" || len([]rune(note)) > 500 {
		return model.UserEnforcementAction{}, ErrUserEnforcementNoteRequired
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.UserEnforcementAction{}, fmt.Errorf("begin revoke user enforcement tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := ensureUserExistsTx(ctx, tx, params.TargetUserID); err != nil {
		return model.UserEnforcementAction{}, err
	}

	action := model.UserEnforcementAction{}
	var capabilities []string
	if err := tx.QueryRowContext(ctx, `
		SELECT id, target_user_id, actor_user_id, action_type, capabilities, expires_at, revoked_at, note, created_at
		FROM user_enforcement_actions
		WHERE id = $1
		  AND target_user_id = $2
	`, params.ActionID, params.TargetUserID).Scan(
		&action.ID,
		&action.TargetUserID,
		&action.ActorUserID,
		&action.ActionType,
		pq.Array(&capabilities),
		&action.ExpiresAt,
		&action.RevokedAt,
		&action.Note,
		&action.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.UserEnforcementAction{}, ErrUserEnforcementActionNotFound
		}
		return model.UserEnforcementAction{}, fmt.Errorf("load user enforcement action: %w", err)
	}
	action.Capabilities = parseUserCapabilities(capabilities)

	if action.ActionType != model.UserEnforcementActionRestriction && action.ActionType != model.UserEnforcementActionSuspension {
		return model.UserEnforcementAction{}, ErrUserEnforcementRevokeInvalid
	}
	if action.RevokedAt != nil {
		return model.UserEnforcementAction{}, ErrUserEnforcementRevokeInvalid
	}
	if action.ExpiresAt != nil && !action.ExpiresAt.After(time.Now()) {
		return model.UserEnforcementAction{}, ErrUserEnforcementRevokeInvalid
	}

	now := time.Now().UTC()
	if err := tx.QueryRowContext(ctx, `
		UPDATE user_enforcement_actions
		SET revoked_at = $2::timestamptz,
		    payload_json = payload_json || jsonb_build_object('revoke_note', $3::text, 'revoked_by', $4::text, 'revoked_at', $2::timestamptz)
		WHERE id = $1
		RETURNING revoked_at
	`, params.ActionID, now, note, params.ActorUserID.String()).Scan(&action.RevokedAt); err != nil {
		return model.UserEnforcementAction{}, fmt.Errorf("revoke user enforcement action: %w", err)
	}

	if action.ActionType == model.UserEnforcementActionSuspension {
		activeSuspension, err := hasActiveSuspensionTx(ctx, tx, params.TargetUserID)
		if err != nil {
			return model.UserEnforcementAction{}, err
		}
		if !activeSuspension {
			if _, err := tx.ExecContext(ctx, `
				UPDATE users
				SET account_status = 'active'::account_status,
					updated_at = now()
				WHERE id = $1
				  AND account_status = 'suspended'::account_status
			`, params.TargetUserID); err != nil {
				return model.UserEnforcementAction{}, fmt.Errorf("clear suspended account status: %w", err)
			}
		}
	}

	if err := insertModerationActionTx(
		ctx,
		tx,
		params.ActorUserID,
		"revoke_user_enforcement",
		"user",
		params.TargetUserID,
		nil,
		map[string]string{
			"action_id":           params.ActionID.String(),
			"revoked_action_type": string(action.ActionType),
		},
		note,
	); err != nil {
		return model.UserEnforcementAction{}, fmt.Errorf("insert revoke moderation action: %w", err)
	}

	if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
		UserID:  params.TargetUserID,
		Type:    "account_enforcement_revoked",
		Message: fmt.Sprintf("A moderator revoked your %s action. Moderator note: %s", action.ActionType, note),
	}); err != nil {
		return model.UserEnforcementAction{}, fmt.Errorf("create revoke notification: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.UserEnforcementAction{}, fmt.Errorf("commit revoke user enforcement tx: %w", err)
	}

	return action, nil
}

func (s *Store) ListUserEnforcementActions(ctx context.Context, params ListUserEnforcementActionsParams) ([]model.UserEnforcementAction, int, error) {
	pagination := normalizePagination(params.Page, params.PerPage, 20, 50)
	perPage := pagination.PerPage

	status := strings.TrimSpace(strings.ToLower(params.Status))
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "expired" && status != "revoked" && status != "all" {
		return nil, 0, ErrUserEnforcementActionInvalid
	}

	where := `WHERE uea.target_user_id = $1`
	args := []any{params.TargetUserID}
	if status == "active" {
		where += ` AND uea.revoked_at IS NULL AND (uea.expires_at IS NULL OR uea.expires_at > now())`
	} else if status == "expired" {
		where += ` AND uea.revoked_at IS NULL AND uea.expires_at IS NOT NULL AND uea.expires_at <= now()`
	} else if status == "revoked" {
		where += ` AND uea.revoked_at IS NOT NULL`
	}

	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM user_enforcement_actions uea %s`, where)
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count user enforcement actions: %w", err)
	}

	args = append(args, perPage, pagination.Offset)
	query := fmt.Sprintf(`
		SELECT
			uea.id,
			uea.target_user_id,
			uea.actor_user_id,
			uea.action_type,
			uea.capabilities,
			uea.expires_at,
			uea.revoked_at,
			uea.note,
			uea.created_at,
			u.username
		FROM user_enforcement_actions uea
		JOIN users u ON u.id = uea.actor_user_id
		%s
		ORDER BY uea.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)-1, len(args))

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list user enforcement actions: %w", err)
	}
	defer rows.Close()

	items := make([]model.UserEnforcementAction, 0)
	now := time.Now().UTC()
	for rows.Next() {
		item := model.UserEnforcementAction{}
		var capabilityValues []string
		var actorUsername string
		if err := rows.Scan(
			&item.ID,
			&item.TargetUserID,
			&item.ActorUserID,
			&item.ActionType,
			pq.Array(&capabilityValues),
			&item.ExpiresAt,
			&item.RevokedAt,
			&item.Note,
			&item.CreatedAt,
			&actorUsername,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user enforcement action: %w", err)
		}
		item.Capabilities = parseUserCapabilities(capabilityValues)
		item.CreatedBy = &model.ReportUserRef{ID: item.ActorUserID, Username: actorUsername}
		item.Status = userEnforcementStatus(item, now)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate user enforcement actions: %w", err)
	}

	return items, total, nil
}
