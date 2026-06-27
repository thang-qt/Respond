package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/coder/websocket"

	"respond/internal/auth"
	"respond/internal/realtime"
)

// WSDebate upgrades the HTTP connection to a WebSocket and subscribes
// the client to a debate room for live event updates.
func (h Handler) WSDebate(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")
	debateID, err := uuid.Parse(rawID)
	if err != nil {
		// Try slug lookup.
		resolved, lookupErr := h.Store.GetDebateIDBySlug(r.Context(), rawID)
		if lookupErr != nil {
			http.Error(w, "debate not found", http.StatusNotFound)
			return
		}
		debateID = resolved
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{originPattern(h.Config.FrontendURL)},
	})
	if err != nil {
		h.Logger.Error("ws accept failed", "error", err)
		return
	}

	client := realtime.NewClient(conn, h.Logger)
	h.Hub.Subscribe(debateID, client)

	// Use a background context for the WebSocket lifetime.
	// r.Context() is cancelled when this handler returns, but the
	// WebSocket connection must outlive the HTTP handler.
	ctx, cancel := context.WithCancel(context.Background())

	// onAuth is called when the client sends an auth message.
	onAuth := func(token string) *uuid.UUID {
		parsed, parseErr := auth.ParseToken(token, h.Config.JWTSecret)
		if parseErr != nil || !parsed.Valid {
			return nil
		}
		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok {
			return nil
		}
		sub, ok := claims["sub"].(string)
		if !ok {
			return nil
		}
		userID, parseErr := uuid.Parse(sub)
		if parseErr != nil {
			return nil
		}
		// Verify user still exists (uses background ctx, not r.Context()).
		if _, dbErr := h.Store.GetUserByID(ctx, userID); dbErr != nil {
			return nil
		}
		return &userID
	}

	// Run read and write pumps. When either finishes the client
	// is unsubscribed and cleaned up.
	go func() {
		defer cancel()
		defer h.Hub.Unsubscribe(debateID, client)
		defer client.Close()

		// WritePump runs in its own goroutine.
		go client.WritePump(ctx)

		// ReadPump blocks until disconnection.
		client.ReadPump(ctx, onAuth)
	}()
}

// originPattern converts a frontend URL like "http://localhost:3000"
// to a pattern for websocket.AcceptOptions.OriginPatterns.
func originPattern(frontendURL string) string {
	// nhooyr/websocket accepts glob patterns for origins.
	// For development "http://localhost:3000" → "localhost:3000"
	// For production "https://respond.im" → "respond.im"
	for _, prefix := range []string{"https://", "http://"} {
		if len(frontendURL) > len(prefix) && frontendURL[:len(prefix)] == prefix {
			return frontendURL[len(prefix):]
		}
	}
	return frontendURL
}
