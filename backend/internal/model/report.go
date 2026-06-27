package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID            uuid.UUID      `json:"id"`
	TargetType    string         `json:"target_type"`
	TargetID      uuid.UUID      `json:"target_id"`
	Reason        string         `json:"reason"`
	Details       *string        `json:"details,omitempty"`
	Status        string         `json:"status"`
	Resolution    *string        `json:"resolution,omitempty"`
	TrustedReport bool           `json:"trusted_report"`
	CreatedAt     time.Time      `json:"created_at"`
	Reporter      *ReportUserRef `json:"reporter,omitempty"`
	TargetAuthor  *ReportUserRef `json:"target_author,omitempty"`
	DebateID      *uuid.UUID     `json:"debate_id,omitempty"`
	DebateSlug    *string        `json:"debate_slug,omitempty"`
	TurnNumber    *int           `json:"turn_number,omitempty"`
}

type ReportDetail struct {
	Report
	ReviewedByUserID *uuid.UUID   `json:"reviewed_by_user_id,omitempty"`
	ReviewedAt       *time.Time   `json:"reviewed_at,omitempty"`
	ResolutionNote   *string      `json:"resolution_note,omitempty"`
	Target           ReportTarget `json:"target"`
}

type ReportTarget struct {
	Hidden  bool   `json:"hidden"`
	Content string `json:"content"`
}

type ReportUserRef struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type ModerationAction struct {
	ID          uuid.UUID       `json:"id"`
	ActorUserID uuid.UUID       `json:"actor_user_id"`
	ActionType  string          `json:"action_type"`
	TargetType  string          `json:"target_type"`
	TargetID    uuid.UUID       `json:"target_id"`
	ReportID    *uuid.UUID      `json:"report_id,omitempty"`
	Payload     json.RawMessage `json:"payload_json,omitempty"`
	Reason      *string         `json:"reason,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type HiddenContentItem struct {
	TargetType   string         `json:"target_type"`
	TargetID     uuid.UUID      `json:"target_id"`
	DebateID     *uuid.UUID     `json:"debate_id,omitempty"`
	DebateSlug   *string        `json:"debate_slug,omitempty"`
	TurnNumber   *int           `json:"turn_number,omitempty"`
	TargetAuthor *ReportUserRef `json:"target_author,omitempty"`
	Content      string         `json:"content"`
	HiddenAt     *time.Time     `json:"hidden_at,omitempty"`
}
