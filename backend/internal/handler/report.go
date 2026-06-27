package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/model"
	"respond/internal/store"
)

type createReportRequest struct {
	TargetType string  `json:"target_type"`
	TargetID   string  `json:"target_id"`
	Reason     string  `json:"reason"`
	Details    *string `json:"details"`
}

func (h Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	reporterID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}
	reportCapability := model.UserCapabilityReport
	if !h.ensureUserCanPerform(w, r, reporterID, &reportCapability) {
		return
	}

	var req createReportRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	req.TargetType = strings.TrimSpace(req.TargetType)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.TargetType != "debate" && req.TargetType != "turn" && req.TargetType != "comment" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "target_type must be debate, turn, or comment.")
		return
	}

	if req.Reason != "hate" && req.Reason != "harassment" && req.Reason != "spam" && req.Reason != "off_topic" && req.Reason != "illegal" && req.Reason != "other" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid reason.")
		return
	}

	targetID, err := uuid.Parse(strings.TrimSpace(req.TargetID))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid target_id.")
		return
	}

	if req.Details != nil {
		details := strings.TrimSpace(*req.Details)
		if len([]rune(details)) > 500 {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "details must be at most 500 characters.")
			return
		}
		req.Details = &details
	}

	report, _, err := h.Store.CreateReport(r.Context(), store.CreateReportParams{
		ReporterID: reporterID,
		TargetType: req.TargetType,
		TargetID:   targetID,
		Reason:     req.Reason,
		Details:    req.Details,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrReportTargetNotFound):
			respondError(w, http.StatusNotFound, "REPORT_TARGET_NOT_FOUND", "Target does not exist.")
		case errors.Is(err, store.ErrReportDuplicate):
			respondError(w, http.StatusConflict, "REPORT_DUPLICATE", "You already reported this target.")
		case errors.Is(err, store.ErrReportSelfNotAllowed):
			respondError(w, http.StatusBadRequest, "REPORT_SELF_NOT_ALLOWED", "You cannot report your own content.")
		case errors.Is(err, store.ErrReportRateLimited):
			respondError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many reports. Try again later.")
		default:
			h.Logger.Error("create report failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusCreated, report)
}
