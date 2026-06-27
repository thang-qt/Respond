package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type EmailVerificationToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Email     string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (s *Store) CreateRefreshToken(ctx context.Context, params CreateRefreshTokenParams) (RefreshToken, error) {
	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, created_at
	`

	var token RefreshToken
	err := s.DB.QueryRowContext(ctx, query, params.UserID, params.TokenHash, params.ExpiresAt).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("create refresh token: %w", err)
	}

	return token, nil
}

func (s *Store) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshToken, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var token RefreshToken
	err := s.DB.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RefreshToken{}, ErrNotFound
		}
		return RefreshToken{}, fmt.Errorf("get refresh token: %w", err)
	}

	return token, nil
}

func (s *Store) DeleteRefreshTokenByHash(ctx context.Context, tokenHash string) error {
	const query = `
		DELETE FROM refresh_tokens
		WHERE token_hash = $1
	`

	_, err := s.DB.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

func (s *Store) DeleteRefreshTokensByUserID(ctx context.Context, userID uuid.UUID) error {
	const query = `
		DELETE FROM refresh_tokens
		WHERE user_id = $1
	`

	_, err := s.DB.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete refresh tokens by user id: %w", err)
	}
	return nil
}

func (s *Store) CreatePasswordResetToken(ctx context.Context, params CreatePasswordResetTokenParams) (PasswordResetToken, error) {
	const query = `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, used_at, created_at
	`

	var token PasswordResetToken
	err := s.DB.QueryRowContext(ctx, query, params.UserID, params.TokenHash, params.ExpiresAt).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if err != nil {
		return PasswordResetToken{}, fmt.Errorf("create password reset token: %w", err)
	}

	return token, nil
}

func (s *Store) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (PasswordResetToken, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`

	var token PasswordResetToken
	err := s.DB.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PasswordResetToken{}, ErrNotFound
		}
		return PasswordResetToken{}, fmt.Errorf("get password reset token: %w", err)
	}

	return token, nil
}

func (s *Store) MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE password_reset_tokens
		SET used_at = now()
		WHERE id = $1
	`

	res, err := s.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark password reset token used: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark password reset token used rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) CreateEmailVerificationToken(ctx context.Context, params CreateEmailVerificationTokenParams) (EmailVerificationToken, error) {
	const query = `
		INSERT INTO email_verification_tokens (user_id, email, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, email, token_hash, expires_at, used_at, created_at
	`

	var token EmailVerificationToken
	err := s.DB.QueryRowContext(ctx, query, params.UserID, params.Email, params.TokenHash, params.ExpiresAt).Scan(
		&token.ID,
		&token.UserID,
		&token.Email,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if err != nil {
		return EmailVerificationToken{}, fmt.Errorf("create email verification token: %w", err)
	}

	return token, nil
}

func (s *Store) GetEmailVerificationTokenByHash(ctx context.Context, tokenHash string) (EmailVerificationToken, error) {
	const query = `
		SELECT id, user_id, email, token_hash, expires_at, used_at, created_at
		FROM email_verification_tokens
		WHERE token_hash = $1
	`

	var token EmailVerificationToken
	err := s.DB.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.Email,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmailVerificationToken{}, ErrNotFound
		}
		return EmailVerificationToken{}, fmt.Errorf("get email verification token: %w", err)
	}

	return token, nil
}

func (s *Store) MarkEmailVerificationTokenUsed(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE email_verification_tokens
		SET used_at = now()
		WHERE id = $1
	`

	res, err := s.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark email verification token used: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark email verification token used rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) DeleteEmailVerificationTokensByUserID(ctx context.Context, userID uuid.UUID) error {
	const query = `
		DELETE FROM email_verification_tokens
		WHERE user_id = $1
	`

	_, err := s.DB.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete email verification tokens: %w", err)
	}
	return nil
}

func (s *Store) CountEmailVerificationTokensByUserSince(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM email_verification_tokens
		WHERE user_id = $1
		  AND created_at >= $2
	`

	var count int
	if err := s.DB.QueryRowContext(ctx, query, userID, since).Scan(&count); err != nil {
		return 0, fmt.Errorf("count email verification tokens: %w", err)
	}
	return count, nil
}

// DeleteExpiredTokens removes expired rows from all three token tables.
func (s *Store) DeleteExpiredTokens(ctx context.Context) error {
	queries := []string{
		`DELETE FROM refresh_tokens WHERE expires_at < now()`,
		`DELETE FROM password_reset_tokens WHERE expires_at < now()`,
		`DELETE FROM email_verification_tokens WHERE expires_at < now()`,
	}
	for _, q := range queries {
		if _, err := s.DB.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("delete expired tokens: %w", err)
		}
	}
	return nil
}

type CreateRefreshTokenParams struct {
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

type CreatePasswordResetTokenParams struct {
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

type CreateEmailVerificationTokenParams struct {
	UserID    uuid.UUID
	Email     string
	TokenHash string
	ExpiresAt time.Time
}
