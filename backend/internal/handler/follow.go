package handler

import (
	"errors"
	"net/http"

	"respond/internal/auth"
	"respond/internal/model"
	"respond/internal/store"
)

func (h Handler) FollowDebate(w http.ResponseWriter, r *http.Request) {
	debateID, ok := h.resolveDebateID(w, r)
	if !ok {
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	followCapability := model.UserCapabilityFollow
	if !h.ensureUserCanPerform(w, r, userID, &followCapability) {
		return
	}
	if !h.ensureDebateVisible(w, r, debateID) {
		return
	}
	if !h.ensureDebateMutable(w, r, debateID) {
		return
	}

	if err := h.Store.FollowDebate(r.Context(), debateID, userID); err != nil {
		switch {
		case errors.Is(err, store.ErrDebateNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateHiddenByBlock):
			respondError(w, http.StatusForbidden, debateHiddenByBlockCode, "This debate is hidden by your safety settings.")
		default:
			h.Logger.Error("follow debate failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondNoContent(w)
}

func (h Handler) UnfollowDebate(w http.ResponseWriter, r *http.Request) {
	debateID, ok := h.resolveDebateID(w, r)
	if !ok {
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	followCapability := model.UserCapabilityFollow
	if !h.ensureUserCanPerform(w, r, userID, &followCapability) {
		return
	}
	if !h.ensureDebateVisible(w, r, debateID) {
		return
	}
	if !h.ensureDebateMutable(w, r, debateID) {
		return
	}

	if err := h.Store.UnfollowDebate(r.Context(), debateID, userID); err != nil {
		switch {
		case errors.Is(err, store.ErrDebateNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateHiddenByBlock):
			respondError(w, http.StatusForbidden, debateHiddenByBlockCode, "This debate is hidden by your safety settings.")
		default:
			h.Logger.Error("unfollow debate failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondNoContent(w)
}
