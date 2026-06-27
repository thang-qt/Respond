package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"respond/internal/auth"
	"respond/internal/config"
	"respond/internal/model"
	"respond/internal/store"
)

func (h Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if err := validateSignup(req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	switch h.Config.SignupMode {
	case config.SignupModeClosed:
		respondError(w, http.StatusForbidden, "AUTH_SIGNUP_CLOSED", "Signup is currently disabled.")
		return
	case config.SignupModeInviteOnly:
		if strings.TrimSpace(req.InviteToken) == "" {
			respondError(w, http.StatusForbidden, "AUTH_INVITE_REQUIRED", "A valid invite is required to sign up.")
			return
		}
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.Logger.Error("hash password failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	var user model.User
	if h.Config.SignupMode == config.SignupModeInviteOnly {
		user, err = h.Store.CreateUserWithSettingsFromInvite(r.Context(), store.CreateUserWithInviteParams{
			Email:        req.Email,
			Username:     req.Username,
			PasswordHash: passwordHash,
			TokenHash:    hashToken(strings.TrimSpace(req.InviteToken)),
		})
		if err != nil {
			switch {
			case errors.Is(err, store.ErrInviteNotFound):
				respondError(w, http.StatusBadRequest, "AUTH_INVITE_INVALID", "Invite token is invalid.")
				return
			case errors.Is(err, store.ErrInviteExpired):
				respondError(w, http.StatusConflict, "AUTH_INVITE_EXPIRED", "Invite token has expired.")
				return
			case errors.Is(err, store.ErrInviteAlreadyUsed):
				respondError(w, http.StatusConflict, "AUTH_INVITE_ALREADY_USED", "Invite token has already been used.")
				return
			case errors.Is(err, store.ErrInviteEmailMismatch):
				respondError(w, http.StatusBadRequest, "AUTH_INVITE_EMAIL_MISMATCH", "Invite token email does not match signup email.")
				return
			}
		}
	} else {
		user, err = h.Store.CreateUserWithSettings(r.Context(), store.CreateUserParams{
			Email:        req.Email,
			Username:     req.Username,
			PasswordHash: passwordHash,
		})
	}
	if err != nil {
		if isUniqueViolation(err, "idx_users_email") {
			respondError(w, http.StatusConflict, "AUTH_EMAIL_TAKEN", "Email already registered.")
			return
		}
		if isUniqueViolation(err, "idx_users_username") {
			respondError(w, http.StatusConflict, "AUTH_USERNAME_TAKEN", "Username already registered.")
			return
		}
		h.Logger.Error("create user failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if h.Config.IsDevelopment() {
		if err := h.Store.UpdateUserEmailVerified(r.Context(), user.ID, true); err != nil {
			h.Logger.Error("mark email verified in development failed", "error", err, "user_id", user.ID)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
			return
		}
		user.EmailVerified = true
	} else {
		rawVerificationToken, err := generateToken()
		if err != nil {
			h.Logger.Error("generate verification token failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
			return
		}
		if _, err := h.Store.CreateEmailVerificationToken(r.Context(), store.CreateEmailVerificationTokenParams{
			UserID:    user.ID,
			Email:     user.Email,
			TokenHash: hashToken(rawVerificationToken),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}); err != nil {
			h.Logger.Error("create verification token failed", "error", err)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
			return
		}
		if h.Email == nil {
			h.Logger.Error("email service unavailable on signup")
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
			return
		}
		if err := h.Email.QueueVerificationEmail(r.Context(), user.Email, rawVerificationToken); err != nil {
			h.Logger.Error("queue verification email failed", "error", err, "user_id", user.ID)
			respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
			return
		}
	}

	accessToken, refreshToken, err := h.issueTokens(r.Context(), user)
	if err != nil {
		h.Logger.Error("issue tokens failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	setRefreshCookie(w, refreshToken, h.Config)
	respondJSON(w, http.StatusCreated, authResponse{
		User:        toUserResponse(user, h.Config),
		AccessToken: accessToken,
	})
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Email)
	}
	if identifier == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email/username and password are required.")
		return
	}

	user, err := h.Store.GetUserByEmailOrUsername(r.Context(), identifier)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondError(w, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Invalid username/email or password.")
			return
		}
		h.Logger.Error("get user by identifier failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
		respondError(w, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Invalid username/email or password.")
		return
	}

	if user.AccountStatus == model.UserAccountStatusSuspended || user.AccountStatus == model.UserAccountStatusBanned {
		h.respondAccountBlockedLoginError(w, r, user)
		return
	}

	accessToken, refreshToken, err := h.issueTokens(r.Context(), user)
	if err != nil {
		h.Logger.Error("issue tokens failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	setRefreshCookie(w, refreshToken, h.Config)
	respondJSON(w, http.StatusOK, authResponse{
		User:        toUserResponse(user, h.Config),
		AccessToken: accessToken,
	})
}

func (h Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := readRefreshCookie(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "Refresh token expired or invalid.")
		return
	}

	tokenHash := hashToken(refreshToken)
	stored, err := h.Store.GetRefreshTokenByHash(r.Context(), tokenHash)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "Refresh token expired or invalid.")
		return
	}

	if time.Now().After(stored.ExpiresAt) {
		_ = h.Store.DeleteRefreshTokenByHash(r.Context(), tokenHash)
		respondError(w, http.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "Refresh token expired or invalid.")
		return
	}

	user, err := h.Store.GetUserByID(r.Context(), stored.UserID)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "Refresh token expired or invalid.")
		return
	}
	if user.AccountStatus == model.UserAccountStatusSuspended || user.AccountStatus == model.UserAccountStatusBanned {
		_ = h.Store.DeleteRefreshTokenByHash(r.Context(), tokenHash)
		h.respondAccountBlockedLoginError(w, r, user)
		return
	}

	if err := h.Store.DeleteRefreshTokenByHash(r.Context(), tokenHash); err != nil {
		h.Logger.Error("delete refresh token failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	accessToken, newRefreshToken, err := h.issueTokens(r.Context(), user)
	if err != nil {
		h.Logger.Error("issue tokens failed", "error", err)
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Something went wrong.")
		return
	}

	setRefreshCookie(w, newRefreshToken, h.Config)
	respondJSON(w, http.StatusOK, map[string]string{"access_token": accessToken})
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := readRefreshCookie(r)
	if err == nil {
		_ = h.Store.DeleteRefreshTokenByHash(r.Context(), hashToken(refreshToken))
	}

	clearRefreshCookie(w, h.Config)
	respondNoContent(w)
}
