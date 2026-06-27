package background

import (
	"context"
	"log/slog"

	"respond/internal/realtime"
	"respond/internal/store"
)

// CheckReplacementExpiry ends debates in waiting_replacement status
// where the 7-day replacement window has elapsed as walkovers.
func CheckReplacementExpiry(ctx context.Context, st *store.Store, hub *realtime.Hub, logger *slog.Logger) {
	updates := st.ProcessReplacementExpiry(ctx, logger)
	broadcastDebateUpdates(hub, updates)
}
