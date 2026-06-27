package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const testDSNEnv = "TEST_DATABASE_URL"

func TestDebateLifecycleCreateJoinSubmitTurnPendingExtension(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "pha")
	sideB := seedTestUser(t, db, "phb")
	tagID := seedTestTag(t, db, "phasecore")

	debate, err := st.CreateDebate(ctx, CreateDebateParams{
		Topic:       fmt.Sprintf("Phase 1 lifecycle test topic %s", uuid.NewString()),
		TagIDs:      []uuid.UUID{tagID},
		TimeMode:    "blitz",
		TurnLimit:   2,
		OpeningTurn: makeLongText(140),
		UserID:      sideA,
	})
	if err != nil {
		t.Fatalf("CreateDebate() error = %v", err)
	}
	t.Cleanup(func() { cleanupDebate(t, db, debate.ID) })

	if debate.Status != "waiting" {
		t.Fatalf("debate status = %s, want waiting", debate.Status)
	}
	if debate.CurrentTurnSide != "b" {
		t.Fatalf("current turn side = %s, want b", debate.CurrentTurnSide)
	}
	if debate.TurnCount != 1 {
		t.Fatalf("turn count = %d, want 1", debate.TurnCount)
	}

	joinResult, err := st.JoinDebate(ctx, debate.ID, sideB)
	if err != nil {
		t.Fatalf("JoinDebate() error = %v", err)
	}
	if joinResult.Status != "active" {
		t.Fatalf("join status = %s, want active", joinResult.Status)
	}
	if joinResult.Side != "b" {
		t.Fatalf("join side = %s, want b", joinResult.Side)
	}
	if joinResult.TurnDeadline == nil {
		t.Fatal("join result must include turn deadline")
	}

	turn, err := st.SubmitTurn(ctx, SubmitTurnParams{
		DebateID: debate.ID,
		UserID:   sideB,
		Content:  makeLongText(160),
	})
	if err != nil {
		t.Fatalf("SubmitTurn() error = %v", err)
	}
	if turn.Side != "b" {
		t.Fatalf("submitted turn side = %s, want b", turn.Side)
	}
	if turn.TurnNumber != 2 {
		t.Fatalf("submitted turn number = %d, want 2", turn.TurnNumber)
	}

	var status string
	var turnCount int
	var currentSide string
	var extensionDeadline sql.NullTime
	var turnDeadline sql.NullTime
	if err := db.QueryRowContext(ctx, `
		SELECT status, turn_count, current_turn_side, extension_deadline, turn_deadline
		FROM debates
		WHERE id = $1
	`, debate.ID).Scan(&status, &turnCount, &currentSide, &extensionDeadline, &turnDeadline); err != nil {
		t.Fatalf("query debate state error = %v", err)
	}

	if status != "pending_extension" {
		t.Fatalf("status = %s, want pending_extension", status)
	}
	if turnCount != 2 {
		t.Fatalf("turn_count = %d, want 2", turnCount)
	}
	if currentSide != "a" {
		t.Fatalf("current_turn_side = %s, want a", currentSide)
	}
	if !extensionDeadline.Valid {
		t.Fatal("extension_deadline should be set")
	}
	if turnDeadline.Valid {
		t.Fatal("turn_deadline should be null while pending extension")
	}
}

