package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"respond/internal/auth"
	"respond/internal/i18n"
	"respond/internal/realtime"
	"respond/internal/store"
)

func (h Handler) JoinDebate(w http.ResponseWriter, r *http.Request) {
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
	if activeCount >= 5 {
		respondError(w, http.StatusForbidden, "DEBATE_LIMIT_REACHED", "You already have 5 active debates.")
		return
	}

	result, err := h.Store.JoinDebate(r.Context(), debateID, userID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateChallengeOnly):
			respondError(w, http.StatusForbidden, "DEBATE_CHALLENGE_ONLY", "This debate is invite-only. Use challenge response instead.")
		case errors.Is(err, store.ErrDebateNotWaiting):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_ACTIVE", "This debate is not open for joining.")
		case errors.Is(err, store.ErrDebateOwnDebate):
			respondError(w, http.StatusForbidden, "DEBATE_OWN_DEBATE", "You can't join your own debate.")
		case errors.Is(err, store.ErrDebateFull):
			respondError(w, http.StatusConflict, "DEBATE_FULL", "This debate already has two sides.")
		case errors.Is(err, store.ErrDebateUserBlocked):
			respondError(w, http.StatusForbidden, "DEBATE_USER_BLOCKED", "You cannot join because one of you has blocked the other.")
		default:
			h.Logger.Error("join debate failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	// Broadcast join event to all connected clients.
	h.Hub.Broadcast(debateID, "debate.joined", realtime.DebateJoinedData{
		Side:         result.Side,
		AnonymousID:  result.AnonymousID,
		TurnDeadline: result.TurnDeadline,
	})

	respondJSON(w, http.StatusOK, result)
}

func (h Handler) GetDebate(w http.ResponseWriter, r *http.Request) {
	debateID, ok := h.resolveDebateID(w, r)
	if !ok {
		return
	}
	if !h.ensureDebateVisible(w, r, debateID) {
		return
	}

	canViewHiddenContent, err := h.canViewHiddenDebateContent(r, debateID)
	if err != nil {
		h.Logger.Error("get hidden debate visibility failed", "error", err, "debate_id", debateID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	debate, err := h.Store.GetDebateByIDLocalized(r.Context(), debateID, canViewHiddenContent, i18n.LocaleFromRequest(r))
	if err != nil {
		if err == store.ErrNotFound {
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
			return
		}
		h.Logger.Error("get debate failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if debate.Hidden && !canViewHiddenContent {
		respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		return
	}

	if viewerID := h.currentViewerID(r); viewerID != nil {
		viewer, err := h.Store.GetDebateViewer(r.Context(), debateID, *viewerID)
		if err == nil {
			debate.Viewer = &viewer
		}
	}

	respondJSON(w, http.StatusOK, debate)
}

type SubmitTurnRequest struct {
	Content    string  `json:"content"`
	AIAssisted bool    `json:"ai_assisted"`
	AINote     *string `json:"ai_note"`
}

func (h Handler) SubmitTurn(w http.ResponseWriter, r *http.Request) {
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

	var req SubmitTurnRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.AINote != nil {
		trimmed := strings.TrimSpace(*req.AINote)
		if trimmed == "" {
			req.AINote = nil
		} else {
			req.AINote = &trimmed
		}
	}

	length := utf8.RuneCountInString(req.Content)
	if length < minOpeningTurnLength {
		respondError(w, http.StatusBadRequest, "TURN_TOO_SHORT", "Turn must be at least 100 characters.")
		return
	}
	if length > maxOpeningTurnLength {
		respondError(w, http.StatusBadRequest, "TURN_TOO_LONG", "Turn must be at most 5,000 characters.")
		return
	}
	if req.AINote != nil {
		if !req.AIAssisted {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ai_note requires ai_assisted=true.")
			return
		}
		if noteLen := utf8.RuneCountInString(*req.AINote); noteLen > maxAINoteLength {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "AI note must be at most 300 characters.")
			return
		}
	}

	turn, err := h.Store.SubmitTurn(r.Context(), store.SubmitTurnParams{
		DebateID:   debateID,
		UserID:     userID,
		Content:    req.Content,
		AIAssisted: req.AIAssisted,
		AINote:     req.AINote,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateNotActive):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_ACTIVE", "This debate is not currently active.")
		case errors.Is(err, store.ErrTurnNotYourTurn):
			respondError(w, http.StatusForbidden, "TURN_NOT_YOUR_TURN", "It's not your turn.")
		default:
			h.Logger.Error("submit turn failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	// Broadcast new turn to all clients watching this debate.
	h.Hub.Broadcast(debateID, "turn.new", turn)

	respondJSON(w, http.StatusCreated, turn)
}
