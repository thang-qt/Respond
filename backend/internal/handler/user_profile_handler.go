package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/i18n"
	"respond/internal/store"
)

func (h Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.PathValue("username"))
	if username == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid username.")
		return
	}

	profile, targetUserID, err := h.Store.GetUserProfileByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
			return
		}
		h.Logger.Error("get user profile failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if viewerID := h.currentViewerID(r); viewerID != nil && *viewerID != targetUserID {
		blocked, err := h.Store.IsEitherUserBlocked(r.Context(), *viewerID, targetUserID)
		if err != nil {
			h.Logger.Error("check user profile blocked visibility failed", "error", err, "viewer_id", *viewerID, "target_user_id", targetUserID)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
			return
		}
		if blocked {
			respondError(w, http.StatusForbidden, "USER_HIDDEN_BY_BLOCK", "This profile is hidden by your safety settings.")
			return
		}
	}

	respondJSON(w, http.StatusOK, userProfileResponse{
		ID:           targetUserID.String(),
		Username:     profile.Username,
		Bio:          profile.Bio,
		Rating:       profile.Rating,
		Wins:         profile.Wins,
		Losses:       profile.Losses,
		Draws:        profile.Draws,
		DebatesCount: profile.DebatesCount,
		CreatedAt:    profile.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h Handler) ListUserDebates(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.PathValue("username"))
	if username == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid username.")
		return
	}

	page, perPage, ok := parsePageParams(w, r, 20, 50)
	if !ok {
		return
	}

	var viewerID *uuid.UUID
	if uid, ok := auth.UserIDFromContext(r.Context()); ok {
		viewerID = &uid
	}

	debates, total, err := h.Store.ListUserDebatesByUsername(r.Context(), store.ListUserDebatesParams{
		Username: username,
		ViewerID: viewerID,
		Page:     page,
		PerPage:  perPage,
		Locale:   i18n.LocaleFromRequest(r),
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
			return
		}
		if errors.Is(err, store.ErrUserHiddenByBlock) {
			respondError(w, http.StatusForbidden, "USER_HIDDEN_BY_BLOCK", "This profile is hidden by your safety settings.")
			return
		}
		h.Logger.Error("list user debates failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondList(w, http.StatusOK, debates, newListMeta(page, perPage, total))
}
