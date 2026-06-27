package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"respond/internal/model"
)

func TestRequireModerator(t *testing.T) {
	tests := []struct {
		name       string
		role       *model.UserRole
		wantStatus int
	}{
		{name: "missing role", role: nil, wantStatus: http.StatusUnauthorized},
		{name: "user role forbidden", role: ptrRole(model.UserRoleUser), wantStatus: http.StatusForbidden},
		{name: "moderator role allowed", role: ptrRole(model.UserRoleModerator), wantStatus: http.StatusNoContent},
		{name: "admin role allowed", role: ptrRole(model.UserRoleAdmin), wantStatus: http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
			handler := RequireModerator()(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.role != nil {
				ctx := context.WithValue(req.Context(), userRoleKey, *tt.role)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	tests := []struct {
		name       string
		role       *model.UserRole
		wantStatus int
	}{
		{name: "missing role", role: nil, wantStatus: http.StatusUnauthorized},
		{name: "user role forbidden", role: ptrRole(model.UserRoleUser), wantStatus: http.StatusForbidden},
		{name: "moderator role forbidden", role: ptrRole(model.UserRoleModerator), wantStatus: http.StatusForbidden},
		{name: "admin role allowed", role: ptrRole(model.UserRoleAdmin), wantStatus: http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
			handler := RequireAdmin()(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.role != nil {
				ctx := context.WithValue(req.Context(), userRoleKey, *tt.role)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestContextExtractors(t *testing.T) {
	ctx := context.Background()
	if _, ok := UserIDFromContext(ctx); ok {
		t.Fatal("expected no user id from empty context")
	}
	if _, ok := UserRoleFromContext(ctx); ok {
		t.Fatal("expected no user role from empty context")
	}
}

func ptrRole(role model.UserRole) *model.UserRole {
	r := role
	return &r
}
