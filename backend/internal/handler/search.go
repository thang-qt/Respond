package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"respond/internal/i18n"
	"respond/internal/store"
)

func (h Handler) ListDebatesSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	queryLen := utf8.RuneCountInString(query)
	if queryLen < 2 || queryLen > 100 {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "q must be 2-100 characters.")
		return
	}

	sortBy := strings.TrimSpace(r.URL.Query().Get("sort"))
	switch sortBy {
	case "", "relevance", "new":
	default:
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid sort.")
		return
	}

	tagMode, ok := parseTagMode(w, r)
	if !ok {
		return
	}

	tagSlugs, ok := h.parseAndValidateTagFilters(w, r, "for search")
	if !ok {
		return
	}

	page, perPage, ok := parsePageParams(w, r, 20, 50)
	if !ok {
		return
	}

	debates, total, err := h.Store.SearchDebates(r.Context(), store.SearchDebatesParams{
		Query:    query,
		Sort:     sortBy,
		TagSlugs: tagSlugs,
		TagMode:  tagMode,
		Page:     page,
		PerPage:  perPage,
		ViewerID: h.currentViewerID(r),
		Locale:   i18n.LocaleFromRequest(r),
	})
	if err != nil {
		if errors.Is(err, store.ErrInvalidSearchSort) {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid sort.")
			return
		}
		h.Logger.Error("search debates failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondList(w, http.StatusOK, debates, newListMeta(page, perPage, total))
}

func (h Handler) ListTagsSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	queryLen := utf8.RuneCountInString(query)
	if queryLen < 1 || queryLen > 50 {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "q must be 1-50 characters.")
		return
	}

	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 50 {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit.")
			return
		}
		limit = value
	}

	tags, err := h.Store.SearchTagsLocalized(r.Context(), query, limit, i18n.LocaleFromRequest(r))
	if err != nil {
		h.Logger.Error("search tags failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, tags)
}

func (h Handler) ListUsersSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query != "" {
		queryLen := utf8.RuneCountInString(query)
		if queryLen < 1 || queryLen > 50 {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "q must be 1-50 characters.")
			return
		}
	}

	page, perPage, ok := parsePageParams(w, r, 20, 50)
	if !ok {
		return
	}

	users, total, err := h.Store.SearchUsers(r.Context(), store.SearchUsersParams{
		Query:    query,
		TagMode:  "any",
		Page:     page,
		PerPage:  perPage,
		ViewerID: h.currentViewerID(r),
	})
	if err != nil {
		h.Logger.Error("search users failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	data := make([]userProfileResponse, 0, len(users))
	for _, user := range users {
		data = append(data, userProfileResponse{
			ID:           user.ID.String(),
			Username:     user.Username,
			Bio:          user.Bio,
			Rating:       user.Rating,
			Wins:         user.Wins,
			Losses:       user.Losses,
			Draws:        user.Draws,
			DebatesCount: user.DebatesCount,
			CreatedAt:    user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	respondList(w, http.StatusOK, data, newListMeta(page, perPage, total))
}

type exploreUserResponse struct {
	ID           string   `json:"id"`
	Username     string   `json:"username"`
	Bio          string   `json:"bio"`
	Rating       int      `json:"rating"`
	Wins         int      `json:"wins"`
	Losses       int      `json:"losses"`
	Draws        int      `json:"draws"`
	DebatesCount int      `json:"debates_count"`
	SharedTags   []string `json:"shared_tags"`
}

func (h Handler) ListExploreUsers(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query != "" {
		queryLen := utf8.RuneCountInString(query)
		if queryLen < 1 || queryLen > 50 {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "q must be 1-50 characters.")
			return
		}
	}

	tagMode := strings.TrimSpace(r.URL.Query().Get("tag_mode"))
	switch tagMode {
	case "", "any", "all":
	default:
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid tag_mode.")
		return
	}

	tagSlugs, ok := h.parseAndValidateTagFilters(w, r, "for explore users")
	if !ok {
		return
	}

	page, perPage, ok := parsePageParams(w, r, 20, 50)
	if !ok {
		return
	}

	users, total, err := h.Store.ListExploreUsers(r.Context(), store.ListExploreUsersParams{
		Query:    query,
		TagSlugs: tagSlugs,
		TagMode:  tagMode,
		Page:     page,
		PerPage:  perPage,
		ViewerID: h.currentViewerID(r),
	})
	if err != nil {
		if errors.Is(err, store.ErrInvalidTagMode) {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid tag_mode.")
			return
		}
		h.Logger.Error("list explore users failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	data := make([]exploreUserResponse, 0, len(users))
	for _, user := range users {
		data = append(data, exploreUserResponse{
			ID:           user.ID.String(),
			Username:     user.Username,
			Bio:          user.Bio,
			Rating:       user.Rating,
			Wins:         user.Wins,
			Losses:       user.Losses,
			Draws:        user.Draws,
			DebatesCount: user.DebatesCount,
			SharedTags:   user.SharedTags,
		})
	}

	respondList(w, http.StatusOK, data, newListMeta(page, perPage, total))
}
