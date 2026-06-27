package store

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"respond/internal/model"
)

func TestDebateEndingsConcedeFinishesDebate(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "pea")
	sideB := seedTestUser(t, db, "peb")
	tagID := seedTestTag(t, db, "phaseend")

	debate := seedActiveDebateForTests(t, st, db, sideA, sideB, tagID, 10)

	result, err := st.ConcedeDebate(ctx, debate.ID, sideA)
	if err != nil {
		t.Fatalf("ConcedeDebate() error = %v", err)
	}

	if result.Status != "finished" {
		t.Fatalf("status = %s, want finished", result.Status)
	}
	if result.Outcome == nil || *result.Outcome != "concession" {
		t.Fatalf("outcome = %v, want concession", result.Outcome)
	}
	if result.WinnerSide == nil || *result.WinnerSide != "b" {
		t.Fatalf("winner side = %v, want b", result.WinnerSide)
	}

	var status string
	var outcome sql.NullString
	var winnerSide sql.NullString
	if err := db.QueryRowContext(ctx, `
		SELECT status, outcome, winner_side
		FROM debates
		WHERE id = $1
	`, debate.ID).Scan(&status, &outcome, &winnerSide); err != nil {
		t.Fatalf("query conceded debate error = %v", err)
	}

	if status != "finished" {
		t.Fatalf("db status = %s, want finished", status)
	}
	if !outcome.Valid || outcome.String != "concession" {
		t.Fatalf("db outcome = %q, want concession", outcome.String)
	}
	if !winnerSide.Valid || winnerSide.String != "b" {
		t.Fatalf("db winner_side = %q, want b", winnerSide.String)
	}
}

func TestDebateEndingsExtensionDeclineEndsAsDraw(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "pda")
	sideB := seedTestUser(t, db, "pdb")
	tagID := seedTestTag(t, db, "phaseext")

	debate := seedActiveDebateForTests(t, st, db, sideA, sideB, tagID, 2)

	if _, err := st.SubmitTurn(ctx, SubmitTurnParams{
		DebateID: debate.ID,
		UserID:   sideB,
		Content:  makeLongText(180),
	}); err != nil {
		t.Fatalf("SubmitTurn() error = %v", err)
	}

	result, err := st.RespondExtension(ctx, debate.ID, sideA, false)
	if err != nil {
		t.Fatalf("RespondExtension(accept=false) error = %v", err)
	}

	if result.Status != "finished" {
		t.Fatalf("status = %s, want finished", result.Status)
	}
	if result.Outcome == nil || *result.Outcome != "draw" {
		t.Fatalf("outcome = %v, want draw", result.Outcome)
	}
	if result.WinnerSide != nil {
		t.Fatalf("winner side = %v, want nil", result.WinnerSide)
	}

	var status string
	var outcome sql.NullString
	var winnerSide sql.NullString
	if err := db.QueryRowContext(ctx, `
		SELECT status, outcome, winner_side
		FROM debates
		WHERE id = $1
	`, debate.ID).Scan(&status, &outcome, &winnerSide); err != nil {
		t.Fatalf("query extension-declined debate error = %v", err)
	}

	if status != "finished" {
		t.Fatalf("db status = %s, want finished", status)
	}
	if !outcome.Valid || outcome.String != "draw" {
		t.Fatalf("db outcome = %q, want draw", outcome.String)
	}
	if winnerSide.Valid {
		t.Fatalf("db winner_side = %q, want null", winnerSide.String)
	}
}

func TestDebateEndingsExtensionAcceptBothSidesReactivates(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "paa")
	sideB := seedTestUser(t, db, "pab")
	tagID := seedTestTag(t, db, "phaseacc")

	debate := seedActiveDebateForTests(t, st, db, sideA, sideB, tagID, 2)

	if _, err := st.SubmitTurn(ctx, SubmitTurnParams{
		DebateID: debate.ID,
		UserID:   sideB,
		Content:  makeLongText(190),
	}); err != nil {
		t.Fatalf("SubmitTurn() error = %v", err)
	}

	first, err := st.RespondExtension(ctx, debate.ID, sideA, true)
	if err != nil {
		t.Fatalf("first RespondExtension(accept=true) error = %v", err)
	}
	if first.Status != "pending_extension" {
		t.Fatalf("first response status = %s, want pending_extension", first.Status)
	}

	second, err := st.RespondExtension(ctx, debate.ID, sideB, true)
	if err != nil {
		t.Fatalf("second RespondExtension(accept=true) error = %v", err)
	}
	if second.Status != "active" {
		t.Fatalf("second response status = %s, want active", second.Status)
	}
	if second.TurnLimit == nil || *second.TurnLimit != 7 {
		t.Fatalf("turn limit = %v, want 7", second.TurnLimit)
	}

	var status string
	var turnLimit int
	var turnDeadline sql.NullTime
	if err := db.QueryRowContext(ctx, `
		SELECT status, turn_limit, turn_deadline
		FROM debates
		WHERE id = $1
	`, debate.ID).Scan(&status, &turnLimit, &turnDeadline); err != nil {
		t.Fatalf("query extension-accepted debate error = %v", err)
	}

	if status != "active" {
		t.Fatalf("db status = %s, want active", status)
	}
	if turnLimit != 7 {
		t.Fatalf("db turn_limit = %d, want 7", turnLimit)
	}
	if !turnDeadline.Valid {
		t.Fatal("db turn_deadline should be set after extension accept")
	}
}

func seedActiveDebateForTests(t *testing.T, st *Store, db *sql.DB, sideA, sideB, tagID uuid.UUID, turnLimit int) model.DebateDetail {
	t.Helper()

	debate, err := st.CreateDebate(context.Background(), CreateDebateParams{
		Topic:       fmt.Sprintf("Phase endings topic %s", uuid.NewString()),
		TagIDs:      []uuid.UUID{tagID},
		TimeMode:    "standard",
		TurnLimit:   turnLimit,
		OpeningTurn: makeLongText(150),
		UserID:      sideA,
	})
	if err != nil {
		t.Fatalf("CreateDebate() error = %v", err)
	}

	t.Cleanup(func() { cleanupDebate(t, db, debate.ID) })

	if _, err := st.JoinDebate(context.Background(), debate.ID, sideB); err != nil {
		t.Fatalf("JoinDebate() error = %v", err)
	}

	return debate
}
