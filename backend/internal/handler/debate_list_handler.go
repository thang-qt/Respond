package handler

import (
	"errors"
	"net/http"

	"respond/internal/auth"
	"respond/internal/i18n"
	"respond/internal/store"
)

func (h Handler) ListDebates(w http.ResponseWriter, r *http.Request) {
	feed := r.URL.Query().Get("feed")
	tagMode, ok := parseTagMode(w, r)
	if !ok {
		return
	}

	tagSlugs, ok := h.parseAndValidateTagFilters(w, r, "")
	if !ok {
		return
	}

	page, perPage, ok := parsePageParams(w, r, 20, 50)
	if !ok {
		return
	}

	if feed == "following" || feed == "following_tags" {
		if _, ok := auth.UserIDFromContext(r.Context()); !ok {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
			return
		}
	}

	debates, total, err := h.Store.ListDebates(r.Context(), store.ListDebatesParams{
		Feed:     feed,
		TagSlugs: tagSlugs,
		TagMode:  tagMode,
		Page:     page,
		PerPage:  perPage,
		ViewerID: h.currentViewerID(r),
		Locale:   i18n.LocaleFromRequest(r),
	})
	if err != nil {
		if errors.Is(err, store.ErrInvalidFeed) {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid feed.")
			return
		}
		h.Logger.Error("list debates failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondList(w, http.StatusOK, debates, newListMeta(page, perPage, total))
}

func (h Handler) ListExplore(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sort")
	tagMode, ok := parseTagMode(w, r)
	if !ok {
		return
	}

	tagSlugs, ok := h.parseAndValidateTagFilters(w, r, "")
	if !ok {
		return
	}

	page, perPage, ok := parsePageParams(w, r, 20, 50)
	if !ok {
		return
	}

	debates, total, err := h.Store.ListExplore(r.Context(), store.ListExploreParams{
		Sort:     sortBy,
		TagSlugs: tagSlugs,
		TagMode:  tagMode,
		Page:     page,
		PerPage:  perPage,
		ViewerID: h.currentViewerID(r),
		Locale:   i18n.LocaleFromRequest(r),
	})
	if err != nil {
		if errors.Is(err, store.ErrInvalidExploreSort) {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid sort.")
			return
		}
		h.Logger.Error("list explore failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondList(w, http.StatusOK, debates, newListMeta(page, perPage, total))
}
