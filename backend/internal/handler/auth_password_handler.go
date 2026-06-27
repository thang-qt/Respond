package handler

import (
	"net/http"
	"strings"
	"time"

	"respond/internal/auth"
	"respond/internal/store"
)

func (h Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email is required.")
		return
	}
	req.Email = strings.TrimSpace(req.Email)

	user, err := h.Store.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		rawToken, err := generateToken()
		if err == nil {
			if _, err := h.Store.CreatePasswordResetToken(r.Context(), store.CreatePasswordResetTokenParams{
				UserID:    user.ID,
				TokenHash: hashToken(rawToken),
				ExpiresAt: time.Now().Add(time.Hour),
			}); err != nil {
				h.Logger.Error("create password reset token failed", "error", err)
			} else if h.Email != nil {
				if err := h.Email.QueuePasswordResetEmail(r.Context(), user.Email, rawToken); err != nil {
					h.Logger.Error("queue password reset email failed", "error", err, "user_id", user.ID)
				}
			} else {
				h.Logger.Warn("email service unavailable for password reset email", "user_id", user.ID)
			}
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "If that email is registered, you'll receive a reset link.",
	})
}

func (h Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if len(req.NewPassword) < minPasswordLength || req.Token == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Token and new password are required.")
		return
	}

	if len(req.NewPassword) > maxPasswordLength {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Password must be at most 128 characters.")
		return
	}

	tokenHash := hashToken(req.Token)
	token, err := h.Store.GetPasswordResetTokenByHash(r.Context(), tokenHash)
	if err != nil || token.UsedAt != nil || time.Now().After(token.ExpiresAt) {
		respondError(w, http.StatusBadRequest, "AUTH_RESET_TOKEN_INVALID", "Reset token is invalid or expired.")
		return
	}

	passwordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		h.Logger.Error("hash password failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if err := h.Store.UpdateUserPassword(r.Context(), token.UserID, passwordHash); err != nil {
		h.Logger.Error("update user password failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if err := h.Store.MarkPasswordResetTokenUsed(r.Context(), token.ID); err != nil {
		h.Logger.Error("mark reset token used failed", "error", err)
	}
	_ = h.Store.DeleteRefreshTokensByUserID(r.Context(), token.UserID)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password has been reset. Please log in.",
	})
}
