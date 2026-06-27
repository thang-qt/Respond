package realtime

import (
	"encoding/json"
	"time"
)

// Event types sent from server to connected clients.
const (
	EventTurnNew           = "turn.new"
	EventDebateJoined      = "debate.joined"
	EventDebateEnded       = "debate.ended"
	EventDebateSeatOpen    = "debate.seat_open"
	EventDebateReplacement = "debate.replacement"
	EventDrawProposed      = "debate.draw_proposed"
	EventDrawResponded     = "debate.draw_responded"
	EventExtensionUpdate   = "debate.extension_update"
	EventDebateEvent       = "debate.event"
	EventTimerSync         = "timer.sync"
	EventNotificationNew   = "notification.new"
)

// Event is the envelope sent over WebSocket.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// MarshalEvent serialises an Event to JSON bytes.
func MarshalEvent(eventType string, data interface{}) ([]byte, error) {
	return json.Marshal(Event{Type: eventType, Data: data})
}

// --- Per-event data payloads ---

type TurnNewData struct {
	ID          string    `json:"id"`
	TurnNumber  int       `json:"turn_number"`
	Side        string    `json:"side"`
	AnonymousID string    `json:"anonymous_id"`
	Content     string    `json:"content"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

type DebateJoinedData struct {
	Side         string     `json:"side"`
	AnonymousID  string     `json:"anonymous_id"`
	TurnDeadline *time.Time `json:"turn_deadline"`
}

type DebateEndedData struct {
	Outcome    *string    `json:"outcome"`
	WinnerSide *string    `json:"winner_side"`
	EndedAt    *time.Time `json:"ended_at"`
}

type DebateSeatOpenData struct {
	Side string `json:"side"`
}

type DebateReplacementData struct {
	Side         string     `json:"side"`
	AnonymousID  string     `json:"anonymous_id"`
	TurnDeadline *time.Time `json:"turn_deadline"`
}

type DrawProposedData struct {
	ProposedBy string `json:"proposed_by"`
}

type DrawRespondedData struct {
	Accepted bool `json:"accepted"`
}

type ExtensionUpdateData struct {
	Status     string  `json:"status"`
	TurnLimit  *int    `json:"turn_limit,omitempty"`
	Outcome    *string `json:"outcome,omitempty"`
	WinnerSide *string `json:"winner_side"`
}

type TimerSyncData struct {
	CurrentTurnSide string     `json:"current_turn_side"`
	TurnDeadline    *time.Time `json:"turn_deadline"`
}

type NotificationNewData struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Message    string  `json:"message"`
	DebateID   *string `json:"debate_id"`
	DebateSlug *string `json:"debate_slug,omitempty"`
	TurnNumber *int    `json:"turn_number,omitempty"`
	CreatedAt  string  `json:"created_at"`
}
