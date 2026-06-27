package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/store"
)

// ListLobbyEntries handles GET /lobby/challenges
// Optional auth — block-filtered when authenticated.
func (h Handler) ListLobbyEntries(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}

	var tagSlugs []string
	if t := q.Get("tags"); t != "" {
		for _, s := range strings.Split(t, ",") {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				tagSlugs = append(tagSlugs, trimmed)
			}
		}
	} else if t := q.Get("tag"); t != "" {
		tagSlugs = []string{strings.TrimSpace(t)}
	}

	tagMode := q.Get("tag_mode")
	if tagMode == "" {
		tagMode = "any"
	}

	var viewerID *uuid.UUID
	if uid, ok := auth.UserIDFromContext(r.Context()); ok {
		viewerID = &uid
	}

	entries, total, err := h.Store.ListLobbyEntries(r.Context(), store.ListLobbyEntriesParams{
		ViewerID: viewerID,
		TagSlugs: tagSlugs,
		TagMode:  tagMode,
		Page:     page,
		PerPage:  perPage,
	})
	if err != nil {
		h.Logger.Error("list lobby entries failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}
	respondList(w, http.StatusOK, entries, &ListMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetMyLobbyEntry handles GET /users/me/lobby
func (h Handler) GetMyLobbyEntry(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	entry, err := h.Store.GetMyLobbyEntry(r.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "LOBBY_ENTRY_NOT_FOUND", "No lobby entry found.")
			return
		}
		h.Logger.Error("get my lobby entry failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, entry)
}

// GetUserLobbyEntry handles GET /users/:username/lobby
// Optional auth — block-filtered.
func (h Handler) GetUserLobbyEntry(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Missing username.")
		return
	}

	var viewerID *uuid.UUID
	if uid, ok := auth.UserIDFromContext(r.Context()); ok {
		viewerID = &uid
	}

	entry, err := h.Store.GetUserLobbyEntry(r.Context(), viewerID, username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "LOBBY_ENTRY_NOT_FOUND", "No lobby entry found.")
			return
		}
		h.Logger.Error("get user lobby entry failed", "error", err, "username", username)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, entry)
}

type upsertLobbyRequest struct {
	BioNote string   `json:"bio_note"`
	TagIDs  []string `json:"tag_ids"`
}

// UpsertMyLobbyEntry handles PUT /users/me/lobby
func (h Handler) UpsertMyLobbyEntry(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	var req upsertLobbyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body.")
		return
	}

	bioNote := strings.TrimSpace(req.BioNote)
	if len([]rune(bioNote)) > 300 {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "bio_note must be 300 characters or fewer.")
		return
	}

	if len(req.TagIDs) > 15 {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Lobby entry may have at most 15 tags.")
		return
	}

	tagIDs := make([]uuid.UUID, 0, len(req.TagIDs))
	for _, rawID := range req.TagIDs {
		id, err := uuid.Parse(rawID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid tag_id: "+rawID)
			return
		}
		tagIDs = append(tagIDs, id)
	}

	// Validate all tag IDs exist.
	if len(tagIDs) > 0 {
		count, err := h.Store.CountTagsByIDs(r.Context(), tagIDs)
		if err != nil {
			h.Logger.Error("count tags by ids failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
			return
		}
		if count != len(tagIDs) {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "One or more tag IDs are invalid.")
			return
		}
	}

	if err := h.Store.UpsertLobbyEntry(r.Context(), userID, bioNote, tagIDs); err != nil {
		if errors.Is(err, store.ErrLobbyTagLimit) {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Lobby entry may have at most 15 tags.")
			return
		}
		h.Logger.Error("upsert lobby entry failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	entry, err := h.Store.GetMyLobbyEntry(r.Context(), userID)
	if err != nil {
		h.Logger.Error("get lobby entry after upsert failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, entry)
}

// DeleteMyLobbyEntry handles DELETE /users/me/lobby
func (h Handler) DeleteMyLobbyEntry(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	if err := h.Store.DeleteLobbyEntry(r.Context(), userID); err != nil {
		h.Logger.Error("delete lobby entry failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondNoContent(w)
}
