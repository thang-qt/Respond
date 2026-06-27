package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestDebateDrawProposeAndDecline(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "pga")
	sideB := seedTestUser(t, db, "pgb")
	tagID := seedTestTag(t, db, "phasedraw")

	debate := seedActiveDebateForTests(t, st, db, sideA, sideB, tagID, 10)

	propose, err := st.ProposeDrawDebate(ctx, debate.ID, sideA)
	if err != nil {
		t.Fatalf("ProposeDrawDebate() error = %v", err)
	}
	if propose.Status != "pending" {
		t.Fatalf("propose status = %s, want pending", propose.Status)
	}
	if propose.ProposedBy != "a" {
		t.Fatalf("proposed by = %s, want a", propose.ProposedBy)
	}

	decline, err := st.RespondDrawDebate(ctx, debate.ID, sideB, false)
	if err != nil {
		t.Fatalf("RespondDrawDebate(accept=false) error = %v", err)
	}
	if decline.DrawStatus == nil || *decline.DrawStatus != "declined" {
		t.Fatalf("draw status = %v, want declined", decline.DrawStatus)
	}

	var drawProposedBy sql.NullString
	if err := db.QueryRowContext(ctx, `
		SELECT draw_proposed_by
		FROM debates
		WHERE id = $1
	`, debate.ID).Scan(&drawProposedBy); err != nil {
		t.Fatalf("query draw state error = %v", err)
	}
	if drawProposedBy.Valid {
		t.Fatalf("draw_proposed_by = %q, want null", drawProposedBy.String)
	}
}

func TestDebateDrawAcceptFinishesAsDraw(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "pca")
	sideB := seedTestUser(t, db, "pcb")
	tagID := seedTestTag(t, db, "phasedrawacc")

	debate := seedActiveDebateForTests(t, st, db, sideA, sideB, tagID, 10)

	if _, err := st.ProposeDrawDebate(ctx, debate.ID, sideA); err != nil {
		t.Fatalf("ProposeDrawDebate() error = %v", err)
	}

	accept, err := st.RespondDrawDebate(ctx, debate.ID, sideB, true)
	if err != nil {
		t.Fatalf("RespondDrawDebate(accept=true) error = %v", err)
	}
	if accept.Status == nil || *accept.Status != "finished" {
		t.Fatalf("status = %v, want finished", accept.Status)
	}
	if accept.Outcome == nil || *accept.Outcome != "draw" {
		t.Fatalf("outcome = %v, want draw", accept.Outcome)
	}

	var status string
	var outcome sql.NullString
	var winnerSide sql.NullString
	if err := db.QueryRowContext(ctx, `
		SELECT status, outcome, winner_side
		FROM debates
		WHERE id = $1
	`, debate.ID).Scan(&status, &outcome, &winnerSide); err != nil {
		t.Fatalf("query accepted draw debate error = %v", err)
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

func TestDebateDrawCooldownAfterDecline(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "pqa")
	sideB := seedTestUser(t, db, "pqb")
	tagID := seedTestTag(t, db, "phasedrawcd")

	debate := seedActiveDebateForTests(t, st, db, sideA, sideB, tagID, 10)

	if _, err := st.ProposeDrawDebate(ctx, debate.ID, sideA); err != nil {
		t.Fatalf("first ProposeDrawDebate() error = %v", err)
	}
	if _, err := st.RespondDrawDebate(ctx, debate.ID, sideB, false); err != nil {
		t.Fatalf("RespondDrawDebate(accept=false) error = %v", err)
	}

	_, err := st.ProposeDrawDebate(ctx, debate.ID, sideA)
	if !errors.Is(err, ErrDrawCooldown) {
		t.Fatalf("second ProposeDrawDebate() error = %v, want ErrDrawCooldown", err)
	}
}
