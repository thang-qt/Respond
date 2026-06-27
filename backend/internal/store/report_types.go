package store

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	reportTargetDebate  = "debate"
	reportTargetTurn    = "turn"
	reportTargetComment = "comment"

	reportLimitPer24h     = 10
	trustedMinAccountAge  = 180 * 24 * time.Hour
	trustedMinRating      = 1100
	trustedMaxRating      = 2200
	trustedDebateTrigger  = 5
	trustedCommentTrigger = 3
	trustedTurnTrigger    = 5
)

var (
	ErrReportTargetNotFound = errors.New("report target not found")
	ErrReportDuplicate      = errors.New("report duplicate")
	ErrReportSelfNotAllowed = errors.New("report self not allowed")
	ErrReportRateLimited    = errors.New("report rate limited")
	ErrReportNotFound       = errors.New("report not found")
	ErrReportAlreadyClosed  = errors.New("report already closed")
)

var systemUserID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

type CreateReportParams struct {
	ReporterID uuid.UUID
	TargetType string
	TargetID   uuid.UUID
	Reason     string
	Details    *string
}

type ResolveReportParams struct {
	ReportID    uuid.UUID
	ReviewerID  uuid.UUID
	Resolution  string
	Note        *string
	TargetType  string
	TargetID    uuid.UUID
	FinalStatus string
	ReviewedAt  time.Time
}

type ListReportsParams struct {
	Status     string
	TargetType string
	Page       int
	PerPage    int
}

type ListHiddenContentParams struct {
	TargetType string
	Limit      int
}
