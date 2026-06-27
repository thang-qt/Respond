package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"respond/internal/model"
)

var ErrNotFound = errors.New("not found")

func (s *Store) CreateUser(ctx context.Context, params CreateUserParams) (model.User, error) {
	const query = `
		INSERT INTO users (email, username, password_hash, invited_by_user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, email_verified, role, account_status, username, password_hash, bio, rating,
			wins, losses, draws, default_reveal, locale, username_changed_at, created_at, updated_at
	`

	var user model.User
	err := s.DB.QueryRowContext(ctx, query, params.Email, params.Username, params.PasswordHash, params.InvitedByUserID).Scan(
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
		return model.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *Store) CreateUserWithSettings(ctx context.Context, params CreateUserParams) (model.User, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.User{}, fmt.Errorf("begin user tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const query = `
		INSERT INTO users (email, username, password_hash, invited_by_user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, email_verified, role, account_status, username, password_hash, bio, rating,
			wins, losses, draws, default_reveal, locale, username_changed_at, created_at, updated_at
	`

	var user model.User
	err = tx.QueryRowContext(ctx, query, params.Email, params.Username, params.PasswordHash, params.InvitedByUserID).Scan(
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
		return model.User{}, fmt.Errorf("create user: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO notification_settings (user_id)
		VALUES ($1)
	`, user.ID); err != nil {
		return model.User{}, fmt.Errorf("create notification settings: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.User{}, fmt.Errorf("commit user: %w", err)
	}

	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	const query = `
		SELECT id, email, email_verified, role, account_status, username, password_hash, bio, rating,
			wins, losses, draws, default_reveal, locale, username_changed_at, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`

	var user model.User
	err := s.DB.QueryRowContext(ctx, query, email).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (s *Store) GetUserByEmailOrUsername(ctx context.Context, identifier string) (model.User, error) {
	const query = `
		SELECT id, email, email_verified, role, account_status, username, password_hash, bio, rating,
			wins, losses, draws, default_reveal, locale, username_changed_at, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
		   OR LOWER(username) = LOWER($1)
	`

	var user model.User
	err := s.DB.QueryRowContext(ctx, query, identifier).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("get user by email or username: %w", err)
	}

	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	const query = `
		SELECT id, email, email_verified, role, account_status, username, password_hash, bio, rating,
			wins, losses, draws, default_reveal, locale, username_changed_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (s *Store) UpdateUserPassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	const query = `
		UPDATE users
		SET password_hash = $2,
			updated_at = now()
		WHERE id = $1
	`

	res, err := s.DB.ExecContext(ctx, query, id, passwordHash)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user password rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) UpdateUserEmailVerified(ctx context.Context, id uuid.UUID, verified bool) error {
	const query = `
		UPDATE users
		SET email_verified = $2,
			updated_at = now()
		WHERE id = $1
	`

	res, err := s.DB.ExecContext(ctx, query, id, verified)
	if err != nil {
		return fmt.Errorf("update user email verified: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user email verified rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) UpdateUserBio(ctx context.Context, id uuid.UUID, bio string) error {
	const query = `
		UPDATE users
		SET bio = $2,
			updated_at = now()
		WHERE id = $1
	`

	res, err := s.DB.ExecContext(ctx, query, id, bio)
	if err != nil {
		return fmt.Errorf("update user bio: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user bio rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) UserEmailVerified(ctx context.Context, id uuid.UUID) (bool, error) {
	const query = `
		SELECT email_verified
		FROM users
		WHERE id = $1
	`

	var verified bool
	err := s.DB.QueryRowContext(ctx, query, id).Scan(&verified)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrNotFound
		}
		return false, fmt.Errorf("user email verified: %w", err)
	}

	return verified, nil
}

type UserProfile struct {
	Username     string
	Bio          string
	Rating       int
	Wins         int
	Losses       int
	Draws        int
	DebatesCount int
	CreatedAt    time.Time
}
