package handler

import (
	"regexp"
	"time"
)

const (
	minPasswordLength              = 8
	maxPasswordLength              = 128
	minUsernameLength              = 5
	maxUsernameLength              = 20
	resendVerificationLimitPerHour = 3
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type signupRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	InviteToken string `json:"invite_token"`
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

type authResponse struct {
	User        userResponse `json:"user"`
	AccessToken string       `json:"access_token"`
}

type userResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Role          string    `json:"role"`
	AccountStatus string    `json:"account_status"`
	Username      string    `json:"username"`
	Bio           string    `json:"bio"`
	Rating        int       `json:"rating"`
	Wins          int       `json:"wins"`
	Losses        int       `json:"losses"`
	Draws         int       `json:"draws"`
	DefaultReveal bool      `json:"default_reveal"`
	Locale        string    `json:"locale"`
	CreatedAt     time.Time `json:"created_at"`
}
