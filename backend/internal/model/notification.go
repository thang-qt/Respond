package model

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID  `json:"id"`
	Type      string     `json:"type"`
	Message   string     `json:"message"`
	DebateID   *uuid.UUID `json:"debate_id"`
	DebateSlug *string    `json:"debate_slug"`
	TurnNumber *int       `json:"turn_number"`
	Read       bool       `json:"read"`
	CreatedAt  time.Time  `json:"created_at"`
}
