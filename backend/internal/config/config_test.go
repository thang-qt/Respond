package config

import (
	"testing"
	"time"
)

func TestGetAddrPrecedence(t *testing.T) {
	t.Setenv("RESPOND_ADDR", ":9090")
	t.Setenv("BACKEND_PORT", "8088")
	if got := getAddr(); got != ":9090" {
		t.Fatalf("getAddr() = %s, want :9090", got)
	}

	t.Setenv("RESPOND_ADDR", "")
	t.Setenv("BACKEND_PORT", "8088")
	if got := getAddr(); got != ":8088" {
		t.Fatalf("getAddr() = %s, want :8088", got)
	}

	t.Setenv("RESPOND_ADDR", "")
	t.Setenv("BACKEND_PORT", "")
	if got := getAddr(); got != ":8080" {
		t.Fatalf("getAddr() = %s, want :8080", got)
	}
}

func TestGetDurationFallbackOnInvalidValue(t *testing.T) {
	t.Setenv("JWT_ACCESS_TTL", "not-a-duration")
	fallback := 15 * time.Minute
	if got := getDuration("JWT_ACCESS_TTL", fallback); got != fallback {
		t.Fatalf("getDuration() = %s, want fallback %s", got, fallback)
	}
}

func TestGetBoolParsing(t *testing.T) {
	t.Setenv("SMTP_REQUIRE_TLS", "true")
	if got := getBool("SMTP_REQUIRE_TLS", false); !got {
		t.Fatal("expected true for value true")
	}

	t.Setenv("SMTP_REQUIRE_TLS", "0")
	if got := getBool("SMTP_REQUIRE_TLS", true); got {
		t.Fatal("expected false for value 0")
	}

	t.Setenv("SMTP_REQUIRE_TLS", "unexpected")
	if got := getBool("SMTP_REQUIRE_TLS", true); !got {
		t.Fatal("expected fallback true for invalid value")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("ENV", "")
	t.Setenv("RESPOND_ADDR", "")
	t.Setenv("BACKEND_PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_ACCESS_TTL", "")
	t.Setenv("JWT_REFRESH_TTL", "")
	t.Setenv("FRONTEND_URL", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM_EMAIL", "")
	t.Setenv("SMTP_FROM_NAME", "")
	t.Setenv("SMTP_REQUIRE_TLS", "")
	t.Setenv("AUTH_SIGNUP_MODE", "")
	t.Setenv("INVITE_TOKEN_TTL", "")
	t.Setenv("INVITE_MIN_ACCOUNT_AGE", "")
	t.Setenv("INVITE_REQUIRE_VERIFIED", "")

	cfg := Load()
	if cfg.Env != "development" {
		t.Fatalf("Env = %s, want development", cfg.Env)
	}
	if cfg.Addr != ":8080" {
		t.Fatalf("Addr = %s, want :8080", cfg.Addr)
	}
	if cfg.JWTSecret != "dev-secret" {
		t.Fatalf("JWTSecret = %s, want dev-secret", cfg.JWTSecret)
	}
	if cfg.SMTPPort != 587 {
		t.Fatalf("SMTPPort = %d, want 587", cfg.SMTPPort)
	}
	if !cfg.SMTPRequireTLS {
		t.Fatal("SMTPRequireTLS should default to true")
	}
	if cfg.SignupMode != SignupModeInviteOnly {
		t.Fatalf("SignupMode = %s, want invite_only", cfg.SignupMode)
	}
	if cfg.InviteTokenTTL != 7*24*time.Hour {
		t.Fatalf("InviteTokenTTL = %s, want 168h", cfg.InviteTokenTTL)
	}
	if cfg.InviteMinAge != 14*24*time.Hour {
		t.Fatalf("InviteMinAge = %s, want 336h", cfg.InviteMinAge)
	}
	if !cfg.InviteRequireVerified {
		t.Fatal("InviteRequireVerified should default to true")
	}
}

func TestParseSignupMode(t *testing.T) {
	if got := parseSignupMode("open"); got != SignupModeOpen {
		t.Fatalf("parseSignupMode(open) = %s, want open", got)
	}
	if got := parseSignupMode("closed"); got != SignupModeClosed {
		t.Fatalf("parseSignupMode(closed) = %s, want closed", got)
	}
	if got := parseSignupMode("invite_only"); got != SignupModeInviteOnly {
		t.Fatalf("parseSignupMode(invite_only) = %s, want invite_only", got)
	}
	if got := parseSignupMode("unexpected"); got != SignupModeInviteOnly {
		t.Fatalf("parseSignupMode(unexpected) = %s, want invite_only", got)
	}
}
