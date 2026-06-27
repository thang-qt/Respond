package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"respond/internal/store"
)

func (h Handler) resolveDebateID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	raw := chi.URLParam(r, "id")
	if raw == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Missing debate id.")
		return uuid.Nil, false
	}
	if id, err := uuid.Parse(raw); err == nil {
		return id, true
	}

	id, err := h.Store.GetDebateIDBySlug(r.Context(), raw)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
			return uuid.Nil, false
		}
		h.Logger.Error("get debate by slug failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return uuid.Nil, false
	}

	return id, true
}
