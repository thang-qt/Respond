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
	ErrInviteNotFound      = errors.New("invite not found")
	ErrInviteExpired       = errors.New("invite expired")
	ErrInviteAlreadyUsed   = errors.New("invite already used")
	ErrInviteEmailMismatch = errors.New("invite email mismatch")
	ErrInviteDuplicate     = errors.New("invite duplicate")
	ErrInviteRevokeInvalid = errors.New("invite revoke invalid")
)

func (s *Store) GetInviteByTokenHash(ctx context.Context, tokenHash string) (model.Invite, error) {
	const query = `
		SELECT id, inviter_user_id, invited_email, token_hash, status, expires_at,
			accepted_by_user_id, accepted_at, revoked_at, created_at, updated_at
		FROM invites
		WHERE token_hash = $1
	`

	var invite model.Invite
	err := s.DB.QueryRowContext(ctx, query, tokenHash).Scan(
		&invite.ID,
		&invite.InviterUserID,
		&invite.InvitedEmail,
		&invite.TokenHash,
		&invite.Status,
		&invite.ExpiresAt,
		&invite.AcceptedByUserID,
		&invite.AcceptedAt,
		&invite.RevokedAt,
		&invite.CreatedAt,
		&invite.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Invite{}, ErrInviteNotFound
		}
		return model.Invite{}, fmt.Errorf("get invite by token hash: %w", err)
	}

	return invite, nil
}

type CreateInviteParams struct {
	InviterUserID uuid.UUID
	InvitedEmail  string
	TokenHash     string
	ExpiresAt     time.Time
}

func (s *Store) CreateInvite(ctx context.Context, params CreateInviteParams) (model.Invite, error) {
	if _, err := s.DB.ExecContext(ctx, `
		UPDATE invites
		SET status = 'expired', updated_at = now()
		WHERE LOWER(invited_email) = LOWER($1)
		  AND status = 'pending'
		  AND expires_at <= now()
	`, params.InvitedEmail); err != nil {
		return model.Invite{}, fmt.Errorf("expire stale invites for email: %w", err)
	}

	const query = `
		INSERT INTO invites (inviter_user_id, invited_email, token_hash, expires_at)
		VALUES ($1, LOWER($2), $3, $4)
		RETURNING id, inviter_user_id, invited_email, token_hash, status, expires_at,
			accepted_by_user_id, accepted_at, revoked_at, created_at, updated_at
	`

	var invite model.Invite
	err := s.DB.QueryRowContext(ctx, query, params.InviterUserID, params.InvitedEmail, params.TokenHash, params.ExpiresAt).Scan(
		&invite.ID,
		&invite.InviterUserID,
		&invite.InvitedEmail,
		&invite.TokenHash,
		&invite.Status,
		&invite.ExpiresAt,
		&invite.AcceptedByUserID,
		&invite.AcceptedAt,
		&invite.RevokedAt,
		&invite.CreatedAt,
		&invite.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "idx_invites_pending_email_unique" {
			return model.Invite{}, ErrInviteDuplicate
		}
		return model.Invite{}, fmt.Errorf("create invite: %w", err)
	}

	return invite, nil
}

type ListInvitesByIssuerParams struct {
	IssuerUserID uuid.UUID
	Status       string
	Page         int
	PerPage      int
}

func (s *Store) ListInvitesByIssuer(ctx context.Context, params ListInvitesByIssuerParams) ([]model.Invite, int, error) {
	pagination := normalizePagination(params.Page, params.PerPage, 20, 50)
	perPage := pagination.PerPage

	status := strings.ToLower(strings.TrimSpace(params.Status))
	if status == "" {
		status = string(model.InviteStatusPending)
	}

	where := `inviter_user_id = $1`
	args := []any{params.IssuerUserID}
	if status != "all" {
		if status == string(model.InviteStatusExpired) {
			where += ` AND status = 'pending' AND expires_at <= now()`
		} else {
			args = append(args, status)
			where += fmt.Sprintf(` AND (
				(status = $%d AND NOT (status = 'pending' AND expires_at <= now()))
				OR ($%d = 'pending' AND status = 'pending' AND expires_at > now())
			)`, len(args), len(args))
		}
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM invites WHERE %s`, where)
	var total int
	if err := s.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count invites by issuer: %w", err)
	}

	args = append(args, perPage, pagination.Offset)
	query := fmt.Sprintf(`
		SELECT id, inviter_user_id, invited_email, token_hash,
			CASE
				WHEN status = 'pending' AND expires_at <= now() THEN 'expired'::invite_status
				ELSE status
			END AS status,
			expires_at, accepted_by_user_id, accepted_at, revoked_at, created_at, updated_at
		FROM invites
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)-1, len(args))

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list invites by issuer: %w", err)
	}
	defer rows.Close()

	invites := make([]model.Invite, 0, perPage)
	for rows.Next() {
		var invite model.Invite
		if err := rows.Scan(
			&invite.ID,
			&invite.InviterUserID,
			&invite.InvitedEmail,
			&invite.TokenHash,
			&invite.Status,
			&invite.ExpiresAt,
			&invite.AcceptedByUserID,
			&invite.AcceptedAt,
			&invite.RevokedAt,
			&invite.CreatedAt,
			&invite.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan invite by issuer: %w", err)
		}
		invites = append(invites, invite)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate invites by issuer: %w", err)
	}

	return invites, total, nil
}

