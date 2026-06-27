package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"respond/internal/auth"
	"respond/internal/model"
	"respond/internal/store"
)

type CreateCommentRequest struct {
	Content      string     `json:"content"`
	ParentID     *uuid.UUID `json:"parent_id"`
	IsReflection *bool      `json:"is_reflection"`
}

type UpdateCommentRequest struct {
	Content string `json:"content"`
}

func (h Handler) ListDebateComments(w http.ResponseWriter, r *http.Request) {
	debateID, ok := h.resolveDebateID(w, r)
	if !ok {
		return
	}
	if !h.ensureDebateVisible(w, r, debateID) {
		return
	}

	status, hidden, err := h.Store.GetDebateStatus(r.Context(), debateID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
			return
		}
		h.Logger.Error("get debate status failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if status != "finished" {
		respondError(w, http.StatusForbidden, "DEBATE_NOT_FINISHED", "Discussion opens after the debate ends.")
		return
	}

	canViewHiddenContent, err := h.canViewHiddenDebateContent(r, debateID)
	if err != nil {
		h.Logger.Error("check hidden comment visibility failed", "error", err, "debate_id", debateID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if hidden && !canViewHiddenContent {
		respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		return
	}

	sort := r.URL.Query().Get("sort")
	page, perPage, ok := parsePageParams(w, r, 20, 50)
	if !ok {
		return
	}

	var viewerID *uuid.UUID
	if id, ok := auth.UserIDFromContext(r.Context()); ok {
		viewerID = &id
	}

	comments, total, err := h.Store.ListDebateComments(r.Context(), store.ListCommentsParams{
		DebateID:             debateID,
		Sort:                 sort,
		Page:                 page,
		PerPage:              perPage,
		ViewerID:             viewerID,
		CanViewHiddenContent: canViewHiddenContent,
	})
	if err != nil {
		if errors.Is(err, store.ErrInvalidSort) {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid sort.")
			return
		}
		h.Logger.Error("list comments failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondList(w, http.StatusOK, comments, newListMeta(page, perPage, total))
}

func (h Handler) CreateDebateComment(w http.ResponseWriter, r *http.Request) {
	debateID, ok := h.resolveDebateID(w, r)
	if !ok {
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	capability := commentCapability()
	if !h.ensureUserCanPerform(w, r, userID, &capability) {
		return
	}
	if !h.ensureDebateVisible(w, r, debateID) {
		return
	}
	if !h.ensureDebateMutable(w, r, debateID) {
		return
	}

	var req CreateCommentRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	content := strings.TrimSpace(req.Content)
	length := utf8.RuneCountInString(content)
	if length < 1 || length > 2000 {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Comment must be 1 to 2000 characters.")
		return
	}

	isReflection := false
	if req.IsReflection != nil {
		isReflection = *req.IsReflection
	}

	comment, err := h.Store.CreateComment(r.Context(), store.CreateCommentParams{
		DebateID:     debateID,
		UserID:       userID,
		Content:      content,
		ParentID:     req.ParentID,
		IsReflection: isReflection,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			respondError(w, http.StatusNotFound, "DEBATE_NOT_FOUND", "This debate doesn't exist or has been removed.")
		case errors.Is(err, store.ErrDebateNotFinished):
			respondError(w, http.StatusForbidden, "DEBATE_NOT_FINISHED", "Discussion opens after the debate ends.")
		case errors.Is(err, store.ErrCommentThreadLocked):
			respondError(w, http.StatusForbidden, "COMMENT_THREAD_LOCKED", "Discussion thread is locked.")
		case errors.Is(err, store.ErrCommentParentNotFound):
			respondError(w, http.StatusNotFound, "COMMENT_PARENT_NOT_FOUND", "Comment parent not found.")
		case errors.Is(err, store.ErrCommentNestedReply):
			respondError(w, http.StatusBadRequest, "COMMENT_NESTED_REPLY", "Replies cannot be nested.")
		case errors.Is(err, store.ErrReflectionNotParticipant):
			respondError(w, http.StatusForbidden, "COMMENT_REFLECTION_NOT_PARTICIPANT", "Only participants can post reflections.")
		case errors.Is(err, store.ErrReflectionExists), isUniqueViolation(err, "idx_comments_reflection_unique"):
			respondError(w, http.StatusConflict, "COMMENT_REFLECTION_EXISTS", "Reflection already posted.")
		default:
			h.Logger.Error("create comment failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusCreated, comment)
}

func (h Handler) UpdateDebateComment(w http.ResponseWriter, r *http.Request) {
	debateID, ok := h.resolveDebateID(w, r)
	if !ok {
		return
	}
	if !h.ensureDebateVisible(w, r, debateID) {
		return
	}
	if !h.ensureDebateMutable(w, r, debateID) {
		return
	}

	commentParam := chi.URLParam(r, "comment_id")
	commentID, err := uuid.Parse(commentParam)
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid comment id.")
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	capability := commentCapability()
	if !h.ensureUserCanPerform(w, r, userID, &capability) {
		return
	}

	var req UpdateCommentRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	content := strings.TrimSpace(req.Content)
	length := utf8.RuneCountInString(content)
	if length < 1 || length > 2000 {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Comment must be 1 to 2000 characters.")
		return
	}

	comment, err := h.Store.UpdateComment(r.Context(), store.UpdateCommentParams{
		DebateID:  debateID,
		CommentID: commentID,
		UserID:    userID,
		Content:   content,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrCommentNotFound):
			respondError(w, http.StatusNotFound, "COMMENT_NOT_FOUND", "Comment not found.")
		case errors.Is(err, store.ErrCommentNotAuthor):
			respondError(w, http.StatusForbidden, "COMMENT_NOT_AUTHOR", "You can only edit your own comments.")
		case errors.Is(err, store.ErrCommentEditExpired):
			respondError(w, http.StatusForbidden, "COMMENT_EDIT_EXPIRED", "Edit window expired.")
		default:
			h.Logger.Error("update comment failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":         comment.ID.String(),
		"content":    comment.Content,
		"updated_at": comment.UpdatedAt,
	})
}

func (h Handler) DeleteDebateComment(w http.ResponseWriter, r *http.Request) {
	debateID, ok := h.resolveDebateID(w, r)
	if !ok {
		return
	}
	if !h.ensureDebateVisible(w, r, debateID) {
		return
	}
	if !h.ensureDebateMutable(w, r, debateID) {
		return
	}

	commentParam := chi.URLParam(r, "comment_id")
	commentID, err := uuid.Parse(commentParam)
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid comment id.")
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "You must be signed in.")
		return
	}
	capability := commentCapability()
	if !h.ensureUserCanPerform(w, r, userID, &capability) {
		return
	}

	err = h.Store.DeleteComment(r.Context(), store.DeleteCommentParams{
		DebateID:  debateID,
		CommentID: commentID,
		UserID:    userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrCommentNotFound):
			respondError(w, http.StatusNotFound, "COMMENT_NOT_FOUND", "Comment not found.")
		case errors.Is(err, store.ErrCommentNotAuthor):
			respondError(w, http.StatusForbidden, "COMMENT_NOT_AUTHOR", "You can only delete your own comments.")
		default:
			h.Logger.Error("delete comment failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func commentCapability() model.UserCapability {
	return model.UserCapabilityComment
}
