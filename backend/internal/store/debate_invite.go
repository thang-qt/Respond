package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrDebateInviteNotCreator  = errors.New("only debate creator can send invites")
	ErrDebateInviteAlreadySent = errors.New("debate invite already sent")
)

// InviteToDebate sends a non-exclusive invite notification for an open waiting debate.
func (s *Store) InviteToDebate(ctx context.Context, debateID, inviterID, invitedUserID uuid.UUID) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var (
		status    string
		sideAUser uuid.UUID
		sideBUser sql.NullString
		topic     string
		invitedID sql.NullString
	)

	if err := tx.QueryRowContext(ctx, `
		SELECT status, side_a_user_id, side_b_user_id, topic, invited_user_id
		FROM debates
		WHERE id = $1
		FOR UPDATE
	`, debateID).Scan(&status, &sideAUser, &sideBUser, &topic, &invitedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("fetch debate: %w", err)
	}

	if sideAUser != inviterID {
		return ErrDebateInviteNotCreator
	}
	if status != "waiting" {
		return ErrDebateNotWaiting
	}
	if invitedID.Valid {
		return ErrDebateChallengeOnly
	}
	if sideBUser.Valid {
		return ErrDebateFull
	}

	blocked, err := isEitherUserBlockedTx(ctx, tx, inviterID, invitedUserID)
	if err != nil {
		return err
	}
	if blocked {
		return ErrDebateUserBlocked
	}

	var alreadySent bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM notifications
			WHERE user_id = $1
			  AND debate_id = $2
			  AND type = 'debate_invited'
			  AND is_read = false
		)
	`, invitedUserID, debateID).Scan(&alreadySent); err != nil {
		return fmt.Errorf("check existing invite notification: %w", err)
	}
	if alreadySent {
		return ErrDebateInviteAlreadySent
	}

	inviterName := "Someone"
	if err := tx.QueryRowContext(ctx, `SELECT username FROM users WHERE id = $1`, inviterID).Scan(&inviterName); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("resolve inviter username: %w", err)
		}
	}
	inviterName = strings.TrimSpace(inviterName)
	if inviterName == "" {
		inviterName = "Someone"
	}

	message := fmt.Sprintf("@%s invited you to join: \"%s\"", inviterName, topic)
	if err := s.createNotificationTx(ctx, tx, CreateNotificationParams{
		UserID:   invitedUserID,
		Type:     "debate_invited",
		Message:  message,
		DebateID: &debateID,
	}); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
