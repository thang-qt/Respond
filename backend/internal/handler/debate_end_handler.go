package handler

import (
	"errors"
	"net/http"

	"respond/internal/auth"
	"respond/internal/realtime"
	"respond/internal/store"
)

func (h Handler) ConcedeDebate(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.Store.ConcedeDebate(r.Context(), debateID, userID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateNotActive):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_ACTIVE", "This debate is not currently active.")
		case errors.Is(err, store.ErrDebateNotParticipant):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_PARTICIPANT", "You are not a participant in this debate.")
		default:
			h.Logger.Error("concede debate failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	// Broadcast debate ended event.
	h.broadcastDebateEvents(debateID, result.Events)
	h.Hub.Broadcast(debateID, "debate.ended", realtime.DebateEndedData{
		Outcome:    result.Outcome,
		WinnerSide: result.WinnerSide,
		EndedAt:    result.EndedAt,
	})

	respondJSON(w, http.StatusOK, result)
}

func (h Handler) ResignDebate(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.Store.ResignDebate(r.Context(), debateID, userID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateNotActive):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_ACTIVE", "This debate is not currently active.")
		case errors.Is(err, store.ErrDebateNotParticipant):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_PARTICIPANT", "You are not a participant in this debate.")
		default:
			h.Logger.Error("resign debate failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	// Broadcast seat open event.
	h.broadcastDebateEvents(debateID, result.Events)
	if result.Status == "waiting_replacement" && result.WinnerSide != nil {
		h.Hub.Broadcast(debateID, "debate.seat_open", realtime.DebateSeatOpenData{
			Side: *result.WinnerSide, // WinnerSide holds open_side for resign results
		})
	}

	respondJSON(w, http.StatusOK, result)
}

func (h Handler) ProposeDraw(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.Store.ProposeDrawDebate(r.Context(), debateID, userID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateNotActive):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_ACTIVE", "This debate is not currently active.")
		case errors.Is(err, store.ErrDebateNotParticipant):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_PARTICIPANT", "You are not a participant in this debate.")
		case errors.Is(err, store.ErrDrawCooldown):
			respondError(w, http.StatusForbidden, "DRAW_COOLDOWN", "A draw was proposed too recently. Wait at least 3 turns.")
		default:
			h.Logger.Error("propose draw failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	// Broadcast draw proposed event.
	h.broadcastDebateEvents(debateID, result.Events)
	h.Hub.Broadcast(debateID, "debate.draw_proposed", realtime.DrawProposedData{
		ProposedBy: result.ProposedBy,
	})

	respondJSON(w, http.StatusOK, result)
}

type RespondDrawRequest struct {
	Accept bool `json:"accept"`
}

func (h Handler) RespondDraw(w http.ResponseWriter, r *http.Request) {
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

	var req RespondDrawRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	result, err := h.Store.RespondDrawDebate(r.Context(), debateID, userID, req.Accept)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateNotActive):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_ACTIVE", "This debate is not currently active.")
		case errors.Is(err, store.ErrDebateNotParticipant):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_PARTICIPANT", "You are not a participant in this debate.")
		case errors.Is(err, store.ErrDrawNotProposed):
			respondError(w, http.StatusForbidden, "DRAW_NOT_PROPOSED", "There is no active draw proposal.")
		case errors.Is(err, store.ErrDrawSelfRespond):
			respondError(w, http.StatusForbidden, "DRAW_SELF_RESPOND", "You cannot respond to your own draw proposal.")
		default:
			h.Logger.Error("respond draw failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	// Broadcast draw response. If accepted, broadcast debate ended;
	// if declined, broadcast draw responded.
	h.broadcastDebateEvents(debateID, result.Events)
	if result.Status != nil && *result.Status == "finished" {
		h.Hub.Broadcast(debateID, "debate.ended", realtime.DebateEndedData{
			Outcome:    result.Outcome,
			WinnerSide: result.WinnerSide,
			EndedAt:    result.EndedAt,
		})
	} else {
		h.Hub.Broadcast(debateID, "debate.draw_responded", realtime.DrawRespondedData{
			Accepted: req.Accept,
		})
	}

	respondJSON(w, http.StatusOK, result)
}

func (h Handler) ReplaceDebate(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.Store.ReplaceDebate(r.Context(), debateID, userID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateNotWaitingReplacement):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_ACTIVE", "This debate is not open for replacement.")
		case errors.Is(err, store.ErrDebateIsRemainingDebater):
			respondError(w, http.StatusForbidden, "DEBATE_OWN_DEBATE", "You are already a participant in this debate.")
		case errors.Is(err, store.ErrDebateIsResignedDebater):
			respondError(w, http.StatusForbidden, "DEBATE_OWN_DEBATE", "You cannot rejoin a debate you resigned from.")
		case errors.Is(err, store.ErrDebateUserBlocked):
			respondError(w, http.StatusForbidden, "DEBATE_USER_BLOCKED", "You cannot replace because one of you has blocked the other.")
		default:
			h.Logger.Error("replace debate failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	// Broadcast replacement event.
	h.broadcastDebateEvents(debateID, result.Events)
	h.Hub.Broadcast(debateID, "debate.replacement", realtime.DebateReplacementData{
		Side:         result.Side,
		AnonymousID:  result.AnonymousID,
		TurnDeadline: result.TurnDeadline,
	})

	respondJSON(w, http.StatusOK, result)
}

type RevealDebateRequest struct {
	Reveal *bool `json:"reveal"`
}

func (h Handler) RevealDebateIdentity(w http.ResponseWriter, r *http.Request) {
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

	var req RevealDebateRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if req.Reveal == nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body.")
		return
	}

	result, err := h.Store.RevealDebateIdentity(r.Context(), debateID, userID, *req.Reveal)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateNotFinished):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_FINISHED", "Debate hasn't ended.")
		case errors.Is(err, store.ErrDebateNotParticipant):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_PARTICIPANT", "You are not a participant in this debate.")
		case errors.Is(err, store.ErrRevealAlreadyChosen):
			respondError(w, http.StatusConflict, "REVEAL_ALREADY_CHOSEN", "You already made a reveal choice.")
		default:
			h.Logger.Error("reveal identity failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

type RespondExtensionRequest struct {
	Accept bool `json:"accept"`
}

func (h Handler) RespondExtension(w http.ResponseWriter, r *http.Request) {
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

	var req RespondExtensionRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	result, err := h.Store.RespondExtension(r.Context(), debateID, userID, req.Accept)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateNotPendingExtension):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_PENDING_EXTENSION", "This debate is not awaiting an extension decision.")
		case errors.Is(err, store.ErrDebateNotParticipant):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_PARTICIPANT", "You are not a participant in this debate.")
		case errors.Is(err, store.ErrExtensionAlreadyResponded):
			respondError(w, http.StatusConflict, "EXTENSION_ALREADY_RESPONDED", "You already responded to this extension.")
		default:
			h.Logger.Error("respond extension failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	// Broadcast extension update (either extended or ended as draw).
	h.broadcastDebateEvents(debateID, result.Events)
	h.Hub.Broadcast(debateID, "debate.extension_update", realtime.ExtensionUpdateData{
		Status:     result.Status,
		TurnLimit:  result.TurnLimit,
		Outcome:    result.Outcome,
		WinnerSide: result.WinnerSide,
	})

	respondJSON(w, http.StatusOK, result)
}
