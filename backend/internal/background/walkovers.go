package background

import (
	"context"
	"log/slog"

	"respond/internal/realtime"
	"respond/internal/store"
)

// CheckWalkovers processes active debates where the turn holder has
// exceeded 2x the turn window (walkover).
func CheckWalkovers(ctx context.Context, st *store.Store, hub *realtime.Hub, logger *slog.Logger) {
	updates := st.ProcessWalkovers(ctx, logger)
	broadcastDebateUpdates(hub, updates)
}