func (s *Store) RevokeInviteByID(ctx context.Context, issuerUserID, inviteID uuid.UUID) (model.Invite, error) {
	const query = `
		UPDATE invites
		SET status = 'revoked',
			revoked_at = now(),
			updated_at = now()
		WHERE id = $1
		  AND inviter_user_id = $2
		  AND status = 'pending'
		  AND expires_at > now()
		RETURNING id, inviter_user_id, invited_email, token_hash, status, expires_at,
			accepted_by_user_id, accepted_at, revoked_at, created_at, updated_at
	`

	var invite model.Invite
	err := s.DB.QueryRowContext(ctx, query, inviteID, issuerUserID).Scan(
		&invite.ID,
		&invite.InviterUserID,
		&invite.InvitedEmail,
		&invite.TokenHash,
		&invite.Status,
		&invite.ExpiresAt,
		&invite.AcceptedByUserID,
		&invite.AcceptedAt,
		&invite.RevokedAt,
		&invite.CreatedAt,
		&invite.UpdatedAt,
	)
	if err == nil {
		return invite, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return model.Invite{}, fmt.Errorf("revoke invite by id: %w", err)
	}

	existing, lookupErr := s.GetInviteByIDForIssuer(ctx, issuerUserID, inviteID)
	if lookupErr != nil {
		if errors.Is(lookupErr, ErrInviteNotFound) {
			return model.Invite{}, ErrInviteNotFound
		}
		return model.Invite{}, fmt.Errorf("lookup invite for revoke: %w", lookupErr)
	}
	if existing.Status == model.InviteStatusPending && time.Now().After(existing.ExpiresAt) {
		return model.Invite{}, ErrInviteRevokeInvalid
	}
	if existing.Status != model.InviteStatusPending {
		return model.Invite{}, ErrInviteRevokeInvalid
	}

	return model.Invite{}, ErrInviteRevokeInvalid
}

func (s *Store) GetInviteByIDForIssuer(ctx context.Context, issuerUserID, inviteID uuid.UUID) (model.Invite, error) {
	const query = `
		SELECT id, inviter_user_id, invited_email, token_hash, status, expires_at,
			accepted_by_user_id, accepted_at, revoked_at, created_at, updated_at
		FROM invites
		WHERE id = $1 AND inviter_user_id = $2
	`

	var invite model.Invite
	err := s.DB.QueryRowContext(ctx, query, inviteID, issuerUserID).Scan(
		&invite.ID,
		&invite.InviterUserID,
		&invite.InvitedEmail,
		&invite.TokenHash,
		&invite.Status,
		&invite.ExpiresAt,
		&invite.AcceptedByUserID,
		&invite.AcceptedAt,
		&invite.RevokedAt,
		&invite.CreatedAt,
		&invite.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Invite{}, ErrInviteNotFound
		}
		return model.Invite{}, fmt.Errorf("get invite by id for issuer: %w", err)
	}

	if invite.Status == model.InviteStatusPending && time.Now().After(invite.ExpiresAt) {
		invite.Status = model.InviteStatusExpired
	}

	return invite, nil
}

type InviteLineageItem struct {
	Depth       int
	UserID      uuid.UUID
	Username    string
	InvitedByID *uuid.UUID
}

