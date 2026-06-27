package handler

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/realtime"
)

// WSNotifications upgrades the HTTP connection to a WebSocket for
// receiving live notification events. Authentication is required —
// the client must send an auth message as the first frame.
func (h Handler) WSNotifications(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{originPattern(h.Config.FrontendURL)},
	})
	if err != nil {
		h.Logger.Error("ws notifications accept failed", "error", err)
		return
	}

	client := realtime.NewClient(conn, h.Logger)

	ctx, cancel := context.WithCancel(context.Background())

	var subscribedUserID *uuid.UUID

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
		if _, dbErr := h.Store.GetUserByID(ctx, userID); dbErr != nil {
			return nil
		}

		// Subscribe to this user's notification channel.
		h.Hub.SubscribeUser(userID, client)
		subscribedUserID = &userID

		return &userID
	}

	go func() {
		defer cancel()
		defer func() {
			if subscribedUserID != nil {
				h.Hub.UnsubscribeUser(*subscribedUserID, client)
			}
		}()
		defer client.Close()

		go client.WritePump(ctx)

		client.ReadPump(ctx, onAuth)
	}()
}
