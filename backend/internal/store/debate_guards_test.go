package store

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestDebateGuardsJoinOwnDebateRejected(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	creator := seedTestUser(t, db, "jwa")
	tagID := seedTestTag(t, db, "phaseguardjoin")

	debate, err := st.CreateDebate(ctx, CreateDebateParams{
		Topic:       fmt.Sprintf("Guard join own debate %s", uuid.NewString()),
		TagIDs:      []uuid.UUID{tagID},
		TimeMode:    "standard",
		TurnLimit:   10,
		OpeningTurn: makeLongText(140),
		UserID:      creator,
	})
	if err != nil {
		t.Fatalf("CreateDebate() error = %v", err)
	}
	t.Cleanup(func() { cleanupDebate(t, db, debate.ID) })

	_, err = st.JoinDebate(ctx, debate.ID, creator)
	if !errors.Is(err, ErrDebateOwnDebate) {
		t.Fatalf("JoinDebate() error = %v, want ErrDebateOwnDebate", err)
	}
}

func TestDebateGuardsSubmitTurnNotYourTurnRejected(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "tna")
	sideB := seedTestUser(t, db, "tnb")
	tagID := seedTestTag(t, db, "phaseguardturn")

	debate := seedActiveDebateForTests(t, st, db, sideA, sideB, tagID, 10)

	_, err := st.SubmitTurn(ctx, SubmitTurnParams{
		DebateID: debate.ID,
		UserID:   sideA,
		Content:  makeLongText(160),
	})
	if !errors.Is(err, ErrTurnNotYourTurn) {
		t.Fatalf("SubmitTurn() error = %v, want ErrTurnNotYourTurn", err)
	}
}

func TestDebateGuardsSubmitTurnWhenNotActiveRejected(t *testing.T) {
	ctx := context.Background()
	st, db := newIntegrationTestStore(t)

	sideA := seedTestUser(t, db, "ina")
	tagID := seedTestTag(t, db, "phaseguardinactive")

	debate, err := st.CreateDebate(ctx, CreateDebateParams{
		Topic:       fmt.Sprintf("Guard inactive submit %s", uuid.NewString()),
		TagIDs:      []uuid.UUID{tagID},
		TimeMode:    "standard",
		TurnLimit:   10,
		OpeningTurn: makeLongText(140),
		UserID:      sideA,
	})
	if err != nil {
		t.Fatalf("CreateDebate() error = %v", err)
	}
	t.Cleanup(func() { cleanupDebate(t, db, debate.ID) })

	_, err = st.SubmitTurn(ctx, SubmitTurnParams{
		DebateID: debate.ID,
		UserID:   sideA,
		Content:  makeLongText(160),
	})
	if !errors.Is(err, ErrDebateNotActive) {
		t.Fatalf("SubmitTurn() error = %v, want ErrDebateNotActive", err)
	}
}
