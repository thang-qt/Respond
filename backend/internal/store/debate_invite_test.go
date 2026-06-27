package store

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInviteToDebateAndJoinMarksInviteRead(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "invita")
	invited := seedTestUser(t, db, "invitb")
	tagID := seedTestTag(t, db, "invite-flow")

	debate, err := st.CreateDebate(ctx, CreateDebateParams{
		Topic:       fmt.Sprintf("Phase 2 invite flow %s", uuid.NewString()),
		TagIDs:      []uuid.UUID{tagID},
		TimeMode:    "standard",
		TurnLimit:   10,
		OpeningTurn: makeLongText(150),
		UserID:      sideA,
	})
	if err != nil {
		t.Fatalf("CreateDebate() error = %v", err)
	}
	t.Cleanup(func() { cleanupDebate(t, db, debate.ID) })

	if err := st.InviteToDebate(ctx, debate.ID, sideA, invited); err != nil {
		t.Fatalf("InviteToDebate() error = %v", err)
	}

	var unreadBeforeJoin int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1
		  AND debate_id = $2
		  AND type = 'debate_invited'
		  AND is_read = false
	`, invited, debate.ID).Scan(&unreadBeforeJoin); err != nil {
		t.Fatalf("count invite notifications before join error = %v", err)
	}
	if unreadBeforeJoin != 1 {
		t.Fatalf("unread invite notifications before join = %d, want 1", unreadBeforeJoin)
	}

	if _, err := st.JoinDebate(ctx, debate.ID, invited); err != nil {
		t.Fatalf("JoinDebate() error = %v", err)
	}

	var unreadAfterJoin int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM notifications
		WHERE debate_id = $1
		  AND type = 'debate_invited'
		  AND is_read = false
	`, debate.ID).Scan(&unreadAfterJoin); err != nil {
		t.Fatalf("count invite notifications after join error = %v", err)
	}
	if unreadAfterJoin != 0 {
		t.Fatalf("unread invite notifications after join = %d, want 0", unreadAfterJoin)
	}
}

func TestInviteToDebateGuards(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "invga")
	other := seedTestUser(t, db, "invgb")
	invitee := seedTestUser(t, db, "invgc")
	tagID := seedTestTag(t, db, "invite-guards")

	debate, err := st.CreateDebate(ctx, CreateDebateParams{
		Topic:       fmt.Sprintf("Phase 2 invite guards %s", uuid.NewString()),
		TagIDs:      []uuid.UUID{tagID},
		TimeMode:    "rapid",
		TurnLimit:   10,
		OpeningTurn: makeLongText(160),
		UserID:      sideA,
	})
	if err != nil {
		t.Fatalf("CreateDebate() error = %v", err)
	}
	t.Cleanup(func() { cleanupDebate(t, db, debate.ID) })

	if err := st.InviteToDebate(ctx, debate.ID, other, invitee); !errors.Is(err, ErrDebateInviteNotCreator) {
		t.Fatalf("InviteToDebate(non-creator) error = %v, want ErrDebateInviteNotCreator", err)
	}

	if err := st.InviteToDebate(ctx, debate.ID, sideA, invitee); err != nil {
		t.Fatalf("InviteToDebate(first) error = %v", err)
	}
	if err := st.InviteToDebate(ctx, debate.ID, sideA, invitee); !errors.Is(err, ErrDebateInviteAlreadySent) {
		t.Fatalf("InviteToDebate(duplicate) error = %v, want ErrDebateInviteAlreadySent", err)
	}

	challengeInvitee := seedTestUser(t, db, "invgd")
	challengeExpiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	challengeDebate, err := st.CreateDebate(ctx, CreateDebateParams{
		Topic:              fmt.Sprintf("Phase 2 invite challenge guard %s", uuid.NewString()),
		TagIDs:             []uuid.UUID{tagID},
		TimeMode:           "standard",
		TurnLimit:          10,
		OpeningTurn:        makeLongText(150),
		UserID:             sideA,
		InvitedUserID:      &challengeInvitee,
		ChallengeExpiresAt: &challengeExpiresAt,
	})
	if err != nil {
		t.Fatalf("CreateDebate(challenge) error = %v", err)
	}
	t.Cleanup(func() { cleanupDebate(t, db, challengeDebate.ID) })

	if err := st.InviteToDebate(ctx, challengeDebate.ID, sideA, invitee); !errors.Is(err, ErrDebateChallengeOnly) {
		t.Fatalf("InviteToDebate(challenge) error = %v, want ErrDebateChallengeOnly", err)
	}

	blockedUser := seedTestUser(t, db, "invge")
	if _, err := db.ExecContext(ctx, `
		INSERT INTO user_blocks (blocker_user_id, blocked_user_id)
		VALUES ($1, $2)
	`, blockedUser, sideA); err != nil {
		t.Fatalf("insert user block error = %v", err)
	}

	if err := st.InviteToDebate(ctx, debate.ID, sideA, blockedUser); !errors.Is(err, ErrDebateUserBlocked) {
		t.Fatalf("InviteToDebate(blocked) error = %v, want ErrDebateUserBlocked", err)
	}
}

func TestProcessExpirationMarksInviteNotificationsRead(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "invea")
	invitee := seedTestUser(t, db, "inveb")
	tagID := seedTestTag(t, db, "invite-expiry")

	debate, err := st.CreateDebate(ctx, CreateDebateParams{
		Topic:       fmt.Sprintf("Phase 2 invite expiry %s", uuid.NewString()),
		TagIDs:      []uuid.UUID{tagID},
		TimeMode:    "standard",
		TurnLimit:   10,
		OpeningTurn: makeLongText(160),
		UserID:      sideA,
	})
	if err != nil {
		t.Fatalf("CreateDebate() error = %v", err)
	}
	t.Cleanup(func() { cleanupDebate(t, db, debate.ID) })

	if err := st.InviteToDebate(ctx, debate.ID, sideA, invitee); err != nil {
		t.Fatalf("InviteToDebate() error = %v", err)
	}

	if _, err := db.ExecContext(ctx, `UPDATE debates SET created_at = now() - INTERVAL '15 days' WHERE id = $1`, debate.ID); err != nil {
		t.Fatalf("backdate debate error = %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	st.ProcessExpirations(ctx, logger)

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM debates WHERE id = $1`, debate.ID).Scan(&status); err != nil {
		t.Fatalf("query debate status error = %v", err)
	}
	if status != "expired" {
		t.Fatalf("debate status = %s, want expired", status)
	}

	var unread int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM notifications
		WHERE debate_id = $1
		  AND type = 'debate_invited'
		  AND is_read = false
	`, debate.ID).Scan(&unread); err != nil {
		t.Fatalf("count unread invite notifications error = %v", err)
	}
	if unread != 0 {
		t.Fatalf("unread invite notifications = %d, want 0", unread)
	}
}
