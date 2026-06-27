package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret"
	userID := "8f3d44b1-7640-4e69-8f70-3d5d5d409630"

	tokenString, err := GenerateAccessToken(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	parsed, err := ParseToken(tokenString, secret)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if !parsed.Valid {
		t.Fatal("expected parsed token to be valid")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected jwt.MapClaims")
	}
	if got := claims["sub"]; got != userID {
		t.Fatalf("sub claim = %v, want %s", got, userID)
	}
}

func TestParseTokenWrongSecret(t *testing.T) {
	tokenString, err := GenerateAccessToken("user-id", "right-secret", time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	parsed, err := ParseToken(tokenString, "wrong-secret")
	if err == nil {
		t.Fatal("expected parse error for wrong secret")
	}
	if parsed != nil && parsed.Valid {
		t.Fatal("token should not be valid with wrong secret")
	}
}

func TestParseTokenExpired(t *testing.T) {
	tokenString, err := GenerateAccessToken("user-id", "test-secret", -1*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	parsed, err := ParseToken(tokenString, "test-secret")
	if err == nil {
		t.Fatal("expected parse error for expired token")
	}
	if parsed != nil && parsed.Valid {
		t.Fatal("expired token should not be valid")
	}
}
