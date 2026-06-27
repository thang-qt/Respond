package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func normalizeRequiredModerationNote(note *string, missingMessage string) (*string, error) {
	if note != nil {
		trimmed := strings.TrimSpace(*note)
		if trimmed == "" {
			note = nil
		} else {
			note = &trimmed
		}
	}
	if note == nil {
		return nil, fmt.Errorf(missingMessage)
	}
	if len([]rune(*note)) > 500 {
		return nil, fmt.Errorf("invalid note length")
	}
	return note, nil
}

func notifyModerationTargetTx(ctx context.Context, s *Store, tx *sql.Tx, reviewerID uuid.UUID, targetType string, targetID uuid.UUID, resolution string, note *string, notificationErrorPrefix string) error {
	notif, err := moderationNotificationTargetTx(ctx, tx, targetType, targetID)
	if err != nil {
		return err
	}
	message := moderationNotificationMessage(resolution, targetType, notif.TurnNumber, notif.DebateTopic, note)
	notifType := "content_hidden"
	if resolution == "restore" {
		notifType = "content_restored"
	}
	for _, userID := range notif.UserIDs {
		if userID == uuid.Nil || userID == reviewerID {
			continue
		}
		if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
			UserID:     userID,
			Type:       notifType,
			Message:    message,
			DebateID:   notif.DebateID,
			TurnNumber: notif.TurnNumber,
		}); err != nil {
			return fmt.Errorf("%s notification: %w", notificationErrorPrefix, err)
		}
	}
	return nil
}
