package store

import (
	"database/sql"

	"github.com/google/uuid"
)

// NotifyFunc is called after a notification is persisted.
// It receives the target user ID, notification type, message,
// optional debate ID, and optional turn number.
type NotifyFunc func(userID uuid.UUID, notifType, message string, debateID *uuid.UUID, turnNumber *int)

type Store struct {
	DB       *sql.DB
	OnNotify NotifyFunc // optional; set to push live WS notifications
}

func New(db *sql.DB) *Store {
	return &Store{DB: db}
}
