package handler

import (
	"net/http"
	"time"

	"respond/internal/auth"
	"respond/internal/store"
)

func (h Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if h.Config.IsDevelopment() {
		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Email verified successfully.",
		})
		return
	}

	var req verifyEmailRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.Token == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Token is required.")
		return
	}

	tokenHash := hashToken(req.Token)
	token, err := h.Store.GetEmailVerificationTokenByHash(r.Context(), tokenHash)
	if err != nil || token.UsedAt != nil || time.Now().After(token.ExpiresAt) {
		respondError(w, http.StatusBadRequest, "AUTH_VERIFY_TOKEN_INVALID", "Verification token is invalid or expired.")
		return
	}

	if err := h.Store.UpdateUserEmailVerified(r.Context(), token.UserID, true); err != nil {
		h.Logger.Error("update email verified failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if err := h.Store.MarkEmailVerificationTokenUsed(r.Context(), token.ID); err != nil {
		h.Logger.Error("mark verification token used failed", "error", err)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Email verified successfully.",
	})
}

func (h Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}
	if h.Config.IsDevelopment() {
		if err := h.Store.UpdateUserEmailVerified(r.Context(), userID, true); err != nil {
			h.Logger.Error("mark email verified in development resend failed", "error", err, "user_id", userID)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Verification email sent.",
		})
		return
	}

	verified, err := h.Store.UserEmailVerified(r.Context(), userID)
	if err != nil {
		h.Logger.Error("email verified lookup failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if verified {
		respondError(w, http.StatusConflict, "AUTH_ALREADY_VERIFIED", "Email is already verified.")
		return
	}

	user, err := h.Store.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required.")
		return
	}

	recentCount, err := h.Store.CountEmailVerificationTokensByUserSince(r.Context(), userID, time.Now().Add(-1*time.Hour))
	if err != nil {
		h.Logger.Error("count recent verification tokens failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if recentCount >= resendVerificationLimitPerHour {
		respondError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many verification resend requests. Try again later.")
		return
	}

	if err := h.Store.DeleteEmailVerificationTokensByUserID(r.Context(), userID); err != nil {
		h.Logger.Error("delete verification tokens failed", "error", err)
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
		Email:     user.Email,
		TokenHash: hashToken(rawToken),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}); err != nil {
		h.Logger.Error("create verification token failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if h.Email == nil {
		h.Logger.Error("email service unavailable on resend verification", "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}
	if err := h.Email.QueueVerificationEmail(r.Context(), user.Email, rawToken); err != nil {
		h.Logger.Error("queue verification email failed", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Verification email sent.",
	})
}
