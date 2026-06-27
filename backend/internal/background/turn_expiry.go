package background

import (
	"context"
	"log/slog"

	"respond/internal/store"
)

// CheckTurnExpiry sends nudge notifications for debates where 75% of the
// turn window has elapsed but the deadline hasn't passed yet.
func CheckTurnExpiry(ctx context.Context, st *store.Store, logger *slog.Logger) {
	st.ProcessTurnExpiry(ctx, logger)
}
