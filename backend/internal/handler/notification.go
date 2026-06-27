package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/store"
)

func (h Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}

	unreadOnly := false
	if raw := r.URL.Query().Get("unread_only"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid unread_only.")
			return
		}
		unreadOnly = value
	}

	page, perPage, ok := parsePageParams(w, r, 20, 50)
	if !ok {
		return
	}

	notifications, total, unreadCount, err := h.Store.ListNotifications(r.Context(), store.ListNotificationsParams{
		UserID:     userID,
		UnreadOnly: unreadOnly,
		Page:       page,
		PerPage:    perPage,
	})
	if err != nil {
		h.Logger.Error("list notifications failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	meta := newListMeta(page, perPage, total)
	meta.UnreadCount = &unreadCount
	respondList(w, http.StatusOK, notifications, meta)
}

func (h Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}

	idParam := chi.URLParam(r, "id")
	notificationID, err := uuid.Parse(idParam)
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid notification id.")
		return
	}

	if err := h.Store.MarkNotificationRead(r.Context(), userID, notificationID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "Notification not found.")
			return
		}
		h.Logger.Error("mark notification read failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondNoContent(w)
}

func (h Handler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}

	if err := h.Store.MarkAllNotificationsRead(r.Context(), userID); err != nil {
		h.Logger.Error("mark all notifications read failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondNoContent(w)
}
