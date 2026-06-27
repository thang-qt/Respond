package background

import (
	"context"
	"log/slog"

	"respond/internal/store"
)

// CheckExpirations expires waiting debates that have been open for 14+ days
// with no challenger.
func CheckExpirations(ctx context.Context, st *store.Store, logger *slog.Logger) {
	st.ProcessExpirations(ctx, logger)
}

// CheckChallengeExpirations expires waiting invite-only challenges after 7 days.
func CheckChallengeExpirations(ctx context.Context, st *store.Store, logger *slog.Logger) {
	st.ProcessChallengeExpirations(ctx, logger)
}
