package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"respond/internal/auth"
	"respond/internal/model"
	"respond/internal/realtime"
	"respond/internal/store"
)

func (h Handler) CreateChallengeDebate(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	createCapability := model.UserCapabilityCreateDebate
	if !h.ensureUserCanPerform(w, r, userID, &createCapability) {
		return
	}

	var req CreateChallengeRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	req.InvitedUsername = strings.TrimSpace(req.InvitedUsername)
	if req.InvitedUsername == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invited_username is required.")
		return
	}

	parsedTagIDs, validationErr, err := h.validateCreateDebateRequest(r.Context(), &req.CreateDebateRequest)
	if err != nil {
		h.Logger.Error("validate create challenge failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if validationErr != "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", validationErr)
		return
	}

	_, invitedUserID, err := h.Store.GetUserProfileByUsername(r.Context(), req.InvitedUsername)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
			return
		}
		h.Logger.Error("resolve invited user failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if invitedUserID == userID {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "You cannot challenge yourself.")
		return
	}

	blocked, err := h.Store.IsEitherUserBlocked(r.Context(), userID, invitedUserID)
	if err != nil {
		h.Logger.Error("challenge blocked relationship check failed", "error", err, "user_id", userID, "invited_user_id", invitedUserID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if blocked {
		respondError(w, http.StatusForbidden, "DEBATE_USER_BLOCKED", "You cannot challenge because one of you has blocked the other.")
		return
	}

	activeCount, err := h.Store.CountActiveDebatesForUser(r.Context(), userID)
	if err != nil {
		h.Logger.Error("count active debates failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if activeCount >= 5 {
		respondError(w, http.StatusForbidden, "DEBATE_LIMIT_REACHED", "You already have 5 active debates.")
		return
	}

	challengerUsername := "Someone"
	if challenger, err := h.Store.GetUserByID(r.Context(), userID); err != nil {
		h.Logger.Warn("resolve challenger username failed", "error", err, "user_id", userID)
	} else if strings.TrimSpace(challenger.Username) != "" {
		challengerUsername = challenger.Username
	}

	challengeExpiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	debate, err := h.Store.CreateDebate(r.Context(), store.CreateDebateParams{
		Topic:                    req.Topic,
		TagIDs:                   parsedTagIDs,
		TimeMode:                 req.TimeMode,
		TurnLimit:                req.TurnLimit,
		Context:                  req.Context,
		OpeningTurn:              req.OpeningTurn,
		OpeningTurnAIAssisted:    req.OpeningTurnAIAssisted,
		OpeningTurnAINote:        req.OpeningTurnAINote,
		UserID:                   userID,
		InvitedUserID:            &invitedUserID,
		ChallengeExpiresAt:       &challengeExpiresAt,
		ChallengeIdentityVisible: true,
	})
	if err != nil {
		h.Logger.Error("create challenge debate failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	message := fmt.Sprintf("@%s challenged you: \"%s\"", challengerUsername, req.Topic)
	if err := h.Store.CreateNotification(r.Context(), store.CreateNotificationParams{
		UserID:   invitedUserID,
		Type:     "challenge_received",
		Message:  message,
		DebateID: &debate.ID,
	}); err != nil {
		h.Logger.Warn("create challenge notification failed", "error", err, "debate_id", debate.ID, "invited_user_id", invitedUserID)
	}

	respondJSON(w, http.StatusCreated, debate)
}

func (h Handler) CreateRechallengeDebate(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	createCapability := model.UserCapabilityCreateDebate
	if !h.ensureUserCanPerform(w, r, userID, &createCapability) {
		return
	}

	sourceDebateID, ok := h.resolveDebateID(w, r)
	if !ok {
		return
	}
	if !h.ensureDebateVisible(w, r, sourceDebateID) {
		return
	}

	var req CreateRechallengeRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	parsedTagIDs, validationErr, err := h.validateCreateDebateRequest(r.Context(), &req.CreateDebateRequest)
	if err != nil {
		h.Logger.Error("validate create rechallenge failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if validationErr != "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", validationErr)
		return
	}

	activeCount, err := h.Store.CountActiveDebatesForUser(r.Context(), userID)
	if err != nil {
		h.Logger.Error("count active debates failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if activeCount >= 5 {
		respondError(w, http.StatusForbidden, "DEBATE_LIMIT_REACHED", "You already have 5 active debates.")
		return
	}

	rechallengeTarget, err := h.Store.ResolveRechallengeInvitedUser(r.Context(), sourceDebateID, userID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateNotFinished):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_FINISHED", "Rechallenge is only available for finished debates.")
		case errors.Is(err, store.ErrDebateNotParticipant):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_PARTICIPANT", "You are not a participant in this debate.")
		case errors.Is(err, store.ErrDebateUserBlocked):
			respondError(w, http.StatusForbidden, "DEBATE_USER_BLOCKED", "You cannot rechallenge because one of you has blocked the other.")
		default:
			h.Logger.Error("resolve rechallenge invited user failed", "error", err, "source_debate_id", sourceDebateID, "user_id", userID)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}
	invitedUserID := rechallengeTarget.InvitedUserID

	challengerUsername := "Someone"
	if challenger, err := h.Store.GetUserByID(r.Context(), userID); err != nil {
		h.Logger.Warn("resolve rechallenge challenger username failed", "error", err, "user_id", userID)
	} else if strings.TrimSpace(challenger.Username) != "" {
		challengerUsername = challenger.Username
	}

	challengeExpiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	debate, err := h.Store.CreateDebate(r.Context(), store.CreateDebateParams{
		Topic:                    req.Topic,
		TagIDs:                   parsedTagIDs,
		TimeMode:                 req.TimeMode,
		TurnLimit:                req.TurnLimit,
		Context:                  req.Context,
		OpeningTurn:              req.OpeningTurn,
		OpeningTurnAIAssisted:    req.OpeningTurnAIAssisted,
		OpeningTurnAINote:        req.OpeningTurnAINote,
		UserID:                   userID,
		InvitedUserID:            &invitedUserID,
		ChallengeExpiresAt:       &challengeExpiresAt,
		ChallengeIdentityVisible: rechallengeTarget.ChallengeIdentityVisible,
	})
	if err != nil {
		h.Logger.Error("create rechallenge debate failed", "error", err, "source_debate_id", sourceDebateID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	message := fmt.Sprintf("Your previous opponent challenged you: \"%s\"", req.Topic)
	if rechallengeTarget.ChallengeIdentityVisible {
		message = fmt.Sprintf("@%s challenged you: \"%s\"", challengerUsername, req.Topic)
	}
	if err := h.Store.CreateNotification(r.Context(), store.CreateNotificationParams{
		UserID:   invitedUserID,
		Type:     "challenge_received",
		Message:  message,
		DebateID: &debate.ID,
	}); err != nil {
		h.Logger.Warn("create rechallenge notification failed", "error", err, "debate_id", debate.ID, "invited_user_id", invitedUserID)
	}

	respondJSON(w, http.StatusCreated, debate)
}

func (h Handler) ListMyChallenges(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}

	box := strings.TrimSpace(r.URL.Query().Get("box"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	page, perPage, ok := parsePageParams(w, r, 20, 50)
	if !ok {
		return
	}

	items, total, err := h.Store.ListChallenges(r.Context(), store.ListChallengesParams{
		UserID:  userID,
		Box:     box,
		Status:  status,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidChallengeBox), errors.Is(err, store.ErrInvalidChallengeStatus):
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid challenge filters.")
		default:
			h.Logger.Error("list my challenges failed", "error", err, "user_id", userID)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondList(w, http.StatusOK, items, newListMeta(page, perPage, total))
}

type RespondChallengeRequest struct {
	Accept bool `json:"accept"`
}

type InviteDebateRequest struct {
	InvitedUsername string `json:"invited_username"`
}

func (h Handler) RespondChallenge(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	if !h.ensureUserCanPerform(w, r, userID, nil) {
		return
	}

	debateID, ok := h.resolveDebateID(w, r)
	if !ok {
		return
	}
	if !h.ensureDebateMutable(w, r, debateID) {
		return
	}

	activeCount, err := h.Store.CountActiveDebatesForUser(r.Context(), userID)
	if err != nil {
		h.Logger.Error("count active debates failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	var req RespondChallengeRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.Accept && activeCount >= 5 {
		respondError(w, http.StatusForbidden, "DEBATE_LIMIT_REACHED", "You already have 5 active debates.")
		return
	}

	result, err := h.Store.RespondChallenge(r.Context(), debateID, userID, req.Accept)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateChallengeOnly):
			respondError(w, http.StatusForbidden, "DEBATE_CHALLENGE_ONLY", "This debate is not an invite-only challenge.")
		case errors.Is(err, store.ErrDebateChallengeNotInvited):
			respondError(w, http.StatusForbidden, "DEBATE_CHALLENGE_NOT_INVITED", "You are not the invited user for this challenge.")
		case errors.Is(err, store.ErrDebateChallengeResponded):
			respondError(w, http.StatusConflict, "DEBATE_CHALLENGE_ALREADY_RESPONDED", "This challenge already has a response.")
		case errors.Is(err, store.ErrDebateChallengeExpired):
			respondError(w, http.StatusConflict, "DEBATE_CHALLENGE_EXPIRED", "This challenge has expired.")
		case errors.Is(err, store.ErrDebateUserBlocked):
			respondError(w, http.StatusForbidden, "DEBATE_USER_BLOCKED", "You cannot accept because one of you has blocked the other.")
		default:
			h.Logger.Error("respond challenge failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	if result.Accepted && result.Side != nil && result.AnonymousID != nil {
		h.Hub.Broadcast(debateID, "debate.joined", realtime.DebateJoinedData{
			Side:         *result.Side,
			AnonymousID:  *result.AnonymousID,
			TurnDeadline: result.TurnDeadline,
		})
	}

	respondJSON(w, http.StatusOK, result)
}

func (h Handler) InviteDebate(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	if !h.ensureUserCanPerform(w, r, userID, nil) {
		return
	}

	debateID, ok := h.resolveDebateID(w, r)
	if !ok {
		return
	}
	if !h.ensureDebateMutable(w, r, debateID) {
		return
	}

	var req InviteDebateRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	req.InvitedUsername = strings.TrimSpace(req.InvitedUsername)
	if req.InvitedUsername == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invited_username is required.")
		return
	}

	_, invitedUserID, err := h.Store.GetUserProfileByUsername(r.Context(), req.InvitedUsername)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
			return
		}
		h.Logger.Error("resolve invited user failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if invitedUserID == userID {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "You cannot invite yourself.")
		return
	}

	if err := h.Store.InviteToDebate(r.Context(), debateID, userID, invitedUserID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateInviteNotCreator):
			respondError(w, http.StatusForbidden, "DEBATE_INVITE_FORBIDDEN", "Only the debate creator can invite users.")
		case errors.Is(err, store.ErrDebateChallengeOnly):
			respondError(w, http.StatusForbidden, "DEBATE_CHALLENGE_ONLY", "Invite actions are for open debates only.")
		case errors.Is(err, store.ErrDebateNotWaiting):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_ACTIVE", "This debate is not open for invites.")
		case errors.Is(err, store.ErrDebateFull):
			respondError(w, http.StatusConflict, "DEBATE_FULL", "This debate already has two sides.")
		case errors.Is(err, store.ErrDebateInviteAlreadySent):
			respondError(w, http.StatusConflict, "DEBATE_INVITE_ALREADY_SENT", "You already sent an invite for this debate.")
		case errors.Is(err, store.ErrDebateUserBlocked):
			respondError(w, http.StatusForbidden, "DEBATE_USER_BLOCKED", "You cannot invite because one of you has blocked the other.")
		default:
			h.Logger.Error("invite debate failed", "error", err, "debate_id", debateID, "invited_user_id", invitedUserID)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"debate_id":        debateID,
		"invited_username": req.InvitedUsername,
	})
}
