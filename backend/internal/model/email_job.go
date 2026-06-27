package model

import (
	"time"

	"github.com/google/uuid"
)

type EmailJobStatus string

const (
	EmailJobStatusPending    EmailJobStatus = "pending"
	EmailJobStatusProcessing EmailJobStatus = "processing"
	EmailJobStatusSent       EmailJobStatus = "sent"
	EmailJobStatusFailed     EmailJobStatus = "failed"
)

type EmailJob struct {
	ID            uuid.UUID      `json:"id"`
	ToEmail       string         `json:"to_email"`
	Template      string         `json:"template"`
	PayloadJSON   []byte         `json:"-"`
	Status        EmailJobStatus `json:"status"`
	Attempts      int            `json:"attempts"`
	MaxAttempts   int            `json:"max_attempts"`
	NextAttemptAt time.Time      `json:"next_attempt_at"`
	LastError     *string        `json:"last_error"`
	SentAt        *time.Time     `json:"sent_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
