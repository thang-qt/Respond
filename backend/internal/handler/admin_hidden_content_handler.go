package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/store"
)

func (h Handler) ListAdminHiddenContent(w http.ResponseWriter, r *http.Request) {
	targetType := strings.TrimSpace(r.URL.Query().Get("target_type"))
	if targetType != "" && targetType != "debate" && targetType != "turn" && targetType != "comment" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid target_type.")
		return
	}

	limit := 100
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 200 {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit.")
			return
		}
		limit = value
	}

	items, err := h.Store.ListHiddenContent(r.Context(), store.ListHiddenContentParams{
		TargetType: targetType,
		Limit:      limit,
	})
	if err != nil {
		h.Logger.Error("list hidden moderation content failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (h Handler) RestoreAdminHiddenContent(w http.ResponseWriter, r *http.Request) {
	reviewerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	targetType := strings.TrimSpace(chi.URLParam(r, "target_type"))
	if targetType != "debate" && targetType != "turn" && targetType != "comment" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid target_type.")
		return
	}

	targetID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "target_id")))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid target_id.")
		return
	}

	var req restoreHiddenContentRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if req.Note != nil {
		note := strings.TrimSpace(*req.Note)
		if note == "" {
			req.Note = nil
		} else {
			req.Note = &note
		}
	}
	if req.Note == nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "A moderator note is required for restore.")
		return
	}
	if len([]rune(*req.Note)) > 500 {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "note must be at most 500 characters.")
		return
	}

	if err := h.Store.RestoreHiddenTarget(r.Context(), reviewerID, targetType, targetID, req.Note); err != nil {
		switch {
		case errors.Is(err, store.ErrReportTargetNotFound):
			respondError(w, http.StatusNotFound, "REPORT_TARGET_NOT_FOUND", "Target does not exist.")
		default:
			h.Logger.Error("restore hidden moderation content failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"target_type": targetType,
		"target_id":   targetID,
		"restored":    true,
	})
}

func (h Handler) HideAdminContent(w http.ResponseWriter, r *http.Request) {
	h.moderateAdminContent(w, r, "hide")
}

func (h Handler) DirectRestoreAdminContent(w http.ResponseWriter, r *http.Request) {
	h.moderateAdminContent(w, r, "restore")
}

func (h Handler) moderateAdminContent(w http.ResponseWriter, r *http.Request, resolution string) {
	reviewerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	targetType := strings.TrimSpace(chi.URLParam(r, "target_type"))
	if targetType != "debate" && targetType != "turn" && targetType != "comment" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid target_type.")
		return
	}

	targetID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "target_id")))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid target_id.")
		return
	}

	var req moderateContentRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if req.Note != nil {
		note := strings.TrimSpace(*req.Note)
		if note == "" {
			req.Note = nil
		} else {
			req.Note = &note
		}
	}
	if req.Note == nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "A moderator note is required for hide/restore.")
		return
	}
	if len([]rune(*req.Note)) > 500 {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "note must be at most 500 characters.")
		return
	}

	if err := h.Store.ModerateTargetVisibility(r.Context(), reviewerID, targetType, targetID, resolution, req.Note); err != nil {
		switch {
		case errors.Is(err, store.ErrReportTargetNotFound):
			respondError(w, http.StatusNotFound, "REPORT_TARGET_NOT_FOUND", "Target does not exist.")
		default:
			h.Logger.Error("direct moderation action failed", "error", err, "resolution", resolution)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"target_type": targetType,
		"target_id":   targetID,
		"resolution":  resolution,
	})
}
