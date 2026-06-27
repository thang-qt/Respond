package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"respond/internal/i18n"
	"respond/internal/model"
)

type ListNotificationsParams struct {
	UserID     uuid.UUID
	UnreadOnly bool
	Page       int
	PerPage    int
}

type CreateNotificationParams struct {
	UserID      uuid.UUID
	Type        string
	Message     string
	MessageKey  string
	MessageVars i18n.Vars
	DebateID    *uuid.UUID
	TurnNumber  *int
}

func (s *Store) ListNotifications(ctx context.Context, params ListNotificationsParams) ([]model.Notification, int, int, error) {
	pagination := normalizePagination(params.Page, params.PerPage, 20, 50)
	perPage := pagination.PerPage

	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1
	`
	if params.UnreadOnly {
		countQuery += " AND is_read = false"
	}
	if err := s.DB.QueryRowContext(ctx, countQuery, params.UserID).Scan(&total); err != nil {
		return nil, 0, 0, fmt.Errorf("count notifications: %w", err)
	}

	var unreadCount int
	if err := s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND is_read = false
	`, params.UserID).Scan(&unreadCount); err != nil {
		return nil, 0, 0, fmt.Errorf("count unread notifications: %w", err)
	}

	query := `
		SELECT n.id, n.type, n.message, n.debate_id, d.slug, n.turn_number, n.is_read, n.created_at
		FROM notifications n
		LEFT JOIN debates d ON d.id = n.debate_id
		WHERE n.user_id = $1
	`
	if params.UnreadOnly {
		query += " AND is_read = false"
	}
	query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"

	rows, err := s.DB.QueryContext(ctx, query, params.UserID, perPage, pagination.Offset)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var item model.Notification
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Message,
			&item.DebateID,
			&item.DebateSlug,
			&item.TurnNumber,
			&item.Read,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, 0, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, 0, fmt.Errorf("iterate notifications: %w", err)
	}

	if notifications == nil {
		notifications = []model.Notification{}
	}

	return notifications, total, unreadCount, nil
}

func (s *Store) MarkNotificationRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE notifications
		SET is_read = true
		WHERE id = $1 AND user_id = $2
	`, notificationID, userID)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark notification read rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) MarkAllNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	if _, err := s.DB.ExecContext(ctx, `
		UPDATE notifications
		SET is_read = true
		WHERE user_id = $1 AND is_read = false
	`, userID); err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}

func (s *Store) CreateNotification(ctx context.Context, params CreateNotificationParams) error {
	if params.UserID == uuid.Nil {
		return errors.New("missing user id")
	}
	if params.Type == "" {
		return errors.New("missing notification type")
	}
	message, err := s.localizedNotificationMessage(ctx, params)
	if err != nil {
		return err
	}

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO notifications (user_id, type, message, debate_id, turn_number)
		VALUES ($1, $2, $3, $4, $5)
	`, params.UserID, params.Type, message, params.DebateID, params.TurnNumber)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}

	if s.OnNotify != nil {
		s.OnNotify(params.UserID, params.Type, message, params.DebateID, params.TurnNumber)
	}
	return nil
}

func (s *Store) createNotificationTx(ctx context.Context, tx *sql.Tx, params CreateNotificationParams) error {
	if params.UserID == uuid.Nil {
		return errors.New("missing user id")
	}
	if params.Type == "" {
		return errors.New("missing notification type")
	}
	message, err := s.localizedNotificationMessage(ctx, params)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO notifications (user_id, type, message, debate_id, turn_number)
		VALUES ($1, $2, $3, $4, $5)
	`, params.UserID, params.Type, message, params.DebateID, params.TurnNumber)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}

	// Note: OnNotify is called here even though the tx hasn't committed yet.
	// This is acceptable because the notification WS push is best-effort;
	// if the tx rolls back the client will simply see a stale notification
	// that gets cleaned up on next poll.
	if s.OnNotify != nil {
		s.OnNotify(params.UserID, params.Type, message, params.DebateID, params.TurnNumber)
	}
	return nil
}

func (s *Store) localizedNotificationMessage(ctx context.Context, params CreateNotificationParams) (string, error) {
	if params.MessageKey == "" {
		if params.Message == "" {
			return "", errors.New("missing notification message")
		}
		return params.Message, nil
	}

	locale, err := s.GetUserLocale(ctx, params.UserID)
	if err != nil {
		return "", err
	}
	return i18n.T(locale, params.MessageKey, params.MessageVars), nil
}

func (s *Store) GetUserLocale(ctx context.Context, userID uuid.UUID) (string, error) {
	var locale string
	if err := s.DB.QueryRowContext(ctx, `SELECT locale FROM users WHERE id = $1`, userID).Scan(&locale); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return i18n.DefaultLocale, ErrNotFound
		}
		return i18n.DefaultLocale, fmt.Errorf("get user locale: %w", err)
	}
	return i18n.NormalizeLocale(locale), nil
}
