package handler

import (
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/config"
	"respond/internal/model"
	"respond/internal/store"
)

type createInviteRequest struct {
	Email string `json:"email"`
}

type inviteResponse struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	Status     string     `json:"status"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type inviteLineageResponse struct {
	Depth       int     `json:"depth"`
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	InvitedByID *string `json:"invited_by_user_id,omitempty"`
}

func (h Handler) CreateMyInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	if h.Config.SignupMode != config.SignupModeInviteOnly {
		respondError(w, http.StatusForbidden, "INVITE_FORBIDDEN", "Invites are disabled in current signup mode.")
		return
	}

	issuer, err := h.Store.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}
	if !canIssueInvite(issuer, h.Config) {
		respondError(w, http.StatusForbidden, "INVITE_FORBIDDEN", "Your account is not eligible to issue invites yet.")
		return
	}

	inviteCapability := model.UserCapabilityInvite
	if !h.ensureUserCanPerform(w, r, userID, &inviteCapability) {
		return
	}

	var req createInviteRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	invitedEmail := strings.ToLower(strings.TrimSpace(req.Email))
	if invitedEmail == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "email is required.")
		return
	}
	if _, err := mail.ParseAddress(invitedEmail); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid email format.")
		return
	}

	token, err := generateToken()
	if err != nil {
		h.Logger.Error("generate invite token failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	invite, err := h.Store.CreateInvite(r.Context(), store.CreateInviteParams{
		InviterUserID: userID,
		InvitedEmail:  invitedEmail,
		TokenHash:     hashToken(token),
		ExpiresAt:     time.Now().Add(h.Config.InviteTokenTTL),
	})
	if err != nil {
		if errors.Is(err, store.ErrInviteDuplicate) {
			respondError(w, http.StatusConflict, "INVITE_DUPLICATE_ACTIVE", "An active invite already exists for this email.")
			return
		}
		h.Logger.Error("create invite failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if h.Email == nil {
		h.Logger.Error("email service unavailable for signup invite", "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if err := h.Email.QueueSignupInviteEmail(r.Context(), invitedEmail, token, issuer.Username); err != nil {
		h.Logger.Error("queue signup invite email failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusCreated, inviteResponse{
		ID:         invite.ID.String(),
		Email:      invite.InvitedEmail,
		Status:     string(invite.Status),
		ExpiresAt:  invite.ExpiresAt,
		AcceptedAt: invite.AcceptedAt,
		CreatedAt:  invite.CreatedAt,
	})
}

func (h Handler) ListMyInvites(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	status := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))
	if status == "" {
		status = "pending"
	}
	if status != "pending" && status != "accepted" && status != "revoked" && status != "expired" && status != "all" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid status.")
		return
	}

	page := 1
	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid page.")
			return
		}
		page = value
	}

	perPage := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("per_page")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 50 {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid per_page.")
			return
		}
		perPage = value
	}

	invites, total, err := h.Store.ListInvitesByIssuer(r.Context(), store.ListInvitesByIssuerParams{
		IssuerUserID: userID,
		Status:       status,
		Page:         page,
		PerPage:      perPage,
	})
	if err != nil {
		h.Logger.Error("list invites failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	response := make([]inviteResponse, 0, len(invites))
	for _, invite := range invites {
		response = append(response, inviteResponse{
			ID:         invite.ID.String(),
			Email:      invite.InvitedEmail,
			Status:     string(invite.Status),
			ExpiresAt:  invite.ExpiresAt,
			AcceptedAt: invite.AcceptedAt,
			CreatedAt:  invite.CreatedAt,
		})
	}

	respondList(w, http.StatusOK, response, newListMeta(page, perPage, total))
}

func (h Handler) RevokeMyInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	inviteID, err := uuid.Parse(strings.TrimSpace(r.PathValue("id")))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid invite id.")
		return
	}

	invite, err := h.Store.RevokeInviteByID(r.Context(), userID, inviteID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInviteNotFound):
			respondError(w, http.StatusNotFound, "INVITE_NOT_FOUND", "Invite not found.")
		case errors.Is(err, store.ErrInviteRevokeInvalid):
			respondError(w, http.StatusConflict, "INVITE_REVOKE_INVALID", "Invite cannot be revoked in current state.")
		default:
			h.Logger.Error("revoke invite failed", "error", err, "user_id", userID, "invite_id", inviteID)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"id":     invite.ID.String(),
		"status": string(invite.Status),
	})
}

func (h Handler) GetAdminInviteLineage(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(strings.TrimSpace(r.PathValue("id")))
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid user id.")
		return
	}

	items, err := h.Store.GetInviteLineage(r.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
			return
		}
		h.Logger.Error("get invite lineage failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	response := make([]inviteLineageResponse, 0, len(items))
	for _, item := range items {
		var invitedByID *string
		if item.InvitedByID != nil {
			value := item.InvitedByID.String()
			invitedByID = &value
		}
		response = append(response, inviteLineageResponse{
			Depth:       item.Depth,
			UserID:      item.UserID.String(),
			Username:    item.Username,
			InvitedByID: invitedByID,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

func canIssueInvite(user model.User, cfg config.Config) bool {
	if user.Role == model.UserRoleModerator || user.Role == model.UserRoleAdmin {
		return true
	}
	if cfg.InviteRequireVerified && !cfg.IsDevelopment() && !user.EmailVerified {
		return false
	}
	if time.Since(user.CreatedAt) < cfg.InviteMinAge {
		return false
	}
	return true
}
