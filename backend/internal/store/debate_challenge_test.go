package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestChallengeDebateVisibilityAndJoinGuard(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "chva")
	invited := seedTestUser(t, db, "chvb")
	other := seedTestUser(t, db, "chvc")
	tagID := seedTestTag(t, db, "challenge-visibility")

	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	debate, err := st.CreateDebate(ctx, CreateDebateParams{
		Topic:              fmt.Sprintf("Challenge visibility test %s", uuid.NewString()),
		TagIDs:             []uuid.UUID{tagID},
		TimeMode:           "standard",
		TurnLimit:          10,
		OpeningTurn:        makeLongText(160),
		UserID:             sideA,
		InvitedUserID:      &invited,
		ChallengeExpiresAt: &expiresAt,
	})
	if err != nil {
		t.Fatalf("CreateDebate() error = %v", err)
	}
	t.Cleanup(func() { cleanupDebate(t, db, debate.ID) })

	visible, err := st.IsDebateVisibleToViewer(ctx, debate.ID, nil)
	if err != nil {
		t.Fatalf("IsDebateVisibleToViewer(nil) error = %v", err)
	}
	if visible {
		t.Fatal("expected challenge debate to be hidden from unauthenticated viewer")
	}

	visible, err = st.IsDebateVisibleToViewer(ctx, debate.ID, &sideA)
	if err != nil {
		t.Fatalf("IsDebateVisibleToViewer(sideA) error = %v", err)
	}
	if !visible {
		t.Fatal("expected challenge debate to be visible to challenger")
	}

	visible, err = st.IsDebateVisibleToViewer(ctx, debate.ID, &invited)
	if err != nil {
		t.Fatalf("IsDebateVisibleToViewer(invited) error = %v", err)
	}
	if !visible {
		t.Fatal("expected challenge debate to be visible to invited user")
	}

	visible, err = st.IsDebateVisibleToViewer(ctx, debate.ID, &other)
	if err != nil {
		t.Fatalf("IsDebateVisibleToViewer(other) error = %v", err)
	}
	if visible {
		t.Fatal("expected challenge debate to be hidden from other users")
	}

	if _, err := db.ExecContext(ctx, `UPDATE debates SET status = 'expired' WHERE id = $1`, debate.ID); err != nil {
		t.Fatalf("expire debate error = %v", err)
	}

	visible, err = st.IsDebateVisibleToViewer(ctx, debate.ID, nil)
	if err != nil {
		t.Fatalf("IsDebateVisibleToViewer(nil, expired) error = %v", err)
	}
	if visible {
		t.Fatal("expected expired challenge debate to be hidden from unauthenticated viewer")
	}

	visible, err = st.IsDebateVisibleToViewer(ctx, debate.ID, &sideA)
	if err != nil {
		t.Fatalf("IsDebateVisibleToViewer(sideA, expired) error = %v", err)
	}
	if !visible {
		t.Fatal("expected expired challenge debate to be visible to challenger")
	}

	visible, err = st.IsDebateVisibleToViewer(ctx, debate.ID, &invited)
	if err != nil {
		t.Fatalf("IsDebateVisibleToViewer(invited, expired) error = %v", err)
	}
	if !visible {
		t.Fatal("expected expired challenge debate to be visible to invited user")
	}

	visible, err = st.IsDebateVisibleToViewer(ctx, debate.ID, &other)
	if err != nil {
		t.Fatalf("IsDebateVisibleToViewer(other, expired) error = %v", err)
	}
	if visible {
		t.Fatal("expected expired challenge debate to be hidden from other users")
	}

	_, err = st.JoinDebate(ctx, debate.ID, other)
	if !errors.Is(err, ErrDebateChallengeOnly) {
		t.Fatalf("JoinDebate() error = %v, want ErrDebateChallengeOnly", err)
	}
}

