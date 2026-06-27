package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type SignupMode string

const (
	SignupModeOpen       SignupMode = "open"
	SignupModeInviteOnly SignupMode = "invite_only"
	SignupModeClosed     SignupMode = "closed"
)

type Config struct {
	Env                   string
	Addr                  string
	DatabaseURL           string
	JWTSecret             string
	JWTAccessTTL          time.Duration
	JWTRefreshTTL         time.Duration
	FrontendURL           string
	SMTPHost              string
	SMTPPort              int
	SMTPUsername          string
	SMTPPassword          string
	SMTPFromEmail         string
	SMTPFromName          string
	SMTPRequireTLS        bool
	SignupMode            SignupMode
	InviteTokenTTL        time.Duration
	InviteMinAge          time.Duration
	InviteRequireVerified bool
}

func (c Config) IsDevelopment() bool {
	return c.Env == "development"
}

func Load() Config {
	cfg := Config{
		Env:                   getEnv("ENV", "development"),
		Addr:                  getAddr(),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		JWTSecret:             getEnv("JWT_SECRET", "dev-secret"),
		JWTAccessTTL:          getDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:         getDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		FrontendURL:           getEnv("FRONTEND_URL", "http://localhost:3000"),
		SMTPHost:              getEnv("SMTP_HOST", ""),
		SMTPPort:              getInt("SMTP_PORT", 587),
		SMTPUsername:          getEnv("SMTP_USERNAME", ""),
		SMTPPassword:          getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:         getEnv("SMTP_FROM_EMAIL", "noreply@localhost"),
		SMTPFromName:          getEnv("SMTP_FROM_NAME", "Respond"),
		SMTPRequireTLS:        getBool("SMTP_REQUIRE_TLS", true),
		SignupMode:            parseSignupMode(getEnv("AUTH_SIGNUP_MODE", string(SignupModeInviteOnly))),
		InviteTokenTTL:        getDuration("INVITE_TOKEN_TTL", 7*24*time.Hour),
		InviteMinAge:          getDuration("INVITE_MIN_ACCOUNT_AGE", 14*24*time.Hour),
		InviteRequireVerified: getBool("INVITE_REQUIRE_VERIFIED", true),
	}

	return cfg
}

func parseSignupMode(raw string) SignupMode {
	switch SignupMode(strings.ToLower(strings.TrimSpace(raw))) {
	case SignupModeOpen:
		return SignupModeOpen
	case SignupModeClosed:
		return SignupModeClosed
	default:
		return SignupModeInviteOnly
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getAddr() string {
	if v := os.Getenv("RESPOND_ADDR"); v != "" {
		return v
	}
	if v := os.Getenv("BACKEND_PORT"); v != "" {
		return ":" + v
	}
	return ":8080"
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		switch v {
		case "1", "true", "TRUE", "yes", "YES":
			return true
		case "0", "false", "FALSE", "no", "NO":
			return false
		}
	}
	return fallback
}
