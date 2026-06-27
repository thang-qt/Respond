package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/coder/websocket"
)

const (
	// Maximum number of queued outbound messages before we drop.
	sendBufferSize = 64

	// Maximum size of an inbound message (only auth token expected).
	maxReadMessageSize = 512

	// How long to wait for a write to succeed.
	writeTimeout = 10 * time.Second

	// Ping interval for keepalive.
	pingInterval = 30 * time.Second
)

// Client represents a single WebSocket connection.
type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	UserID *uuid.UUID // nil for unauthenticated spectators
	logger *slog.Logger
	once   sync.Once
	done   chan struct{}
}

// NewClient wraps a websocket.Conn into a Client.
func NewClient(conn *websocket.Conn, logger *slog.Logger) *Client {
	return &Client{
		conn:   conn,
		send:   make(chan []byte, sendBufferSize),
		logger: logger,
		done:   make(chan struct{}),
	}
}

// Send queues a message for delivery. Drops the message if the
// buffer is full (slow client) rather than blocking the broadcaster.
func (c *Client) Send(msg []byte) {
	select {
	case c.send <- msg:
	default:
		c.logger.Warn("dropping message for slow client", "user_id", c.UserID)
	}
}

// Close initiates a graceful close of the connection.
func (c *Client) Close() {
	c.once.Do(func() {
		close(c.done)
	})
}

// authMessage is the only expected inbound message from clients.
type authMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// ReadPump reads messages from the WebSocket. It only expects an
// optional auth message as the first frame; all subsequent inbound
// messages are discarded. ReadPump blocks until the connection
// closes.
func (c *Client) ReadPump(ctx context.Context, onAuth func(token string) *uuid.UUID) {
	defer c.Close()

	c.conn.SetReadLimit(maxReadMessageSize)

	for {
		msgType, data, err := c.conn.Read(ctx)
		if err != nil {
			return // connection closed or error
		}
		if msgType != websocket.MessageText {
			continue
		}

		// Try to parse as auth message.
		var msg authMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		if msg.Type == "auth" && msg.Token != "" && c.UserID == nil {
			if userID := onAuth(msg.Token); userID != nil {
				c.UserID = userID
				c.logger.Debug("ws client authenticated", "user_id", userID)
			}
		}
		// After auth, we keep reading to detect disconnection
		// but ignore all other messages.
	}
}

// WritePump sends queued messages to the WebSocket and handles
// periodic pings for keepalive. WritePump blocks until the
// connection closes.
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				return
			}

		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}

		case <-c.done:
			return

		case <-ctx.Done():
			return
		}
	}
}
