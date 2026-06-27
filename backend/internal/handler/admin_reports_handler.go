package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/store"
)

type resolveReportRequest struct {
	Resolution string  `json:"resolution"`
	Note       *string `json:"note"`
}

type updateUserRoleRequest struct {
	Role string `json:"role"`
}

type restoreHiddenContentRequest struct {
	Note *string `json:"note"`
}

type moderateContentRequest struct {
	Note *string `json:"note"`
}

type createUserEnforcementActionRequest struct {
	Action       string   `json:"action"`
	Capabilities []string `json:"capabilities"`
	ExpiresAt    *string  `json:"expires_at"`
	Note         *string  `json:"note"`
}

type revokeUserEnforcementActionRequest struct {
	Note *string `json:"note"`
}

func (h Handler) ListAdminReports(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = "open"
	}
	if status != "open" && status != "dismissed" && status != "actioned" && status != "all" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid status.")
		return
	}

	targetType := strings.TrimSpace(r.URL.Query().Get("target_type"))
	if targetType != "" && targetType != "debate" && targetType != "turn" && targetType != "comment" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid target_type.")
		return
	}

	page, perPage, ok := parsePageParams(w, r, 20, 50)
	if !ok {
		return
	}

	reports, total, err := h.Store.ListReports(r.Context(), store.ListReportsParams{
		Status:     status,
		TargetType: targetType,
		Page:       page,
		PerPage:    perPage,
	})
	if err != nil {
		h.Logger.Error("list admin reports failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondList(w, http.StatusOK, reports, newListMeta(page, perPage, total))
}

func (h Handler) GetAdminReport(w http.ResponseWriter, r *http.Request) {
	reportID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid report id.")
		return
	}

	report, err := h.Store.GetReportByID(r.Context(), reportID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrReportNotFound):
			respondError(w, http.StatusNotFound, "REPORT_NOT_FOUND", "Report not found.")
		case errors.Is(err, store.ErrReportTargetNotFound):
			respondError(w, http.StatusNotFound, "REPORT_TARGET_NOT_FOUND", "Target does not exist.")
		default:
			h.Logger.Error("get admin report failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, report)
}

func (h Handler) ResolveAdminReport(w http.ResponseWriter, r *http.Request) {
	reviewerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	reportID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid report id.")
		return
	}

	var req resolveReportRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	req.Resolution = strings.TrimSpace(req.Resolution)
	if req.Resolution != "dismiss" && req.Resolution != "hide" && req.Resolution != "restore" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid resolution.")
		return
	}
	if req.Note != nil {
		note := strings.TrimSpace(*req.Note)
		if note == "" {
			req.Note = nil
		} else {
			req.Note = &note
		}
	}

	if req.Resolution == "hide" || req.Resolution == "restore" {
		if req.Note == nil {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "A moderator note is required for hide/restore.")
			return
		}
	}

	if req.Note != nil && len([]rune(*req.Note)) > 500 {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "note must be at most 500 characters.")
		return
	}

	resolved, err := h.Store.ResolveReport(r.Context(), store.ResolveReportParams{
		ReportID:   reportID,
		ReviewerID: reviewerID,
		Resolution: req.Resolution,
		Note:       req.Note,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrReportNotFound):
			respondError(w, http.StatusNotFound, "REPORT_NOT_FOUND", "Report not found.")
		case errors.Is(err, store.ErrReportAlreadyClosed):
			respondError(w, http.StatusConflict, "REPORT_ALREADY_RESOLVED", "Report already resolved.")
		case errors.Is(err, store.ErrReportTargetNotFound):
			respondError(w, http.StatusNotFound, "REPORT_TARGET_NOT_FOUND", "Target does not exist.")
		default:
			h.Logger.Error("resolve admin report failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":          resolved.ReportID,
		"status":      resolved.FinalStatus,
		"resolution":  resolved.Resolution,
		"note":        resolved.Note,
		"reviewed_at": resolved.ReviewedAt,
	})
}
