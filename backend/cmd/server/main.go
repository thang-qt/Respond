package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"respond/internal/auth"
	"respond/internal/background"
	"respond/internal/config"
	"respond/internal/email"
	"respond/internal/handler"
	internalMiddleware "respond/internal/middleware"
	"respond/internal/realtime"
	"respond/internal/store"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if cfg.DatabaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	if cfg.Env == "production" && (cfg.JWTSecret == "" || cfg.JWTSecret == "dev-secret") {
		logger.Error("JWT_SECRET must be set to a strong value in production")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("db open failed", "error", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		logger.Error("db ping failed", "error", err)
		os.Exit(1)
	}

	st := store.New(db)
	hub := realtime.NewHub(logger)
	emailSvc := email.NewService(st, cfg)

	var emailWorker *email.Worker
	smtpMailer, err := email.NewSMTPMailer(email.SMTPConfig{
		Host:       cfg.SMTPHost,
		Port:       cfg.SMTPPort,
		Username:   cfg.SMTPUsername,
		Password:   cfg.SMTPPassword,
		FromEmail:  cfg.SMTPFromEmail,
		FromName:   cfg.SMTPFromName,
		RequireTLS: cfg.SMTPRequireTLS,
	})
	if err != nil {
		logger.Warn("smtp mailer disabled", "error", err)
	} else if cfg.IsDevelopment() {
		logger.Info("smtp worker disabled in development")
	} else {
		emailWorker = email.NewWorker(st, smtpMailer, logger, cfg.FrontendURL)
	}

	// Wire live notification push: whenever a notification is persisted,
	// push it to the user's WebSocket channel (if connected).
	st.OnNotify = func(userID uuid.UUID, notifType, message string, debateID *uuid.UUID, turnNumber *int) {
		var debateIDStr *string
		var debateSlug *string
		if debateID != nil {
			s := debateID.String()
			debateIDStr = &s

			// Include slug in WS payload so frontend can route to /debate/{slug}.
			var slug string
			if err := db.QueryRowContext(context.Background(), `SELECT slug FROM debates WHERE id = $1`, *debateID).Scan(&slug); err == nil && slug != "" {
				debateSlug = &slug
			}
		}
		hub.NotifyUser(userID, realtime.EventNotificationNew, realtime.NotificationNewData{
			Type:       notifType,
			Message:    message,
			DebateID:   debateIDStr,
			DebateSlug: debateSlug,
			TurnNumber: turnNumber,
		})
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(internalMiddleware.Logging)
	r.Use(internalMiddleware.CORS(cfg))
	r.Use(internalMiddleware.RateLimit)

	h := handler.New(cfg, st, logger, hub, emailSvc)

	r.Get("/health", h.Health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", h.Signup)
			r.Post("/login", h.Login)
			r.Post("/refresh", h.Refresh)
			r.Post("/logout", h.Logout)
			r.Post("/forgot-password", h.ForgotPassword)
			r.Post("/reset-password", h.ResetPassword)
			r.Post("/verify-email", h.VerifyEmail)
			r.With(auth.RequireAuth(st, cfg)).Post("/resend-verification", h.ResendVerification)
		})

		r.Route("/users", func(r chi.Router) {
			r.With(auth.RequireAuth(st, cfg)).Get("/me", h.GetMe)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/me/invites", h.CreateMyInvite)
			r.With(auth.RequireAuth(st, cfg)).Get("/me/invites", h.ListMyInvites)
			r.With(auth.RequireAuth(st, cfg)).Post("/me/invites/{id}/revoke", h.RevokeMyInvite)
			r.With(auth.RequireAuth(st, cfg)).Get("/me/challenges", h.ListMyChallenges)
			r.With(auth.RequireAuth(st, cfg)).Put("/me", h.UpdateMe)
			r.With(auth.RequireAuth(st, cfg)).Put("/me/password", h.UpdateMyPassword)
			r.With(auth.RequireAuth(st, cfg)).Put("/me/email", h.UpdateMyEmail)
			r.With(auth.RequireAuth(st, cfg)).Get("/me/settings/notifications", h.GetMyNotificationSettings)
			r.With(auth.RequireAuth(st, cfg)).Put("/me/settings/notifications", h.UpdateMyNotificationSettings)
			r.With(auth.RequireAuth(st, cfg)).Get("/me/blocks", h.ListMyBlockedUsers)
			r.With(auth.RequireAuth(st, cfg)).Get("/me/tag-follows", h.ListMyTagFollows)
			r.With(auth.RequireAuth(st, cfg)).Put("/me/tag-follows", h.ReplaceMyTagFollows)
			r.With(auth.RequireAuth(st, cfg)).Get("/me/notifications", h.ListNotifications)
			r.With(auth.RequireAuth(st, cfg)).Put("/me/notifications/read-all", h.MarkAllNotificationsRead)
			r.With(auth.RequireAuth(st, cfg)).Put("/me/notifications/{id}/read", h.MarkNotificationRead)
			r.With(auth.RequireAuth(st, cfg)).Get("/me/lobby", h.GetMyLobbyEntry)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Put("/me/lobby", h.UpsertMyLobbyEntry)
			r.With(auth.RequireAuth(st, cfg)).Delete("/me/lobby", h.DeleteMyLobbyEntry)
			r.With(auth.OptionalAuth(st, cfg)).Get("/search", h.ListUsersSearch)
			r.With(auth.RequireAuth(st, cfg)).Post("/{username}/block", h.BlockUser)
			r.With(auth.RequireAuth(st, cfg)).Delete("/{username}/block", h.UnblockUser)
			r.With(auth.OptionalAuth(st, cfg)).Get("/{username}", h.GetUserProfile)
			r.With(auth.OptionalAuth(st, cfg)).Get("/{username}/debates", h.ListUserDebates)
			r.With(auth.OptionalAuth(st, cfg)).Get("/{username}/lobby", h.GetUserLobbyEntry)
		})

		r.Get("/tags/search", h.ListTagsSearch)
		r.Get("/tags", h.ListTags)
		r.With(auth.OptionalAuth(st, cfg)).Get("/lobby/challenges", h.ListLobbyEntries)
		r.With(auth.OptionalAuth(st, cfg)).Get("/explore", h.ListExplore)
		r.With(auth.OptionalAuth(st, cfg)).Get("/explore/users", h.ListExploreUsers)

		r.Route("/debates", func(r chi.Router) {
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/", h.CreateDebate)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/challenges", h.CreateChallengeDebate)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/rechallenge", h.CreateRechallengeDebate)
			r.With(auth.OptionalAuth(st, cfg)).Get("/", h.ListDebates)
			r.With(auth.OptionalAuth(st, cfg)).Get("/search", h.ListDebatesSearch)
			r.With(auth.OptionalAuth(st, cfg)).Get("/{id}", h.GetDebate)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/challenge/respond", h.RespondChallenge)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/invites", h.InviteDebate)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/join", h.JoinDebate)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/turns", h.SubmitTurn)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/concede", h.ConcedeDebate)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/resign", h.ResignDebate)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/replace", h.ReplaceDebate)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/draw/propose", h.ProposeDraw)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/draw/respond", h.RespondDraw)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/reveal", h.RevealDebateIdentity)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/extend/respond", h.RespondExtension)
			r.With(auth.OptionalAuth(st, cfg)).Get("/{id}/comments", h.ListDebateComments)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/comments", h.CreateDebateComment)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Put("/{id}/comments/{comment_id}", h.UpdateDebateComment)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Delete("/{id}/comments/{comment_id}", h.DeleteDebateComment)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/vote", h.ToggleDebateVote)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/follow", h.FollowDebate)
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Delete("/{id}/follow", h.UnfollowDebate)
		})

		r.Route("/comments", func(r chi.Router) {
			r.With(auth.RequireAuth(st, cfg), auth.RequireVerified(st, cfg)).Post("/{id}/vote", h.ToggleCommentVote)
		})

		r.Route("/reports", func(r chi.Router) {
			r.With(auth.RequireAuth(st, cfg)).Post("/", h.CreateReport)
		})

		r.Route("/admin", func(r chi.Router) {
			r.With(auth.RequireAuth(st, cfg), auth.RequireModerator()).Get("/reports", h.ListAdminReports)
			r.With(auth.RequireAuth(st, cfg), auth.RequireModerator()).Get("/reports/{id}", h.GetAdminReport)
			r.With(auth.RequireAuth(st, cfg), auth.RequireModerator()).Post("/reports/{id}/resolve", h.ResolveAdminReport)
			r.With(auth.RequireAuth(st, cfg), auth.RequireModerator()).Get("/content/hidden", h.ListAdminHiddenContent)
			r.With(auth.RequireAuth(st, cfg), auth.RequireModerator()).Post("/content/{target_type}/{target_id}/hide", h.HideAdminContent)
			r.With(auth.RequireAuth(st, cfg), auth.RequireModerator()).Post("/content/{target_type}/{target_id}/restore", h.DirectRestoreAdminContent)
			r.With(auth.RequireAuth(st, cfg), auth.RequireModerator()).Get("/users/{id}/enforcement-actions", h.ListAdminUserEnforcementActions)
			r.With(auth.RequireAuth(st, cfg), auth.RequireModerator()).Post("/users/{id}/enforcement-actions", h.CreateAdminUserEnforcementAction)
			r.With(auth.RequireAuth(st, cfg), auth.RequireModerator()).Post("/users/{id}/enforcement-actions/{action_id}/revoke", h.RevokeAdminUserEnforcementAction)
			r.With(auth.RequireAuth(st, cfg), auth.RequireModerator()).Get("/users/{id}/invite-lineage", h.GetAdminInviteLineage)
			r.With(auth.RequireAuth(st, cfg), auth.RequireAdmin()).Post("/users/{id}/role", h.UpdateAdminUserRole)
		})
	})

	// WebSocket routes (outside /api/v1).
	r.Get("/ws/debates/{id}", h.WSDebate)
	r.Get("/ws/notifications", h.WSNotifications)

	// Background jobs: walkovers, turn expiry nudges, debate expirations,
	// replacement expiry, extension expiry, and expired token cleanup.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	background.StartJobs(ctx, st, hub, logger, emailWorker)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	// Graceful shutdown on SIGINT / SIGTERM.
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info("shutting down server")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown error", "error", err)
		}
		cancel() // stop background goroutines
	}()

	logger.Info("server started", "addr", cfg.Addr, "env", cfg.Env)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}
