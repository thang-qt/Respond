package model

import (
	"time"

	"github.com/google/uuid"
)

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusRevoked  InviteStatus = "revoked"
	InviteStatusExpired  InviteStatus = "expired"
)

type Invite struct {
	ID               uuid.UUID    `json:"id"`
	InviterUserID    uuid.UUID    `json:"inviter_user_id"`
	InvitedEmail     string       `json:"invited_email"`
	TokenHash        string       `json:"-"`
	Status           InviteStatus `json:"status"`
	ExpiresAt        time.Time    `json:"expires_at"`
	AcceptedByUserID *uuid.UUID   `json:"accepted_by_user_id,omitempty"`
	AcceptedAt       *time.Time   `json:"accepted_at,omitempty"`
	RevokedAt        *time.Time   `json:"revoked_at,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}
