package store

import (
	"context"
	"fmt"
)

func (s *Store) ListActiveDebateTimers(ctx context.Context) ([]ActiveDebateTimer, error) {
	const query = `
		SELECT id, current_turn_side, turn_deadline
		FROM debates
		WHERE status = 'active'
		  AND turn_deadline IS NOT NULL
	`
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list active debate timers: %w", err)
	}
	defer rows.Close()

	var timers []ActiveDebateTimer
	for rows.Next() {
		var t ActiveDebateTimer
		if err := rows.Scan(&t.DebateID, &t.CurrentTurnSide, &t.TurnDeadline); err != nil {
			return nil, fmt.Errorf("scan active debate timer: %w", err)
		}
		timers = append(timers, t)
	}
	return timers, rows.Err()
}
