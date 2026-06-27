package handler

import (
	"net/http"

	"respond/internal/auth"
	"respond/internal/model"
	"respond/internal/store"
)

func (h Handler) GetMyNotificationSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	settings, err := h.Store.GetNotificationSettings(r.Context(), userID)
	if err != nil {
		h.Logger.Error("get notification settings failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, notificationSettingsResponse(settings))
}

type updateNotificationSettingsRequest struct {
	EmailYourTurn     *bool `json:"email_your_turn"`
	EmailDebateJoined *bool `json:"email_debate_joined"`
	EmailDebateEnded  *bool `json:"email_debate_ended"`
	EmailTurnExpiring *bool `json:"email_turn_expiring"`
	EmailSeatOpen     *bool `json:"email_seat_open"`
	EmailDrawProposed *bool `json:"email_draw_proposed"`
}

func (h Handler) UpdateMyNotificationSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	var req updateNotificationSettingsRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if req.EmailYourTurn == nil ||
		req.EmailDebateJoined == nil ||
		req.EmailDebateEnded == nil ||
		req.EmailTurnExpiring == nil ||
		req.EmailSeatOpen == nil ||
		req.EmailDrawProposed == nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "All notification setting fields are required.")
		return
	}

	settings, err := h.Store.UpdateNotificationSettings(r.Context(), userID, store.UpdateNotificationSettingsParams{
		EmailYourTurn:     *req.EmailYourTurn,
		EmailDebateJoined: *req.EmailDebateJoined,
		EmailDebateEnded:  *req.EmailDebateEnded,
		EmailTurnExpiring: *req.EmailTurnExpiring,
		EmailSeatOpen:     *req.EmailSeatOpen,
		EmailDrawProposed: *req.EmailDrawProposed,
	})
	if err != nil {
		h.Logger.Error("update notification settings failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, notificationSettingsResponse(settings))
}

func notificationSettingsResponse(settings model.NotificationSettings) map[string]bool {
	return map[string]bool{
		"email_your_turn":     settings.EmailYourTurn,
		"email_debate_joined": settings.EmailDebateJoined,
		"email_debate_ended":  settings.EmailDebateEnded,
		"email_turn_expiring": settings.EmailTurnExpiring,
		"email_seat_open":     settings.EmailSeatOpen,
		"email_draw_proposed": settings.EmailDrawProposed,
	}
}

type userProfileResponse struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Bio          string `json:"bio"`
	Rating       int    `json:"rating"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	Draws        int    `json:"draws"`
	DebatesCount int    `json:"debates_count"`
	CreatedAt    string `json:"created_at"`
}

type blockedUserResponse struct {
	Username  string `json:"username"`
	BlockedAt string `json:"blocked_at"`
}
