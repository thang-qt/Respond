package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"respond/internal/auth"
	"respond/internal/store"
)

func (h Handler) ListMyBlockedUsers(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	blockedUsers, err := h.Store.ListBlockedUsers(r.Context(), userID)
	if err != nil {
		h.Logger.Error("list blocked users failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	response := make([]blockedUserResponse, 0, len(blockedUsers))
	for _, item := range blockedUsers {
		response = append(response, blockedUserResponse{
			Username:  item.Username,
			BlockedAt: item.CreatedAt.Format(time.RFC3339),
		})
	}

	respondJSON(w, http.StatusOK, response)
}

func (h Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	username := strings.TrimSpace(r.PathValue("username"))
	if username == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid username.")
		return
	}

	targetProfile, targetUserID, err := h.Store.GetUserProfileByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
			return
		}
		h.Logger.Error("resolve block target failed", "error", err, "user_id", userID, "target_username", username)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if targetUserID == userID {
		respondError(w, http.StatusBadRequest, "USER_BLOCK_SELF", "You cannot block yourself.")
		return
	}

	if err := h.Store.BlockUser(r.Context(), userID, targetUserID); err != nil {
		switch {
		case errors.Is(err, store.ErrUserBlockSelf):
			respondError(w, http.StatusBadRequest, "USER_BLOCK_SELF", "You cannot block yourself.")
		case errors.Is(err, store.ErrUserAlreadyBlocked):
			respondError(w, http.StatusConflict, "USER_ALREADY_BLOCKED", "User is already blocked.")
		default:
			h.Logger.Error("block user failed", "error", err, "user_id", userID, "target_user_id", targetUserID)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"username": targetProfile.Username,
		"blocked":  true,
	})
}

func (h Handler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	username := strings.TrimSpace(r.PathValue("username"))
	if username == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid username.")
		return
	}

	_, targetUserID, err := h.Store.GetUserProfileByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
			return
		}
		h.Logger.Error("resolve unblock target failed", "error", err, "user_id", userID, "target_username", username)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if targetUserID == userID {
		respondError(w, http.StatusBadRequest, "USER_BLOCK_SELF", "You cannot unblock yourself.")
		return
	}

	if err := h.Store.UnblockUser(r.Context(), userID, targetUserID); err != nil {
		switch {
		case errors.Is(err, store.ErrUserNotBlocked):
			respondError(w, http.StatusNotFound, "USER_NOT_BLOCKED", "User is not currently blocked.")
		default:
			h.Logger.Error("unblock user failed", "error", err, "user_id", userID, "target_user_id", targetUserID)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondNoContent(w)
}