func TestRespondChallengeAcceptFlow(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "chaa")
	invited := seedTestUser(t, db, "chab")
	other := seedTestUser(t, db, "chac")
	tagID := seedTestTag(t, db, "challenge-accept")

	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	debate, err := st.CreateDebate(ctx, CreateDebateParams{
		Topic:              fmt.Sprintf("Challenge accept test %s", uuid.NewString()),
		TagIDs:             []uuid.UUID{tagID},
		TimeMode:           "rapid",
		TurnLimit:          10,
		OpeningTurn:        makeLongText(170),
		UserID:             sideA,
		InvitedUserID:      &invited,
		ChallengeExpiresAt: &expiresAt,
	})
	if err != nil {
		t.Fatalf("CreateDebate() error = %v", err)
	}
	t.Cleanup(func() { cleanupDebate(t, db, debate.ID) })

	_, err = st.RespondChallenge(ctx, debate.ID, other, true)
	if !errors.Is(err, ErrDebateChallengeNotInvited) {
		t.Fatalf("RespondChallenge(non-invited) error = %v, want ErrDebateChallengeNotInvited", err)
	}

	result, err := st.RespondChallenge(ctx, debate.ID, invited, true)
	if err != nil {
		t.Fatalf("RespondChallenge(accept) error = %v", err)
	}
	if !result.Accepted {
		t.Fatal("expected accepted challenge response")
	}
	if result.Status != "active" {
		t.Fatalf("result.Status = %s, want active", result.Status)
	}
	if result.Side == nil || *result.Side != "b" {
		t.Fatalf("result.Side = %v, want b", result.Side)
	}
	if result.TurnDeadline == nil {
		t.Fatal("expected turn deadline on challenge accept")
	}

	_, err = st.RespondChallenge(ctx, debate.ID, invited, true)
	if !errors.Is(err, ErrDebateChallengeResponded) {
		t.Fatalf("RespondChallenge(second accept) error = %v, want ErrDebateChallengeResponded", err)
	}

	var (
		status   string
		sideBID  sql.NullString
		startAt  sql.NullTime
		deadline sql.NullTime
	)
	if err := db.QueryRowContext(ctx, `
		SELECT status, side_b_user_id, started_at, turn_deadline
		FROM debates
		WHERE id = $1
	`, debate.ID).Scan(&status, &sideBID, &startAt, &deadline); err != nil {
		t.Fatalf("query accepted challenge debate state error = %v", err)
	}
	if status != "active" {
		t.Fatalf("status = %s, want active", status)
	}
	if !sideBID.Valid || sideBID.String != invited.String() {
		t.Fatalf("side_b_user_id = %q, want %s", sideBID.String, invited.String())
	}
	if !startAt.Valid || !deadline.Valid {
		t.Fatal("expected started_at and turn_deadline to be set")
	}
}

func TestRespondChallengeDeclineAndList(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "chda")
	invited := seedTestUser(t, db, "chdb")
	tagID := seedTestTag(t, db, "challenge-decline")

	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	debate, err := st.CreateDebate(ctx, CreateDebateParams{
		Topic:              fmt.Sprintf("Challenge decline test %s", uuid.NewString()),
		TagIDs:             []uuid.UUID{tagID},
		TimeMode:           "standard",
		TurnLimit:          10,
		OpeningTurn:        makeLongText(180),
		UserID:             sideA,
		InvitedUserID:      &invited,
		ChallengeExpiresAt: &expiresAt,
	})
	if err != nil {
		t.Fatalf("CreateDebate() error = %v", err)
	}
	t.Cleanup(func() { cleanupDebate(t, db, debate.ID) })

	result, err := st.RespondChallenge(ctx, debate.ID, invited, false)
	if err != nil {
		t.Fatalf("RespondChallenge(decline) error = %v", err)
	}
	if result.Accepted {
		t.Fatal("expected declined challenge response")
	}
	if result.Status != "expired" {
		t.Fatalf("result.Status = %s, want expired", result.Status)
	}

	outbox, total, err := st.ListChallenges(ctx, ListChallengesParams{
		UserID:  sideA,
		Box:     "outbox",
		Status:  "expired",
		Page:    1,
		PerPage: 20,
	})
	if err != nil {
		t.Fatalf("ListChallenges(outbox, expired) error = %v", err)
	}
	if total != 1 || len(outbox) != 1 {
		t.Fatalf("ListChallenges(outbox, expired) got total=%d len=%d, want 1", total, len(outbox))
	}
	if outbox[0].DebateID != debate.ID {
		t.Fatalf("ListChallenges(outbox) debate_id = %s, want %s", outbox[0].DebateID, debate.ID)
	}
}
