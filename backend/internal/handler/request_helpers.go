package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"respond/internal/auth"
)

func (h Handler) currentViewerID(r *http.Request) *uuid.UUID {
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		return &userID
	}
	return nil
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		respondErrorKey(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "error.invalidRequestBody", nil)
		return false
	}
	return true
}

func parsePageParams(w http.ResponseWriter, r *http.Request, defaultPerPage, maxPerPage int) (page, perPage int, ok bool) {
	page = 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			respondErrorKey(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "error.invalidPage", nil)
			return 0, 0, false
		}
		page = value
	}

	perPage = defaultPerPage
	if raw := r.URL.Query().Get("per_page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > maxPerPage {
			respondErrorKey(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "error.invalidPerPage", nil)
			return 0, 0, false
		}
		perPage = value
	}

	return page, perPage, true
}

func newListMeta(page, perPage, total int) *ListMeta {
	totalPages := 0
	if perPage > 0 && total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	return &ListMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
