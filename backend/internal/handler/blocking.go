package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/model"
	"respond/internal/store"
)

const debateHiddenByBlockCode = "DEBATE_HIDDEN_BY_BLOCK"

func (h Handler) ensureDebateVisible(w http.ResponseWriter, r *http.Request, debateID uuid.UUID) bool {
	viewerID := h.currentViewerID(r)

	visible, err := h.Store.IsDebateVisibleToViewer(r.Context(), debateID, viewerID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
			return false
		}
		h.Logger.Error("check debate private visibility failed", "error", err, "debate_id", debateID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return false
	}
	if !visible {
		respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		return false
	}

	if viewerID == nil {
		return true
	}

	blocked, err := h.Store.IsDebateBlockedForViewer(r.Context(), debateID, *viewerID)
	if err != nil {
		h.Logger.Error("check debate block visibility failed", "error", err, "debate_id", debateID, "viewer_id", *viewerID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return false
	}
	if blocked {
		respondError(w, http.StatusForbidden, debateHiddenByBlockCode, "This debate is hidden by your safety settings.")
		return false
	}

	return true
}

func (h Handler) canViewHiddenModerationContent(r *http.Request) bool {
	role, ok := auth.UserRoleFromContext(r.Context())
	if !ok {
		return false
	}
	return role == model.UserRoleModerator || role == model.UserRoleAdmin
}

func (h Handler) canViewHiddenDebateContent(r *http.Request, debateID uuid.UUID) (bool, error) {
	if h.canViewHiddenModerationContent(r) {
		return true, nil
	}

	viewerID := h.currentViewerID(r)
	if viewerID == nil {
		return false, nil
	}

	viewer, err := h.Store.GetDebateViewer(r.Context(), debateID, *viewerID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	return viewer.IsParticipant, nil
}

func (h Handler) ensureDebateMutable(w http.ResponseWriter, r *http.Request, debateID uuid.UUID) bool {
	_, hidden, err := h.Store.GetDebateStatus(r.Context(), debateID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
			return false
		}
		h.Logger.Error("get debate status for mutability failed", "error", err, "debate_id", debateID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return false
	}

	if hidden {
		respondError(w, http.StatusForbidden, "DEBATE_HIDDEN_READ_ONLY", "This debate is hidden and read-only.")
		return false
	}

	return true
}

func (h Handler) ensureUserCanPerform(w http.ResponseWriter, r *http.Request, userID uuid.UUID, capability *model.UserCapability) bool {
	err := h.Store.EnsureUserCanPerform(r.Context(), userID, capability)
	if err == nil {
		return true
	}

	switch {
	case errors.Is(err, store.ErrUserBanned):
		respondError(w, http.StatusForbidden, "USER_BANNED", "This account is permanently banned.")
		return false
	case errors.Is(err, store.ErrUserSuspended):
		respondError(w, http.StatusForbidden, "USER_SUSPENDED", "This account is temporarily suspended.")
		return false
	case errors.Is(err, store.ErrUserRestricted):
		code, message := restrictionCapabilityError(capability)
		respondError(w, http.StatusForbidden, code, message)
		return false
	case errors.Is(err, store.ErrNotFound):
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return false
	default:
		h.Logger.Error("check user enforcement failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return false
	}
}

func restrictionCapabilityError(capability *model.UserCapability) (string, string) {
	if capability == nil {
		return "USER_ENFORCEMENT_ACTION_INVALID", "This action is restricted by moderation policy."
	}

	switch *capability {
	case model.UserCapabilityCreateDebate:
		return "USER_RESTRICTED_CREATE_DEBATE", "You cannot create debates while restricted."
	case model.UserCapabilityComment:
		return "USER_RESTRICTED_COMMENT", "You cannot comment while restricted."
	case model.UserCapabilityVote:
		return "USER_RESTRICTED_VOTE", "You cannot vote while restricted."
	case model.UserCapabilityFollow:
		return "USER_RESTRICTED_FOLLOW", "You cannot follow debates while restricted."
	case model.UserCapabilityReport:
		return "USER_RESTRICTED_REPORT", "You cannot submit reports while restricted."
	case model.UserCapabilityInvite:
		return "USER_RESTRICTED_INVITE", "You cannot send invites while restricted."
	default:
		return "USER_ENFORCEMENT_ACTION_INVALID", "This action is restricted by moderation policy."
	}
}
