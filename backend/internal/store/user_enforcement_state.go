package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/model"
)

func (s *Store) EnsureUserCanPerform(ctx context.Context, userID uuid.UUID, capability *model.UserCapability) error {
	state, err := s.GetUserEnforcementState(ctx, userID)
	if err != nil {
		return err
	}

	if state.AccountStatus == model.UserAccountStatusBanned {
		return ErrUserBanned
	}
	if state.AccountStatus == model.UserAccountStatusSuspended {
		return ErrUserSuspended
	}

	if capability != nil && state.RestrictedCapabilities[*capability] {
		return ErrUserRestricted
	}

	return nil
}

func (s *Store) GetUserEnforcementState(ctx context.Context, userID uuid.UUID) (model.UserEnforcementState, error) {
	state := model.UserEnforcementState{
		RestrictedCapabilities: map[model.UserCapability]bool{},
	}

	status, err := s.getUserAccountStatus(ctx, userID)
	if err != nil {
		return model.UserEnforcementState{}, err
	}

	if status == model.UserAccountStatusSuspended {
		hasSuspension, err := s.hasActiveSuspension(ctx, userID)
		if err != nil {
			return model.UserEnforcementState{}, err
		}
		if !hasSuspension {
			if err := s.UpdateUserAccountStatus(ctx, userID, model.UserAccountStatusActive); err != nil {
				return model.UserEnforcementState{}, err
			}
			status = model.UserAccountStatusActive
		}
	}

	state.AccountStatus = status

	rows, err := s.DB.QueryContext(ctx, `
		SELECT DISTINCT unnest(uea.capabilities)::text
		FROM user_enforcement_actions uea
		WHERE uea.target_user_id = $1
		  AND uea.action_type = 'restriction'::user_enforcement_action_type
		  AND uea.revoked_at IS NULL
		  AND (uea.expires_at IS NULL OR uea.expires_at > now())
	`, userID)
	if err != nil {
		return model.UserEnforcementState{}, fmt.Errorf("list restricted capabilities: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var capability string
		if err := rows.Scan(&capability); err != nil {
			return model.UserEnforcementState{}, fmt.Errorf("scan restricted capability: %w", err)
		}
		state.RestrictedCapabilities[model.UserCapability(capability)] = true
	}
	if err := rows.Err(); err != nil {
		return model.UserEnforcementState{}, fmt.Errorf("iterate restricted capabilities: %w", err)
	}

	return state, nil
}

func (s *Store) GetCurrentAccountBlockReason(ctx context.Context, userID uuid.UUID) (model.AccountBlockReason, error) {
	status, err := s.getUserAccountStatus(ctx, userID)
	if err != nil {
		return model.AccountBlockReason{}, err
	}

	if status != model.UserAccountStatusSuspended && status != model.UserAccountStatusBanned {
		return model.AccountBlockReason{}, ErrUserEnforcementActionNotFound
	}

	reason := model.AccountBlockReason{}
	query := `
		SELECT action_type, expires_at, note
		FROM user_enforcement_actions
		WHERE target_user_id = $1
		  AND revoked_at IS NULL
	`
	if status == model.UserAccountStatusBanned {
		query += ` AND action_type = 'ban'::user_enforcement_action_type`
	} else {
		query += `
		  AND action_type = 'suspension'::user_enforcement_action_type
		  AND (expires_at IS NULL OR expires_at > now())
		`
	}
	query += ` ORDER BY created_at DESC LIMIT 1`

	if err := s.DB.QueryRowContext(ctx, query, userID).Scan(&reason.ActionType, &reason.ExpiresAt, &reason.Note); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if status == model.UserAccountStatusSuspended {
				if err := s.UpdateUserAccountStatus(ctx, userID, model.UserAccountStatusActive); err != nil {
					return model.AccountBlockReason{}, err
				}
			}
			return model.AccountBlockReason{}, ErrUserEnforcementActionNotFound
		}
		return model.AccountBlockReason{}, fmt.Errorf("get current account block reason: %w", err)
	}

	return reason, nil
}

func (s *Store) UpdateUserAccountStatus(ctx context.Context, userID uuid.UUID, status model.UserAccountStatus) error {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE users
		SET account_status = $2::account_status,
			updated_at = now()
		WHERE id = $1
	`, userID, string(status))
	if err != nil {
		return fmt.Errorf("update user account status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user account status rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) getUserAccountStatus(ctx context.Context, userID uuid.UUID) (model.UserAccountStatus, error) {
	var status model.UserAccountStatus
	if err := s.DB.QueryRowContext(ctx, `
		SELECT account_status
		FROM users
		WHERE id = $1
	`, userID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("get user account status: %w", err)
	}
	return status, nil
}

func (s *Store) hasActiveSuspension(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	if err := s.DB.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM user_enforcement_actions
			WHERE target_user_id = $1
			  AND action_type = 'suspension'::user_enforcement_action_type
			  AND revoked_at IS NULL
			  AND (expires_at IS NULL OR expires_at > now())
		)
	`, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check active suspension: %w", err)
	}
	return exists, nil
}

func hasActiveSuspensionTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (bool, error) {
	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM user_enforcement_actions
			WHERE target_user_id = $1
			  AND action_type = 'suspension'::user_enforcement_action_type
			  AND revoked_at IS NULL
			  AND (expires_at IS NULL OR expires_at > now())
		)
	`, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check active suspension in tx: %w", err)
	}
	return exists, nil
}

func ensureUserExistsTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (bool, error) {
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}
	if !exists {
		return false, ErrNotFound
	}
	return true, nil
}
