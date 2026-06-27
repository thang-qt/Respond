package model

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleUser      UserRole = "user"
	UserRoleModerator UserRole = "moderator"
	UserRoleAdmin     UserRole = "admin"
)

type User struct {
	ID                uuid.UUID         `json:"id"`
	Email             string            `json:"email"`
	EmailVerified     bool              `json:"email_verified"`
	Role              UserRole          `json:"role"`
	AccountStatus     UserAccountStatus `json:"account_status"`
	Username          string            `json:"username"`
	PasswordHash      string            `json:"-"`
	Bio               string            `json:"bio"`
	Rating            int               `json:"rating"`
	Wins              int               `json:"wins"`
	Losses            int               `json:"losses"`
	Draws             int               `json:"draws"`
	DefaultReveal     bool              `json:"default_reveal"`
	Locale            string            `json:"locale"`
	UsernameChangedAt *time.Time        `json:"username_changed_at,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}
