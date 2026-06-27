package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"respond/internal/config"
	"respond/internal/model"
	"respond/internal/store"
)

type contextKey string

const userIDKey contextKey = "user_id"
const userRoleKey contextKey = "user_role"

// RequireAuth validates the JWT access token and injects the user ID into the
// request context. It confirms the user exists to honor deletes/bans.
func RequireAuth(st *store.Store, cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				respondUnauthorized(w)
				return
			}

			tokenString := strings.TrimPrefix(header, "Bearer ")
			parsed, err := ParseToken(tokenString, cfg.JWTSecret)
			if err != nil || !parsed.Valid {
				respondUnauthorized(w)
				return
			}

			claims, ok := parsed.Claims.(jwt.MapClaims)
			if !ok {
				respondUnauthorized(w)
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok {
				respondUnauthorized(w)
				return
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				respondUnauthorized(w)
				return
			}

			user, err := st.GetUserByID(r.Context(), userID)
			if err != nil {
				respondUnauthorized(w)
				return
			}
			if user.AccountStatus == model.UserAccountStatusSuspended {
				respondSuspended(w)
				return
			}
			if user.AccountStatus == model.UserAccountStatusBanned {
				respondBanned(w)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, userRoleKey, user.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

func UserRoleFromContext(ctx context.Context) (model.UserRole, bool) {
	role, ok := ctx.Value(userRoleKey).(model.UserRole)
	return role, ok
}

// RequireVerified ensures the authenticated user has a verified email.
// This middleware must run after RequireAuth.
func RequireVerified(st *store.Store, cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.IsDevelopment() {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				respondUnauthorized(w)
				return
			}

			verified, err := st.UserEmailVerified(r.Context(), userID)
			if err != nil {
				respondUnauthorized(w)
				return
			}
			if !verified {
				respondEmailNotVerified(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth parses the JWT if present and injects user ID into context.
// It never rejects the request; invalid tokens are ignored.
func OptionalAuth(st *store.Store, cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := strings.TrimPrefix(header, "Bearer ")
			parsed, err := ParseToken(tokenString, cfg.JWTSecret)
			if err != nil || !parsed.Valid {
				next.ServeHTTP(w, r)
				return
			}

			claims, ok := parsed.Claims.(jwt.MapClaims)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := st.GetUserByID(r.Context(), userID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if user.AccountStatus == model.UserAccountStatusSuspended || user.AccountStatus == model.UserAccountStatusBanned {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, userRoleKey, user.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireModerator allows moderator and admin roles.
// This middleware must run after RequireAuth.
func RequireModerator() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := UserRoleFromContext(r.Context())
			if !ok {
				respondUnauthorized(w)
				return
			}
			if role != model.UserRoleModerator && role != model.UserRoleAdmin {
				respondForbiddenRole(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin allows admin role only.
// This middleware must run after RequireAuth.
func RequireAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := UserRoleFromContext(r.Context())
			if !ok {
				respondUnauthorized(w)
				return
			}
			if role != model.UserRoleAdmin {
				respondForbiddenRole(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func respondUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    "AUTH_UNAUTHORIZED",
			"message": "Authentication required.",
		},
	})
}

func respondEmailNotVerified(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    "AUTH_EMAIL_NOT_VERIFIED",
			"message": "Verify your email to perform this action.",
		},
	})
}

func respondForbiddenRole(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    "AUTH_FORBIDDEN_ROLE",
			"message": "You do not have permission to perform this action.",
		},
	})
}

func respondSuspended(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    "USER_SUSPENDED",
			"message": "This account is temporarily suspended.",
		},
	})
}

func respondBanned(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    "USER_BANNED",
			"message": "This account is permanently banned.",
		},
	})
}
