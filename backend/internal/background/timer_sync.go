package background

import (
	"context"
	"log/slog"

	"respond/internal/realtime"
	"respond/internal/store"
)

// BroadcastTimerSync queries all active debates and broadcasts a
// timer.sync event to each debate's WebSocket room. This keeps
// spectators and participants in sync with the server clock.
func BroadcastTimerSync(ctx context.Context, st *store.Store, hub *realtime.Hub, logger *slog.Logger) {
	timers, err := st.ListActiveDebateTimers(ctx)
	if err != nil {
		logger.Error("timer sync: list active debates failed", "error", err)
		return
	}

	for _, t := range timers {
		// Only broadcast to rooms that actually have connected clients.
		if hub.RoomSize(t.DebateID) == 0 {
			continue
		}

		hub.Broadcast(t.DebateID, realtime.EventTimerSync, realtime.TimerSyncData{
			CurrentTurnSide: t.CurrentTurnSide,
			TurnDeadline:    t.TurnDeadline,
		})
	}

	if len(timers) > 0 {
		logger.Debug("timer sync broadcast", "active_debates", len(timers))
	}
}
