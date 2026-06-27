package store

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"respond/internal/model"
)

func (s *Store) listDebateTurns(ctx context.Context, debateID uuid.UUID, canViewHiddenContent bool) ([]model.DebateTurn, error) {
	const query = `
		SELECT id, turn_number, side, anonymous_id, content, hidden, ai_assisted, ai_note, is_system, created_at
		FROM turns
		WHERE debate_id = $1
		ORDER BY turn_number ASC
	`

	rows, err := s.DB.QueryContext(ctx, query, debateID)
	if err != nil {
		return nil, fmt.Errorf("list turns: %w", err)
	}
	defer rows.Close()

	var turns []model.DebateTurn
	for rows.Next() {
		var (
			turn   model.DebateTurn
			aiNote sql.NullString
		)
		if err := rows.Scan(&turn.ID, &turn.TurnNumber, &turn.Side, &turn.AnonymousID, &turn.Content, &turn.Hidden, &turn.AIAssisted, &aiNote, &turn.IsSystem, &turn.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan turn: %w", err)
		}
		if turn.Hidden && !canViewHiddenContent {
			turn.Content = "This content has been flagged for review."
		}
		if aiNote.Valid {
			turn.AINote = &aiNote.String
		}
		turns = append(turns, turn)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate turns: %w", err)
	}

	return turns, nil
}

func (s *Store) listDebateEvents(ctx context.Context, debateID uuid.UUID) ([]model.DebateEvent, error) {
	const query = `
		SELECT id, event_type, side, actor_user_id, payload_json, created_at
		FROM debate_events
		WHERE debate_id = $1
		ORDER BY created_at ASC, id ASC
	`

	rows, err := s.DB.QueryContext(ctx, query, debateID)
	if err != nil {
		return nil, fmt.Errorf("list debate events: %w", err)
	}
	defer rows.Close()

	events := make([]model.DebateEvent, 0)
	for rows.Next() {
		var (
			event       model.DebateEvent
			side        sql.NullString
			actorUserID sql.NullString
			payload     []byte
		)
		if err := rows.Scan(&event.ID, &event.EventType, &side, &actorUserID, &payload, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan debate event: %w", err)
		}
		if side.Valid {
			s := side.String
			event.Side = &s
		}
		if actorUserID.Valid {
			parsed, err := uuid.Parse(actorUserID.String)
			if err != nil {
				return nil, fmt.Errorf("parse event actor user id: %w", err)
			}
			event.ActorUserID = &parsed
		}
		event.Payload = payload
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate debate events: %w", err)
	}

	return events, nil
}

func (s *Store) listDebateSeatStints(ctx context.Context, debateID uuid.UUID) (model.DebateParticipantHistory, error) {
	const query = `
		SELECT id, side, user_id, anonymous_id, stint_index, joined_at, left_at, left_reason, replaced_by_stint_id
		FROM debate_seat_stints
		WHERE debate_id = $1
		ORDER BY side ASC, joined_at ASC, id ASC
	`

	rows, err := s.DB.QueryContext(ctx, query, debateID)
	if err != nil {
		return model.DebateParticipantHistory{}, fmt.Errorf("list debate seat stints: %w", err)
	}
	defer rows.Close()

	history := model.DebateParticipantHistory{
		SideA: make([]model.DebateSeatStint, 0),
		SideB: make([]model.DebateSeatStint, 0),
	}

	for rows.Next() {
		var (
			stint             model.DebateSeatStint
			leftAt            sql.NullTime
			leftReason        sql.NullString
			replacedByStintID sql.NullString
		)
		if err := rows.Scan(
			&stint.ID,
			&stint.Side,
			&stint.UserID,
			&stint.AnonymousID,
			&stint.StintIndex,
			&stint.JoinedAt,
			&leftAt,
			&leftReason,
			&replacedByStintID,
		); err != nil {
			return model.DebateParticipantHistory{}, fmt.Errorf("scan debate seat stint: %w", err)
		}

		if leftAt.Valid {
			t := leftAt.Time
			stint.LeftAt = &t
		}
		if leftReason.Valid {
			r := leftReason.String
			stint.LeftReason = &r
		}
		if replacedByStintID.Valid {
			parsed, err := uuid.Parse(replacedByStintID.String)
			if err != nil {
				return model.DebateParticipantHistory{}, fmt.Errorf("parse replaced_by_stint_id: %w", err)
			}
			stint.ReplacedByStintID = &parsed
		}

		if stint.Side == "a" {
			history.SideA = append(history.SideA, stint)
		} else {
			history.SideB = append(history.SideB, stint)
		}
	}
	if err := rows.Err(); err != nil {
		return model.DebateParticipantHistory{}, fmt.Errorf("iterate debate seat stints: %w", err)
	}

	return history, nil
}

func buildDebateTimeline(turns []model.DebateTurn, events []model.DebateEvent) []model.DebateTimelineItem {
	timeline := make([]model.DebateTimelineItem, 0, len(turns)+len(events))
	for i := range turns {
		turn := turns[i]
		timeline = append(timeline, model.DebateTimelineItem{
			Type:      "turn",
			CreatedAt: turn.CreatedAt,
			Turn:      &turn,
		})
	}
	for i := range events {
		event := events[i]
		timeline = append(timeline, model.DebateTimelineItem{
			Type:      "event",
			CreatedAt: event.CreatedAt,
			Event:     &event,
		})
	}

	sort.SliceStable(timeline, func(i, j int) bool {
		left := timeline[i]
		right := timeline[j]

		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.Before(right.CreatedAt)
		}

		leftPriority := 1
		rightPriority := 1
		if left.Type == "turn" {
			leftPriority = 0
		}
		if right.Type == "turn" {
			rightPriority = 0
		}
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}

		leftID := ""
		rightID := ""
		if left.Turn != nil {
			leftID = left.Turn.ID.String()
		} else if left.Event != nil {
			leftID = left.Event.ID.String()
		}
		if right.Turn != nil {
			rightID = right.Turn.ID.String()
		} else if right.Event != nil {
			rightID = right.Event.ID.String()
		}
		return leftID < rightID
	})

	return timeline
}

func selectSideUser(status string, revealed bool, user *model.UserSummary) *model.UserSummary {
	if status != "finished" || !revealed {
		return nil
	}
	return user
}

// ActiveDebateTimer holds the minimal info needed for timer.sync broadcasts.
type ActiveDebateTimer struct {
	DebateID        uuid.UUID
	CurrentTurnSide string
	TurnDeadline    *time.Time
}

// ListActiveDebateTimers returns the debate ID, current turn side,
// and turn deadline for all active debates. Used by the timer sync job.
