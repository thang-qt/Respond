package handler

import (
	"log/slog"

	"respond/internal/config"
	"respond/internal/email"
	"respond/internal/realtime"
	"respond/internal/store"
)

type Handler struct {
	Config config.Config
	Store  *store.Store
	Logger *slog.Logger
	Hub    *realtime.Hub
	Email  *email.Service
}

func New(cfg config.Config, st *store.Store, logger *slog.Logger, hub *realtime.Hub, emailSvc *email.Service) Handler {
	return Handler{
		Config: cfg,
		Store:  st,
		Logger: logger,
		Hub:    hub,
		Email:  emailSvc,
	}
}
