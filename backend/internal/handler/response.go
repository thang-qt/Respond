package handler

import (
	"encoding/json"
	"net/http"

	"respond/internal/i18n"
)

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

type ListResponse struct {
	Data interface{} `json:"data"`
	Meta *ListMeta   `json:"meta,omitempty"`
}

type ListMeta struct {
	Page        int  `json:"page"`
	PerPage     int  `json:"per_page"`
	Total       int  `json:"total"`
	TotalPages  int  `json:"total_pages"`
	UnreadCount *int `json:"unread_count,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(SuccessResponse{Data: data})
}

func respondList(w http.ResponseWriter, status int, data interface{}, meta *ListMeta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ListResponse{Data: data, Meta: meta})
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}

func respondErrorKey(w http.ResponseWriter, r *http.Request, status int, code, key string, vars i18n.Vars) {
	respondError(w, status, code, i18n.T(i18n.LocaleFromRequest(r), key, vars))
}

func respondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
