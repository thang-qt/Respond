package email

import (
	"context"
	"encoding/json"
	"fmt"

	"respond/internal/config"
	"respond/internal/store"
)

const (
	TemplateVerifyEmail   = "verify_email"
	TemplateResetPassword = "password_reset"
	TemplateSignupInvite  = "signup_invite"
)

type Service struct {
	Store        *store.Store
	DisableQueue bool
}

func NewService(st *store.Store, cfg config.Config) *Service {
	return &Service{Store: st, DisableQueue: cfg.IsDevelopment()}
}

func (s *Service) QueueVerificationEmail(ctx context.Context, toEmail, token string) error {
	if s.DisableQueue {
		return nil
	}

	payload, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return fmt.Errorf("marshal verification payload: %w", err)
	}
	_, err = s.Store.UpsertPendingVerificationEmailJob(ctx, toEmail, payload)
	if err != nil {
		return fmt.Errorf("queue verification email: %w", err)
	}
	return nil
}

func (s *Service) QueuePasswordResetEmail(ctx context.Context, toEmail, token string) error {
	if s.DisableQueue {
		return nil
	}

	payload, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return fmt.Errorf("marshal reset payload: %w", err)
	}
	_, err = s.Store.CreateEmailJob(ctx, store.CreateEmailJobParams{
		ToEmail:     toEmail,
		Template:    TemplateResetPassword,
		PayloadJSON: payload,
	})
	if err != nil {
		return fmt.Errorf("queue password reset email: %w", err)
	}
	return nil
}

func (s *Service) QueueSignupInviteEmail(ctx context.Context, toEmail, token, inviterUsername string) error {
	if s.DisableQueue {
		return nil
	}

	payload, err := json.Marshal(map[string]string{
		"token":            token,
		"inviter_username": inviterUsername,
	})
	if err != nil {
		return fmt.Errorf("marshal signup invite payload: %w", err)
	}
	_, err = s.Store.CreateEmailJob(ctx, store.CreateEmailJobParams{
		ToEmail:     toEmail,
		Template:    TemplateSignupInvite,
		PayloadJSON: payload,
	})
	if err != nil {
		return fmt.Errorf("queue signup invite email: %w", err)
	}
	return nil
}
