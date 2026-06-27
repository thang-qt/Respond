package realtime

import (
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// Hub manages rooms of connected WebSocket clients.
// Each debate has a room identified by the debate UUID.
// Users also have personal notification channels keyed by user UUID.
type Hub struct {
	mu     sync.RWMutex
	rooms  map[uuid.UUID]map[*Client]struct{}
	users  map[uuid.UUID]map[*Client]struct{} // per-user notification channels
	logger *slog.Logger
}

// NewHub creates a ready-to-use Hub.
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		rooms:  make(map[uuid.UUID]map[*Client]struct{}),
		users:  make(map[uuid.UUID]map[*Client]struct{}),
		logger: logger,
	}
}

// Subscribe adds a client to a debate room.
func (h *Hub) Subscribe(debateID uuid.UUID, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[debateID]
	if !ok {
		room = make(map[*Client]struct{})
		h.rooms[debateID] = room
	}
	room[c] = struct{}{}

	h.logger.Debug("client subscribed",
		"debate_id", debateID,
		"user_id", c.UserID,
		"room_size", len(room),
	)
}

// Unsubscribe removes a client from a debate room.
// Cleans up the room map entry when empty.
func (h *Hub) Unsubscribe(debateID uuid.UUID, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.rooms[debateID]
	if !ok {
		return
	}
	delete(room, c)

	if len(room) == 0 {
		delete(h.rooms, debateID)
	}

	h.logger.Debug("client unsubscribed",
		"debate_id", debateID,
		"user_id", c.UserID,
		"room_size", len(room),
	)
}

// Broadcast sends an event to every client in a debate room.
func (h *Hub) Broadcast(debateID uuid.UUID, eventType string, data interface{}) {
	msg, err := MarshalEvent(eventType, data)
	if err != nil {
		h.logger.Error("failed to marshal event",
			"event", eventType,
			"debate_id", debateID,
			"error", err,
		)
		return
	}

	h.mu.RLock()
	room, ok := h.rooms[debateID]
	if !ok {
		h.mu.RUnlock()
		return
	}
	// Snapshot clients so we can release the lock before sending.
	clients := make([]*Client, 0, len(room))
	for c := range room {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		c.Send(msg)
	}

	h.logger.Debug("broadcast event",
		"event", eventType,
		"debate_id", debateID,
		"recipients", len(clients),
	)
}

// RoomSize returns the number of connected clients in a debate room.
func (h *Hub) RoomSize(debateID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[debateID])
}

// SubscribeUser adds a client to a user's notification channel.
func (h *Hub) SubscribeUser(userID uuid.UUID, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch, ok := h.users[userID]
	if !ok {
		ch = make(map[*Client]struct{})
		h.users[userID] = ch
	}
	ch[c] = struct{}{}

	h.logger.Debug("user subscribed to notifications",
		"user_id", userID,
		"channel_size", len(ch),
	)
}

// UnsubscribeUser removes a client from a user's notification channel.
func (h *Hub) UnsubscribeUser(userID uuid.UUID, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch, ok := h.users[userID]
	if !ok {
		return
	}
	delete(ch, c)

	if len(ch) == 0 {
		delete(h.users, userID)
	}

	h.logger.Debug("user unsubscribed from notifications",
		"user_id", userID,
		"channel_size", len(ch),
	)
}

// NotifyUser sends an event to all of a user's connected notification clients.
func (h *Hub) NotifyUser(userID uuid.UUID, eventType string, data interface{}) {
	msg, err := MarshalEvent(eventType, data)
	if err != nil {
		h.logger.Error("failed to marshal notification event",
			"event", eventType,
			"user_id", userID,
			"error", err,
		)
		return
	}

	h.mu.RLock()
	ch, ok := h.users[userID]
	if !ok {
		h.mu.RUnlock()
		return
	}
	clients := make([]*Client, 0, len(ch))
	for c := range ch {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		c.Send(msg)
	}
}
