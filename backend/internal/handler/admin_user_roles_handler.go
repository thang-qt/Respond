package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/model"
	"respond/internal/store"
)

func (h Handler) UpdateAdminUserRole(w http.ResponseWriter, r *http.Request) {
	actorID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid user id.")
		return
	}

	var req updateUserRoleRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	role := model.UserRole(strings.TrimSpace(req.Role))
	if role != model.UserRoleUser && role != model.UserRoleModerator && role != model.UserRoleAdmin {
		respondError(w, http.StatusBadRequest, "ROLE_INVALID", "Invalid role.")
		return
	}

	if userID == actorID && role != model.UserRoleAdmin {
		respondError(w, http.StatusForbidden, "ROLE_CHANGE_FORBIDDEN", "You cannot remove your own admin role.")
		return
	}

	if err := h.Store.UpdateUserRoleWithAudit(r.Context(), actorID, userID, role); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
			return
		}
		h.Logger.Error("update admin user role failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":   userID,
		"role": role,
	})
}
