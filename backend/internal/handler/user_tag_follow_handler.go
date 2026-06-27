package handler

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/i18n"
)

func (h Handler) ListMyTagFollows(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}

	tags, err := h.Store.ListUserTagFollowsLocalized(r.Context(), userID, i18n.LocaleFromRequest(r))
	if err != nil {
		h.Logger.Error("list user tag follows failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, tags)
}

type ReplaceTagFollowsRequest struct {
	TagIDs []string `json:"tag_ids"`
}

func (h Handler) ReplaceMyTagFollows(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}

	var req ReplaceTagFollowsRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.TagIDs == nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "tag_ids is required.")
		return
	}
	if len(req.TagIDs) > maxFollowedTags {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "tag_ids must contain at most 30 tags.")
		return
	}

	parsedTagIDs := make([]uuid.UUID, 0, len(req.TagIDs))
	seenTagIDs := make(map[uuid.UUID]struct{}, len(req.TagIDs))
	for _, rawTagID := range req.TagIDs {
		tagID, err := uuid.Parse(strings.TrimSpace(rawTagID))
		if err != nil {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid tag_ids.")
			return
		}
		if _, exists := seenTagIDs[tagID]; exists {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "tag_ids must be unique.")
			return
		}
		seenTagIDs[tagID] = struct{}{}
		parsedTagIDs = append(parsedTagIDs, tagID)
	}

	tagCount, err := h.Store.CountTagsByIDs(r.Context(), parsedTagIDs)
	if err != nil {
		h.Logger.Error("count tags by ids failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if tagCount != len(parsedTagIDs) {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid tag_ids.")
		return
	}

	if err := h.Store.ReplaceUserTagFollows(r.Context(), userID, parsedTagIDs); err != nil {
		h.Logger.Error("replace user tag follows failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	tags, err := h.Store.ListUserTagFollowsLocalized(r.Context(), userID, i18n.LocaleFromRequest(r))
	if err != nil {
		h.Logger.Error("list user tag follows failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, tags)
}
