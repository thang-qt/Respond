package background

import (
	"context"
	"log/slog"
	"time"

	"respond/internal/email"
	"respond/internal/realtime"
	"respond/internal/store"
)

// StartJobs launches all periodic background tasks as goroutines.
// The provided context controls shutdown: when cancelled, all jobs stop.
func StartJobs(ctx context.Context, st *store.Store, hub *realtime.Hub, logger *slog.Logger, emailWorker *email.Worker) {
	// Walkovers: active debate where turn deadline + turn window elapsed (2x grace)
	go runEvery(ctx, 5*time.Minute, "CheckWalkovers", logger, func() {
		CheckWalkovers(ctx, st, hub, logger)
	})

	// Turn expiry nudge: 75% of turn window elapsed, send one notification
	go runEvery(ctx, 5*time.Minute, "CheckTurnExpiry", logger, func() {
		CheckTurnExpiry(ctx, st, logger)
	})

	// Waiting debate expiration: waiting for 14+ days → expired
	go runEvery(ctx, 1*time.Hour, "CheckExpirations", logger, func() {
		CheckExpirations(ctx, st, logger)
	})

	// Pending challenge expiration: waiting invite-only challenge past deadline.
	go runEvery(ctx, 1*time.Hour, "CheckChallengeExpirations", logger, func() {
		CheckChallengeExpirations(ctx, st, logger)
	})

	// Replacement expiry: waiting_replacement for 7+ days → walkover
	go runEvery(ctx, 1*time.Hour, "CheckReplacementExpiry", logger, func() {
		CheckReplacementExpiry(ctx, st, hub, logger)
	})

	// Extension expiry: pending_extension past deadline → draw
	go runEvery(ctx, 5*time.Minute, "CheckExtensionExpiry", logger, func() {
		CheckExtensionExpiry(ctx, st, hub, logger)
	})

	// Expired tokens: clean up expired refresh/reset/verification tokens
	go runEvery(ctx, 1*time.Hour, "CleanExpiredTokens", logger, func() {
		if err := st.DeleteExpiredTokens(ctx); err != nil {
			logger.Error("expired token cleanup failed", "error", err)
		}
	})

	// Timer sync: broadcast current turn side + deadline to active debate rooms
	go runEvery(ctx, 1*time.Minute, "TimerSync", logger, func() {
		BroadcastTimerSync(ctx, st, hub, logger)
	})

	if emailWorker != nil {
		// Email queue processing: claim ready jobs and send via SMTP.
		go runEvery(ctx, 15*time.Second, "ProcessEmailQueue", logger, func() {
			emailWorker.Process(ctx)
		})
	}

	logger.Info("background jobs started")
}

// runEvery runs fn on a fixed interval until ctx is cancelled.
func runEvery(ctx context.Context, interval time.Duration, name string, logger *slog.Logger, fn func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logger.Debug("running background job", "job", name)
			fn()
		case <-ctx.Done():
			logger.Info("stopping background job", "job", name)
			return
		}
	}
}
