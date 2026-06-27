package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/model"
	"respond/internal/store"
)

type CommentVoteResponse struct {
	CommentID   string `json:"comment_id"`
	Voted       bool   `json:"voted"`
	UpvoteCount int    `json:"upvote_count"`
}

type DebateVoteResponse struct {
	DebateID    string `json:"debate_id"`
	Voted       bool   `json:"voted"`
	UpvoteCount int    `json:"upvote_count"`
}

func (h Handler) ToggleDebateVote(w http.ResponseWriter, r *http.Request) {
	debateID, ok := h.resolveDebateID(w, r)
	if !ok {
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	voteCapability := model.UserCapabilityVote
	if !h.ensureUserCanPerform(w, r, userID, &voteCapability) {
		return
	}
	if !h.ensureDebateVisible(w, r, debateID) {
		return
	}
	if !h.ensureDebateMutable(w, r, debateID) {
		return
	}

	result, err := h.Store.ToggleDebateVote(r.Context(), debateID, userID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrDebateNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateHiddenByBlock):
			respondError(w, http.StatusForbidden, debateHiddenByBlockCode, "This debate is hidden by your safety settings.")
		default:
			h.Logger.Error("toggle debate vote failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, DebateVoteResponse{
		DebateID:    result.DebateID.String(),
		Voted:       result.Voted,
		UpvoteCount: result.UpvoteCount,
	})
}

func (h Handler) ToggleCommentVote(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	commentID, err := uuid.Parse(idParam)
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid comment id.")
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	voteCapability := model.UserCapabilityVote
	if !h.ensureUserCanPerform(w, r, userID, &voteCapability) {
		return
	}

	result, err := h.Store.ToggleCommentVote(r.Context(), commentID, userID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrCommentNotFound):
			respondError(w, http.StatusNotFound, "COMMENT_NOT_FOUND", "Comment not found.")
		case errors.Is(err, store.ErrDebateHiddenByBlock):
			respondError(w, http.StatusForbidden, debateHiddenByBlockCode, "This debate is hidden by your safety settings.")
		default:
			h.Logger.Error("toggle comment vote failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, CommentVoteResponse{
		CommentID:   result.CommentID.String(),
		Voted:       result.Voted,
		UpvoteCount: result.UpvoteCount,
	})
}
