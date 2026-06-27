package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"respond/internal/auth"
	"respond/internal/store"
)

const (
	maxFollowedTags = 30
	maxBioLength    = 200
)

func (h Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	user, err := h.Store.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
			return
		}
		h.Logger.Error("get me failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, toUserResponse(user, h.Config))
}

type updateMeRequest struct {
	Bio           *string `json:"bio"`
	DefaultReveal *bool   `json:"default_reveal"`
	Locale        *string `json:"locale"`
}

func (h Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	var req updateMeRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.Bio == nil && req.DefaultReveal == nil && req.Locale == nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "No updatable fields provided.")
		return
	}

	if req.Bio != nil {
		bio := *req.Bio
		if len([]rune(bio)) > maxBioLength {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "bio must be at most 200 characters.")
			return
		}
	}

	if req.Locale != nil {
		locale := strings.TrimSpace(strings.ToLower(*req.Locale))
		if locale != "en" && locale != "vi" {
			respondErrorKey(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "error.localeInvalid", nil)
			return
		}
		req.Locale = &locale
	}

	if err := h.Store.UpdateUserProfile(r.Context(), userID, store.UpdateUserProfileParams{
		Bio:           req.Bio,
		DefaultReveal: req.DefaultReveal,
		Locale:        req.Locale,
	}); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
			return
		}
		h.Logger.Error("update me failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	user, err := h.Store.GetUserByID(r.Context(), userID)
	if err != nil {
		h.Logger.Error("get updated user failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, toUserResponse(user, h.Config))
}

type updatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h Handler) UpdateMyPassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	var req updatePasswordRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	req.CurrentPassword = strings.TrimSpace(req.CurrentPassword)
	if req.CurrentPassword == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "current_password is required.")
		return
	}
	if len(req.NewPassword) < minPasswordLength || len(req.NewPassword) > maxPasswordLength {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "new_password must be 8-128 characters.")
		return
	}

	user, err := h.Store.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
			return
		}
		h.Logger.Error("get user for password update failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, req.CurrentPassword); err != nil {
		respondError(w, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Current password is incorrect.")
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		h.Logger.Error("hash updated password failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if err := h.Store.UpdateUserPassword(r.Context(), userID, newHash); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
			return
		}
		h.Logger.Error("update password failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	// Rotate sessions: user should log back in on all devices.
	if err := h.Store.DeleteRefreshTokensByUserID(r.Context(), userID); err != nil {
		h.Logger.Error("delete refresh tokens after password update failed", "error", err)
	}
	clearRefreshCookie(w, h.Config)
	respondNoContent(w)
}

type updateEmailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h Handler) UpdateMyEmail(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	var req updateEmailRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "email and password are required.")
		return
	}

	user, err := h.Store.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
			return
		}
		h.Logger.Error("get user for email update failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
		respondError(w, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Password is incorrect.")
		return
	}

	if err := h.Store.UpdateUserEmail(r.Context(), userID, req.Email); err != nil {
		if isUniqueViolation(err, "idx_users_email") {
			respondError(w, http.StatusConflict, "AUTH_EMAIL_TAKEN", "Email already registered.")
			return
		}
		if errors.Is(err, store.ErrInvalidEmail) {
			respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid email format.")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
			return
		}
		h.Logger.Error("update email failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if h.Config.IsDevelopment() {
		if err := h.Store.UpdateUserEmailVerified(r.Context(), userID, true); err != nil {
			h.Logger.Error("mark email verified in development email update failed", "error", err, "user_id", userID)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
			return
		}

		respondJSON(w, http.StatusOK, map[string]any{
			"email":          req.Email,
			"email_verified": true,
		})
		return
	}

	if err := h.Store.DeleteEmailVerificationTokensByUserID(r.Context(), userID); err != nil {
		h.Logger.Error("delete prior verification tokens failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	rawToken, err := generateToken()
	if err != nil {
		h.Logger.Error("generate verification token failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if _, err := h.Store.CreateEmailVerificationToken(r.Context(), store.CreateEmailVerificationTokenParams{
		UserID:    userID,
		Email:     req.Email,
		TokenHash: hashToken(rawToken),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}); err != nil {
		h.Logger.Error("create verification token after email update failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if h.Email == nil {
		h.Logger.Error("email service unavailable on email update", "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if err := h.Email.QueueVerificationEmail(r.Context(), req.Email, rawToken); err != nil {
		h.Logger.Error("queue verification email after email update failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	updatedUser, err := h.Store.GetUserByID(r.Context(), userID)
	if err != nil {
		h.Logger.Error("get user after email update failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"email":          updatedUser.Email,
		"email_verified": updatedUser.EmailVerified,
	})
}
