package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/model"
	"respond/internal/store"
)

func (h Handler) CreateAdminUserEnforcementAction(w http.ResponseWriter, r *http.Request) {
	actorID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	targetUserID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid user id.")
		return
	}

	var req createUserEnforcementActionRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	action := model.UserEnforcementActionType(strings.TrimSpace(req.Action))
	capabilities := make([]model.UserCapability, 0, len(req.Capabilities))
	seenCapabilities := make(map[model.UserCapability]struct{}, len(req.Capabilities))
	for _, rawCapability := range req.Capabilities {
		capability := model.UserCapability(strings.TrimSpace(rawCapability))
		if capability == "" {
			continue
		}
		if _, exists := seenCapabilities[capability]; exists {
			continue
		}
		seenCapabilities[capability] = struct{}{}
		capabilities = append(capabilities, capability)
	}

	var note string
	if req.Note != nil {
		note = strings.TrimSpace(*req.Note)
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		trimmed := strings.TrimSpace(*req.ExpiresAt)
		if trimmed != "" {
			parsed, err := time.Parse(time.RFC3339, trimmed)
			if err != nil {
				respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "expires_at must be RFC3339.")
				return
			}
			expiresAt = &parsed
		}
	}

	actionItem, err := h.Store.CreateUserEnforcementAction(r.Context(), store.CreateUserEnforcementActionParams{
		ActorUserID:  actorID,
		TargetUserID: targetUserID,
		ActionType:   action,
		Capabilities: capabilities,
		ExpiresAt:    expiresAt,
		Note:         note,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
		case errors.Is(err, store.ErrUserEnforcementActionInvalid):
			respondError(w, http.StatusBadRequest, "USER_ENFORCEMENT_ACTION_INVALID", "Invalid enforcement action.")
		case errors.Is(err, store.ErrUserEnforcementCapabilityInvalid):
			respondError(w, http.StatusBadRequest, "USER_ENFORCEMENT_CAPABILITY_INVALID", "Invalid enforcement capabilities.")
		case errors.Is(err, store.ErrUserEnforcementNoteRequired):
			respondError(w, http.StatusBadRequest, "USER_ENFORCEMENT_NOTE_REQUIRED", "A moderator note is required and must be 1-500 characters.")
		default:
			h.Logger.Error("create user enforcement action failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, actionItem)
}

func (h Handler) RevokeAdminUserEnforcementAction(w http.ResponseWriter, r *http.Request) {
	actorID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	targetUserID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid user id.")
		return
	}

	actionID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "action_id")))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid action id.")
		return
	}

	var req revokeUserEnforcementActionRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	var note string
	if req.Note != nil {
		note = strings.TrimSpace(*req.Note)
	}

	actionItem, err := h.Store.RevokeUserEnforcementAction(r.Context(), store.RevokeUserEnforcementActionParams{
		ActorUserID:  actorID,
		TargetUserID: targetUserID,
		ActionID:     actionID,
		Note:         note,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
		case errors.Is(err, store.ErrUserEnforcementActionNotFound):
			respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User enforcement action not found.")
		case errors.Is(err, store.ErrUserEnforcementRevokeInvalid):
			respondError(w, http.StatusConflict, "USER_ENFORCEMENT_REVOKE_INVALID", "Action cannot be revoked in current state.")
		case errors.Is(err, store.ErrUserEnforcementNoteRequired):
			respondError(w, http.StatusBadRequest, "USER_ENFORCEMENT_NOTE_REQUIRED", "A moderator note is required and must be 1-500 characters.")
		default:
			h.Logger.Error("revoke user enforcement action failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":         actionItem.ID,
		"status":     "revoked",
		"revoked_at": actionItem.RevokedAt,
	})
}

func (h Handler) ListAdminUserEnforcementActions(w http.ResponseWriter, r *http.Request) {
	targetUserID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid user id.")
		return
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "expired" && status != "revoked" && status != "all" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid status.")
		return
	}

	page, perPage, ok := parsePageParams(w, r, 20, 50)
	if !ok {
		return
	}

	items, total, err := h.Store.ListUserEnforcementActions(r.Context(), store.ListUserEnforcementActionsParams{
		TargetUserID: targetUserID,
		Status:       status,
		Page:         page,
		PerPage:      perPage,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrUserEnforcementActionInvalid):
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid status.")
		default:
			h.Logger.Error("list user enforcement actions failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondList(w, http.StatusOK, items, newListMeta(page, perPage, total))
}
