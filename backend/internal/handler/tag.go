package handler

import (
	"net/http"

	"respond/internal/i18n"
)

func (h Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.Store.ListTagsLocalized(r.Context(), i18n.LocaleFromRequest(r))
	if err != nil {
		h.Logger.Error("list tags failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, tags)
}
