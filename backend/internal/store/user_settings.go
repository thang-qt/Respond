package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"

	"github.com/google/uuid"

	"respond/internal/model"
)

var ErrInvalidEmail = errors.New("invalid email")

type UpdateUserProfileParams struct {
	Bio           *string
	DefaultReveal *bool
	Locale        *string
}

func (s *Store) UpdateUserProfile(ctx context.Context, userID uuid.UUID, params UpdateUserProfileParams) error {
	if params.Bio == nil && params.DefaultReveal == nil && params.Locale == nil {
		return errors.New("no fields to update")
	}

	const query = `
		UPDATE users
		SET
			bio = COALESCE($2, bio),
			default_reveal = COALESCE($3, default_reveal),
			locale = COALESCE($4, locale),
			updated_at = now()
		WHERE id = $1
	`

	res, err := s.DB.ExecContext(ctx, query, userID, params.Bio, params.DefaultReveal, params.Locale)
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user profile rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) UpdateUserEmail(ctx context.Context, userID uuid.UUID, email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}

	const query = `
		UPDATE users
		SET
			email = $2,
			email_verified = false,
			updated_at = now()
		WHERE id = $1
	`

	res, err := s.DB.ExecContext(ctx, query, userID, email)
	if err != nil {
		return fmt.Errorf("update user email: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user email rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) GetNotificationSettings(ctx context.Context, userID uuid.UUID) (model.NotificationSettings, error) {
	// Backfill safety for users created before notification settings initialization.
	if _, err := s.DB.ExecContext(ctx, `
		INSERT INTO notification_settings (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`, userID); err != nil {
		return model.NotificationSettings{}, fmt.Errorf("ensure notification settings: %w", err)
	}

	const query = `
		SELECT
			email_your_turn,
			email_debate_joined,
			email_debate_ended,
			email_turn_expiring,
			email_seat_open,
			email_draw_proposed,
			updated_at
		FROM notification_settings
		WHERE user_id = $1
	`

	var settings model.NotificationSettings
	err := s.DB.QueryRowContext(ctx, query, userID).Scan(
		&settings.EmailYourTurn,
		&settings.EmailDebateJoined,
		&settings.EmailDebateEnded,
		&settings.EmailTurnExpiring,
		&settings.EmailSeatOpen,
		&settings.EmailDrawProposed,
		&settings.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.NotificationSettings{}, ErrNotFound
		}
		return model.NotificationSettings{}, fmt.Errorf("get notification settings: %w", err)
	}

	return settings, nil
}

type UpdateNotificationSettingsParams struct {
	EmailYourTurn     bool
	EmailDebateJoined bool
	EmailDebateEnded  bool
	EmailTurnExpiring bool
	EmailSeatOpen     bool
	EmailDrawProposed bool
}

func (s *Store) UpdateNotificationSettings(ctx context.Context, userID uuid.UUID, params UpdateNotificationSettingsParams) (model.NotificationSettings, error) {
	if _, err := s.DB.ExecContext(ctx, `
		INSERT INTO notification_settings (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`, userID); err != nil {
		return model.NotificationSettings{}, fmt.Errorf("ensure notification settings: %w", err)
	}

	const query = `
		UPDATE notification_settings
		SET
			email_your_turn = $2,
			email_debate_joined = $3,
			email_debate_ended = $4,
			email_turn_expiring = $5,
			email_seat_open = $6,
			email_draw_proposed = $7,
			updated_at = now()
		WHERE user_id = $1
		RETURNING
			email_your_turn,
			email_debate_joined,
			email_debate_ended,
			email_turn_expiring,
			email_seat_open,
			email_draw_proposed,
			updated_at
	`

	var settings model.NotificationSettings
	err := s.DB.QueryRowContext(ctx, query,
		userID,
		params.EmailYourTurn,
		params.EmailDebateJoined,
		params.EmailDebateEnded,
		params.EmailTurnExpiring,
		params.EmailSeatOpen,
		params.EmailDrawProposed,
	).Scan(
		&settings.EmailYourTurn,
		&settings.EmailDebateJoined,
		&settings.EmailDebateEnded,
		&settings.EmailTurnExpiring,
		&settings.EmailSeatOpen,
		&settings.EmailDrawProposed,
		&settings.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.NotificationSettings{}, ErrNotFound
		}
		return model.NotificationSettings{}, fmt.Errorf("update notification settings: %w", err)
	}

	return settings, nil
}