func TestDebateLifecycleResignAndReplace(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "pra")
	sideB := seedTestUser(t, db, "prb")
	replacement := seedTestUser(t, db, "prr")
	tagID := seedTestTag(t, db, "phaserepl")

	debate, err := st.CreateDebate(ctx, CreateDebateParams{
		Topic:       fmt.Sprintf("Phase 1 resign replace topic %s", uuid.NewString()),
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

	if _, err := st.JoinDebate(ctx, debate.ID, sideB); err != nil {
		t.Fatalf("JoinDebate() error = %v", err)
	}

	resignResult, err := st.ResignDebate(ctx, debate.ID, sideB)
	if err != nil {
		t.Fatalf("ResignDebate() error = %v", err)
	}
	if resignResult.Status != "waiting_replacement" {
		t.Fatalf("resign status = %s, want waiting_replacement", resignResult.Status)
	}

	replaceResult, err := st.ReplaceDebate(ctx, debate.ID, replacement)
	if err != nil {
		t.Fatalf("ReplaceDebate() error = %v", err)
	}
	if replaceResult.Status != "active" {
		t.Fatalf("replace status = %s, want active", replaceResult.Status)
	}
	if replaceResult.Side != "b" {
		t.Fatalf("replace side = %s, want b", replaceResult.Side)
	}

	var status string
	var openSide sql.NullString
	var resigned sql.NullString
	var sideBUser sql.NullString
	if err := db.QueryRowContext(ctx, `
		SELECT status, open_side, resigned_user_id, side_b_user_id
		FROM debates
		WHERE id = $1
	`, debate.ID).Scan(&status, &openSide, &resigned, &sideBUser); err != nil {
		t.Fatalf("query replacement debate state error = %v", err)
	}

	if status != "active" {
		t.Fatalf("status = %s, want active", status)
	}
	if openSide.Valid {
		t.Fatal("open_side should be null after replacement")
	}
	if resigned.Valid {
		t.Fatal("resigned_user_id should be cleared after replacement")
	}
	if !sideBUser.Valid || sideBUser.String != replacement.String() {
		t.Fatalf("side_b_user_id = %q, want %s", sideBUser.String, replacement.String())
	}
}

func newIntegrationTestStore(t *testing.T) (*Store, *sql.DB) {
	t.Helper()

	dsn := os.Getenv(testDSNEnv)
	if dsn == "" {
		t.Skipf("%s is not set", testDSNEnv)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("db.PingContext() error = %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return New(db), db
}

func seedTestUser(t *testing.T, db *sql.DB, prefix string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	username := fmt.Sprintf("%s%s", prefix, id.String()[:8])
	email := fmt.Sprintf("%s-%s@example.com", prefix, id.String()[:8])

	_, err := db.Exec(`
		INSERT INTO users (id, email, email_verified, username, password_hash, bio)
		VALUES ($1, $2, true, $3, $4, '')
	`, id, email, username, "$2a$12$8Q4j5j.5AUSNf7PzMslYhOiZwqY6ZsHIKh4j/wcgUp3NEPoPFcAck")
	if err != nil {
		t.Fatalf("seed user insert error = %v", err)
	}

	t.Cleanup(func() { cleanupUser(t, db, id) })
	return id
}

func seedTestTag(t *testing.T, db *sql.DB, prefix string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	suffix := id.String()[:8]
	slug := fmt.Sprintf("%s-%s", prefix, suffix)
	name := fmt.Sprintf("%s %s", prefix, suffix)

	_, err := db.Exec(`
		INSERT INTO tags (id, slug, name)
		VALUES ($1, $2, $3)
	`, id, slug, name)
	if err != nil {
		t.Fatalf("seed tag insert error = %v", err)
	}

	t.Cleanup(func() { cleanupTag(t, db, id) })
	return id
}

func cleanupDebate(t *testing.T, db *sql.DB, debateID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(`DELETE FROM debates WHERE id = $1`, debateID); err != nil {
		t.Fatalf("cleanup debate error = %v", err)
	}
}

func cleanupUser(t *testing.T, db *sql.DB, userID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(`DELETE FROM users WHERE id = $1`, userID); err != nil {
		t.Fatalf("cleanup user error = %v", err)
	}
}

func cleanupTag(t *testing.T, db *sql.DB, tagID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(`DELETE FROM tags WHERE id = $1`, tagID); err != nil {
		t.Fatalf("cleanup tag error = %v", err)
	}
}

func makeLongText(length int) string {
	b := make([]byte, 0, length)
	for len(b) < length {
		b = append(b, 'a')
	}
	return string(b)
}
