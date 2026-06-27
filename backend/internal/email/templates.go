package email

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type TemplateData struct {
	Token           string `json:"token"`
	InviterUsername string `json:"inviter_username"`
}

func Render(templateName string, payload []byte, frontendURL string) (subject string, body string, err error) {
	var data TemplateData
	if err := json.Unmarshal(payload, &data); err != nil {
		return "", "", fmt.Errorf("unmarshal email payload: %w", err)
	}
	if data.Token == "" {
		return "", "", errors.New("email payload missing token")
	}

	base := strings.TrimRight(frontendURL, "/")
	verifyURL := base + "/verify-email?token=" + data.Token
	resetURL := base + "/reset-password?token=" + data.Token
	signupURL := base + "/auth/signup?invite=" + data.Token

	switch templateName {
	case TemplateVerifyEmail:
		return "Verify your email", fmt.Sprintf(
			"Welcome to Respond.\n\nVerify your email using this link:\n%s\n\nIf your client cannot open links, use this token:\n%s\n\nThis verification token expires in 24 hours.",
			verifyURL,
			data.Token,
		), nil
	case TemplateResetPassword:
		return "Reset your password", fmt.Sprintf(
			"We received a request to reset your Respond password.\n\nReset link:\n%s\n\nIf your client cannot open links, use this token:\n%s\n\nThis reset token expires in 1 hour. If you did not request this, you can ignore this email.",
			resetURL,
			data.Token,
		), nil
	case TemplateSignupInvite:
		inviter := data.InviterUsername
		if inviter == "" {
			inviter = "A Respond user"
		}
		return "You are invited to join Respond", fmt.Sprintf(
			"%s invited you to join Respond.\n\nCreate your account using this invite link:\n%s\n\nIf your client cannot open links, use this invite token on signup:\n%s\n\nThis invite token is single-use and expires in 7 days.",
			inviter,
			signupURL,
			data.Token,
		), nil
	default:
		return "", "", fmt.Errorf("unknown email template: %s", templateName)
	}
}
