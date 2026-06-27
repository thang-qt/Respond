package background

import (
	"context"
	"log/slog"

	"respond/internal/realtime"
	"respond/internal/store"
)

// CheckExtensionExpiry ends debates in pending_extension status
// where the extension_deadline has passed as draws.
func CheckExtensionExpiry(ctx context.Context, st *store.Store, hub *realtime.Hub, logger *slog.Logger) {
	updates := st.ProcessExtensionExpiry(ctx, logger)
	broadcastDebateUpdates(hub, updates)
}
