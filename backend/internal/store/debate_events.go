package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"respond/internal/model"
)

const (
	debateEventSeatOpened       = "seat_opened"
	debateEventReplacementJoin  = "replacement_joined"
	debateEventConceded         = "conceded"
	debateEventDrawProposed     = "draw_proposed"
	debateEventDrawDeclined     = "draw_declined"
	debateEventDrawAccepted     = "draw_accepted"
	debateEventExtensionPropose = "extension_proposed"
	debateEventExtensionAccept  = "extension_accepted"
	debateEventExtensionDecline = "extension_declined"
	debateEventWalkover         = "walkover"
	debateEventReplaceExpired   = "replacement_expired"
	debateEventExtendExpired    = "extension_expired"
)

type insertDebateEventParams struct {
	debateID    uuid.UUID
	eventType   string
	side        *string
	actorUserID *uuid.UUID
	payload     map[string]any
	createdAt   *time.Time
}

func insertDebateEventTx(ctx context.Context, tx *sql.Tx, p insertDebateEventParams) (model.DebateEvent, error) {
	payload := p.payload
	if payload == nil {
		payload = map[string]any{}
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return model.DebateEvent{}, fmt.Errorf("marshal event payload: %w", err)
	}

	const q = `
		INSERT INTO debate_events (
			debate_id, event_type, side, actor_user_id, payload_json, created_at
		)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, now()))
		RETURNING id, event_type, side, actor_user_id, payload_json, created_at
	`

	var event model.DebateEvent
	var side sql.NullString
	var actorUserID sql.NullString
	if err := tx.QueryRowContext(
		ctx,
		q,
		p.debateID,
		p.eventType,
		p.side,
		p.actorUserID,
		payloadJSON,
		p.createdAt,
	).Scan(&event.ID, &event.EventType, &side, &actorUserID, &event.Payload, &event.CreatedAt); err != nil {
		return model.DebateEvent{}, fmt.Errorf("insert debate event: %w", err)
	}

	if side.Valid {
		s := side.String
		event.Side = &s
	}
	if actorUserID.Valid {
		parsed, err := uuid.Parse(actorUserID.String)
		if err != nil {
			return model.DebateEvent{}, fmt.Errorf("parse event actor user id: %w", err)
		}
		event.ActorUserID = &parsed
	}

	return event, nil
}

func createSeatStintTx(ctx context.Context, tx *sql.Tx, debateID uuid.UUID, side string, userID uuid.UUID, anonymousID string, joinedAt time.Time) (uuid.UUID, error) {
	const q = `
		INSERT INTO debate_seat_stints (
			debate_id, side, user_id, anonymous_id, stint_index, joined_at
		)
		SELECT $1, $2, $3, $4, COALESCE(MAX(stint_index), 0) + 1, $5
		FROM debate_seat_stints
		WHERE debate_id = $1 AND side = $2
		RETURNING id
	`
	var stintID uuid.UUID
	if err := tx.QueryRowContext(ctx, q, debateID, side, userID, anonymousID, joinedAt).Scan(&stintID); err != nil {
		return uuid.Nil, fmt.Errorf("create seat stint: %w", err)
	}
	return stintID, nil
}

func closeActiveSeatStintTx(ctx context.Context, tx *sql.Tx, debateID uuid.UUID, side, leftReason string, leftAt time.Time) (uuid.UUID, error) {
	const q = `
		UPDATE debate_seat_stints
		SET left_at = $3,
			left_reason = $4
		WHERE id = (
			SELECT id
			FROM debate_seat_stints
			WHERE debate_id = $1
			  AND side = $2
			  AND left_at IS NULL
			ORDER BY joined_at DESC
			LIMIT 1
		)
		RETURNING id
	`
	var stintID uuid.UUID
	if err := tx.QueryRowContext(ctx, q, debateID, side, leftAt, leftReason).Scan(&stintID); err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, nil
		}
		return uuid.Nil, fmt.Errorf("close seat stint: %w", err)
	}
	return stintID, nil
}

func closeAllActiveSeatStintsTx(ctx context.Context, tx *sql.Tx, debateID uuid.UUID, leftReason string, leftAt time.Time) error {
	const q = `
		UPDATE debate_seat_stints
		SET left_at = $2,
			left_reason = $3
		WHERE debate_id = $1
		  AND left_at IS NULL
	`
	if _, err := tx.ExecContext(ctx, q, debateID, leftAt, leftReason); err != nil {
		return fmt.Errorf("close active seat stints: %w", err)
	}
	return nil
}

func linkReplacementStintTx(ctx context.Context, tx *sql.Tx, debateID uuid.UUID, side string, replacementStintID uuid.UUID) error {
	const q = `
		WITH candidate AS (
			SELECT id
			FROM debate_seat_stints
			WHERE debate_id = $1
			  AND side = $2
			  AND left_reason = 'resigned'
			  AND replaced_by_stint_id IS NULL
			ORDER BY left_at DESC NULLS LAST, joined_at DESC
			LIMIT 1
		)
		UPDATE debate_seat_stints
		SET replaced_by_stint_id = $3
		WHERE id IN (SELECT id FROM candidate)
	`
	if _, err := tx.ExecContext(ctx, q, debateID, side, replacementStintID); err != nil {
		return fmt.Errorf("link replacement stint: %w", err)
	}
	return nil
}
