package background

import (
	"respond/internal/realtime"
	"respond/internal/store"
)

func broadcastDebateUpdates(hub *realtime.Hub, updates []store.BackgroundDebateUpdate) {
	for _, update := range updates {
		hub.Broadcast(update.DebateID, realtime.EventDebateEvent, update.Event)
		if update.Outcome != nil && update.EndedAt != nil {
			hub.Broadcast(update.DebateID, realtime.EventDebateEnded, realtime.DebateEndedData{
				Outcome:    update.Outcome,
				WinnerSide: update.WinnerSide,
				EndedAt:    update.EndedAt,
			})
		}
	}
}
