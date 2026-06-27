package handler

import (
	"net/http"

	"respond/internal/auth"
	"respond/internal/model"
	"respond/internal/store"
)

func (h Handler) CreateDebate(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	createCapability := model.UserCapabilityCreateDebate
	if !h.ensureUserCanPerform(w, r, userID, &createCapability) {
		return
	}

	var req CreateDebateRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	parsedTagIDs, validationErr, err := h.validateCreateDebateRequest(r.Context(), &req)
	if err != nil {
		h.Logger.Error("validate create debate failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if validationErr != "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", validationErr)
		return
	}

	activeCount, err := h.Store.CountActiveDebatesForUser(r.Context(), userID)
	if err != nil {
		h.Logger.Error("count active debates failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if activeCount >= 5 {
		respondError(w, http.StatusForbidden, "DEBATE_LIMIT_REACHED", "You already have 5 active debates.")
		return
	}

	debate, err := h.Store.CreateDebate(r.Context(), store.CreateDebateParams{
		Topic:                 req.Topic,
		TagIDs:                parsedTagIDs,
		TimeMode:              req.TimeMode,
		TurnLimit:             req.TurnLimit,
		Context:               req.Context,
		OpeningTurn:           req.OpeningTurn,
		OpeningTurnAIAssisted: req.OpeningTurnAIAssisted,
		OpeningTurnAINote:     req.OpeningTurnAINote,
		UserID:                userID,
	})
	if err != nil {
		h.Logger.Error("create debate failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusCreated, debate)
}
