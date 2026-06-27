package model

import (
	"time"

	"github.com/google/uuid"
)

type UserAccountStatus string

const (
	UserAccountStatusActive    UserAccountStatus = "active"
	UserAccountStatusSuspended UserAccountStatus = "suspended"
	UserAccountStatusBanned    UserAccountStatus = "banned"
)

type UserEnforcementActionType string

const (
	UserEnforcementActionWarning     UserEnforcementActionType = "warning"
	UserEnforcementActionRestriction UserEnforcementActionType = "restriction"
	UserEnforcementActionSuspension  UserEnforcementActionType = "suspension"
	UserEnforcementActionBan         UserEnforcementActionType = "ban"
	UserEnforcementActionRevoke      UserEnforcementActionType = "revoke"
)

type UserCapability string

const (
	UserCapabilityCreateDebate UserCapability = "create_debate"
	UserCapabilityComment      UserCapability = "comment"
	UserCapabilityVote         UserCapability = "vote"
	UserCapabilityFollow       UserCapability = "follow"
	UserCapabilityReport       UserCapability = "report"
	UserCapabilityInvite       UserCapability = "invite"
)

type UserEnforcementAction struct {
	ID           uuid.UUID                 `json:"id"`
	TargetUserID uuid.UUID                 `json:"target_user_id"`
	ActorUserID  uuid.UUID                 `json:"actor_user_id"`
	ActionType   UserEnforcementActionType `json:"action"`
	Capabilities []UserCapability          `json:"capabilities"`
	ExpiresAt    *time.Time                `json:"expires_at,omitempty"`
	RevokedAt    *time.Time                `json:"revoked_at,omitempty"`
	Note         string                    `json:"note"`
	CreatedAt    time.Time                 `json:"created_at"`
	CreatedBy    *ReportUserRef            `json:"created_by,omitempty"`
	Status       string                    `json:"status,omitempty"`
}

type UserEnforcementState struct {
	AccountStatus          UserAccountStatus
	RestrictedCapabilities map[UserCapability]bool
}

type AccountBlockReason struct {
	ActionType UserEnforcementActionType
	ExpiresAt  *time.Time
	Note       string
}
