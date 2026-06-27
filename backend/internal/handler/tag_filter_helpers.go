package handler

import (
	"net/http"
	"strings"
)

func parseTagMode(w http.ResponseWriter, r *http.Request) (string, bool) {
	tagMode := r.URL.Query().Get("tag_mode")
	switch tagMode {
	case "", "any", "all":
		return tagMode, true
	default:
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid tag_mode.")
		return "", false
	}
}

func (h Handler) parseAndValidateTagFilters(w http.ResponseWriter, r *http.Request, logContext string) ([]string, bool) {
	var tagSlugs []string
	if singleTag := strings.TrimSpace(r.URL.Query().Get("tag")); singleTag != "" {
		tagSlugs = append(tagSlugs, strings.ToLower(singleTag))
	}
	if rawTags := strings.TrimSpace(r.URL.Query().Get("tags")); rawTags != "" {
		for _, rawTag := range strings.Split(rawTags, ",") {
			if tag := strings.TrimSpace(rawTag); tag != "" {
				tagSlugs = append(tagSlugs, strings.ToLower(tag))
			}
		}
	}

	if len(tagSlugs) == 0 {
		return nil, true
	}

	seen := make(map[string]struct{}, len(tagSlugs))
	deduped := make([]string, 0, len(tagSlugs))
	for _, tag := range tagSlugs {
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		deduped = append(deduped, tag)
	}
	tagSlugs = deduped

	tagCount, err := h.Store.CountTagsBySlugs(r.Context(), tagSlugs)
	if err != nil {
		h.Logger.Error("count tags by slugs "+logContext+" failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return nil, false
	}
	if tagCount != len(tagSlugs) {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid tag filter.")
		return nil, false
	}

	return tagSlugs, true
}
