package model

import "time"

type NotificationSettings struct {
	EmailYourTurn     bool      `json:"email_your_turn"`
	EmailDebateJoined bool      `json:"email_debate_joined"`
	EmailDebateEnded  bool      `json:"email_debate_ended"`
	EmailTurnExpiring bool      `json:"email_turn_expiring"`
	EmailSeatOpen     bool      `json:"email_seat_open"`
	EmailDrawProposed bool      `json:"email_draw_proposed"`
	UpdatedAt         time.Time `json:"updated_at"`
}
