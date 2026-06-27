package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"net/mail"
	"time"

	"github.com/lib/pq"

	"respond/internal/auth"
	"respond/internal/config"
	"respond/internal/model"
	"respond/internal/store"
)

func (h Handler) issueTokens(ctx context.Context, user model.User) (string, string, error) {
	accessToken, err := auth.GenerateAccessToken(user.ID.String(), h.Config.JWTSecret, h.Config.JWTAccessTTL)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateToken()
	if err != nil {
		return "", "", err
	}

	_, err = h.Store.CreateRefreshToken(ctx, store.CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: time.Now().Add(h.Config.JWTRefreshTTL),
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func validateSignup(req signupRequest) error {
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("Invalid email format.")
	}
	if len(req.Username) < minUsernameLength || len(req.Username) > maxUsernameLength {
		return errors.New("Username must be 5–20 characters.")
	}
	if !usernamePattern.MatchString(req.Username) {
		return errors.New("Username can only contain letters, numbers, and underscores.")
	}
	if len(req.Password) < minPasswordLength {
		return errors.New("Password must be at least 8 characters.")
	}
	if len(req.Password) > maxPasswordLength {
		return errors.New("Password must be at most 128 characters.")
	}
	return nil
}

func toUserResponse(user model.User, cfg config.Config) userResponse {
	emailVerified := user.EmailVerified
	if cfg.IsDevelopment() {
		emailVerified = true
	}

	return userResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		EmailVerified: emailVerified,
		Role:          string(user.Role),
		AccountStatus: string(user.AccountStatus),
		Username:      user.Username,
		Bio:           user.Bio,
		Rating:        user.Rating,
		Wins:          user.Wins,
		Losses:        user.Losses,
		Draws:         user.Draws,
		DefaultReveal: user.DefaultReveal,
		Locale:        user.Locale,
		CreatedAt:     user.CreatedAt,
	}
}

func (h Handler) respondAccountBlockedLoginError(w http.ResponseWriter, r *http.Request, user model.User) {
	reason, err := h.Store.GetCurrentAccountBlockReason(r.Context(), user.ID)
	if err != nil && !errors.Is(err, store.ErrUserEnforcementActionNotFound) {
		h.Logger.Error("load account block reason failed", "error", err, "user_id", user.ID)
	}

	if user.AccountStatus == model.UserAccountStatusBanned {
		message := "Your account is permanently banned."
		if reason.Note != "" {
			message += " Moderator note: " + reason.Note
		}
		respondError(w, http.StatusForbidden, "USER_BANNED", message)
		return
	}

	message := "Your account is temporarily suspended."
	if reason.ExpiresAt != nil {
		message += " Suspension ends at " + reason.ExpiresAt.UTC().Format(time.RFC3339) + "."
	}
	if reason.Note != "" {
		message += " Moderator note: " + reason.Note
	}
	respondError(w, http.StatusForbidden, "USER_SUSPENDED", message)
}

func setRefreshCookie(w http.ResponseWriter, token string, cfg config.Config) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   cfg.Env == "production",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(cfg.JWTRefreshTTL.Seconds()),
	})
}

func clearRefreshCookie(w http.ResponseWriter, cfg config.Config) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   cfg.Env == "production",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

func readRefreshCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		return "", errors.New("missing refresh token")
	}
	return cookie.Value, nil
}

func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func isUniqueViolation(err error, constraint string) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && pqErr.Constraint == constraint
	}
	return false
}
