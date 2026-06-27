package model

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID               uuid.UUID   `json:"id"`
	DebateID         uuid.UUID   `json:"debate_id"`
	ParentID         *uuid.UUID  `json:"parent_id"`
	User             UserSummary `json:"user"`
	Content          string      `json:"content"`
	IsReflection     bool        `json:"is_reflection"`
	IsDebater        bool        `json:"is_debater"`
	DebaterSide      *string     `json:"debater_side"`
	DebaterAnonymous *string     `json:"debater_anonymous_id"`
	Hidden           bool        `json:"hidden"`
	UpvoteCount      int         `json:"upvote_count"`
	ViewerHasUpvoted bool        `json:"viewer_has_upvoted"`
	IsAuthor         bool        `json:"is_author"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        *time.Time  `json:"updated_at"`
	Replies          []Comment   `json:"replies"`
}