func (s *Store) GetInviteLineage(ctx context.Context, userID uuid.UUID) ([]InviteLineageItem, error) {
	const query = `
		WITH RECURSIVE lineage AS (
			SELECT 0 AS depth, id, username, invited_by_user_id
			FROM users
			WHERE id = $1
			UNION ALL
			SELECT l.depth + 1, u.id, u.username, u.invited_by_user_id
			FROM lineage l
			JOIN users u ON u.id = l.invited_by_user_id
		)
		SELECT depth, id, username, invited_by_user_id
		FROM lineage
		ORDER BY depth ASC
	`

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get invite lineage: %w", err)
	}
	defer rows.Close()

	items := make([]InviteLineageItem, 0, 4)
	for rows.Next() {
		var item InviteLineageItem
		if err := rows.Scan(&item.Depth, &item.UserID, &item.Username, &item.InvitedByID); err != nil {
			return nil, fmt.Errorf("scan invite lineage: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate invite lineage: %w", err)
	}

	if len(items) == 0 {
		return nil, ErrNotFound
	}

	return items, nil
}

type CreateUserWithInviteParams struct {
	Email        string
	Username     string
	PasswordHash string
	TokenHash    string
}

func (s *Store) CreateUserWithSettingsFromInvite(ctx context.Context, params CreateUserWithInviteParams) (model.User, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.User{}, fmt.Errorf("begin user invite tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	invite, err := lockPendingInviteByTokenHash(ctx, tx, params.TokenHash)
	if err != nil {
		if errors.Is(err, ErrInviteNotFound) || errors.Is(err, ErrInviteAlreadyUsed) {
			return model.User{}, err
		}
		return model.User{}, fmt.Errorf("lock invite: %w", err)
	}

	if time.Now().After(invite.ExpiresAt) {
		if _, updateErr := tx.ExecContext(ctx, `
			UPDATE invites
			SET status = 'expired', updated_at = now()
			WHERE id = $1 AND status = 'pending'
		`, invite.ID); updateErr != nil {
			return model.User{}, fmt.Errorf("mark invite expired: %w", updateErr)
		}
		return model.User{}, ErrInviteExpired
	}

	if !strings.EqualFold(strings.TrimSpace(params.Email), invite.InvitedEmail) {
		return model.User{}, ErrInviteEmailMismatch
	}

	const insertUserQuery = `
		INSERT INTO users (email, username, password_hash, invited_by_user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, email_verified, role, account_status, username, password_hash, bio, rating,
			wins, losses, draws, default_reveal, locale, username_changed_at, created_at, updated_at
	`

	var user model.User
	err = tx.QueryRowContext(ctx, insertUserQuery, params.Email, params.Username, params.PasswordHash, invite.InviterUserID).Scan(
		&user.ID,
		&user.Email,
		&user.EmailVerified,
		&user.Role,
		&user.AccountStatus,
		&user.Username,
		&user.PasswordHash,
		&user.Bio,
		&user.Rating,
		&user.Wins,
		&user.Losses,
		&user.Draws,
		&user.DefaultReveal,
		&user.Locale,
		&user.UsernameChangedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("create user from invite: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO notification_settings (user_id)
		VALUES ($1)
	`, user.ID); err != nil {
		return model.User{}, fmt.Errorf("create notification settings from invite: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE invites
		SET status = 'accepted',
			accepted_by_user_id = $2,
			accepted_at = now(),
			updated_at = now()
		WHERE id = $1
		  AND status = 'pending'
	`, invite.ID, user.ID)
	if err != nil {
		return model.User{}, fmt.Errorf("consume invite: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.User{}, fmt.Errorf("consume invite rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return model.User{}, ErrInviteAlreadyUsed
	}

	if err := tx.Commit(); err != nil {
		return model.User{}, fmt.Errorf("commit user invite signup: %w", err)
	}

	return user, nil
}

func lockPendingInviteByTokenHash(ctx context.Context, tx *sql.Tx, tokenHash string) (model.Invite, error) {
	const query = `
		SELECT id, inviter_user_id, invited_email, token_hash, status, expires_at,
			accepted_by_user_id, accepted_at, revoked_at, created_at, updated_at
		FROM invites
		WHERE token_hash = $1
		FOR UPDATE
	`

	var invite model.Invite
	err := tx.QueryRowContext(ctx, query, tokenHash).Scan(
		&invite.ID,
		&invite.InviterUserID,
		&invite.InvitedEmail,
		&invite.TokenHash,
		&invite.Status,
		&invite.ExpiresAt,
		&invite.AcceptedByUserID,
		&invite.AcceptedAt,
		&invite.RevokedAt,
		&invite.CreatedAt,
		&invite.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Invite{}, ErrInviteNotFound
		}
		return model.Invite{}, err
	}

	if invite.Status != model.InviteStatusPending {
		return model.Invite{}, ErrInviteAlreadyUsed
	}

	return invite, nil
}
