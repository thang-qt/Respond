package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID   uuid.UUID `json:"id"`
	Slug string    `json:"slug"`
	Name string    `json:"name"`
}

type UserSummary struct {
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

type DebateSide struct {
	AnonymousID *string      `json:"anonymous_id"`
	Revealed    bool         `json:"revealed"`
	User        *UserSummary `json:"user"`
}

type DebateFeedItem struct {
	ID                       uuid.UUID  `json:"id"`
	Slug                     string     `json:"slug"`
	Topic                    string     `json:"topic"`
	IsChallenge              bool       `json:"is_challenge"`
	InvitedUsername          *string    `json:"invited_username,omitempty"`
	ChallengeIdentityVisible bool       `json:"challenge_identity_visible"`
	Tags                     []Tag      `json:"tags"`
	TimeMode                 string     `json:"time_mode"`
	TurnLimit                int        `json:"turn_limit"`
	Context                  *string    `json:"context"`
	LatestTurn               *TurnBrief `json:"latest_turn"`
	Status                   string     `json:"status"`
	SideA                    DebateSide `json:"side_a"`
	SideB                    DebateSide `json:"side_b"`
	Outcome                  *string    `json:"outcome"`
	WinnerSide               *string    `json:"winner_side"`
	CurrentTurnSide          string     `json:"current_turn_side"`
	TurnCount                int        `json:"turn_count"`
	TurnDeadline             *time.Time `json:"turn_deadline"`
	DrawProposedBy           *string    `json:"draw_proposed_by"`
	OpenSide                 *string    `json:"open_side"`
	ExtensionDeadline        *time.Time `json:"extension_deadline"`
	ExtensionAAccepted       *bool      `json:"extension_a_accepted"`
	ExtensionBAccepted       *bool      `json:"extension_b_accepted"`
	UpvoteCount              int        `json:"upvote_count"`
	ViewerHasUpvoted         bool       `json:"viewer_has_upvoted"`
	IsFollowing              bool       `json:"is_following"`
	ViewerIsParticipant      bool       `json:"viewer_is_participant"`
	SpectatorCount           int        `json:"spectator_count"`
	CommentCount             int        `json:"comment_count"`
	IsDailyPrompt            bool       `json:"is_daily_prompt"`
	CreatedAt                time.Time  `json:"created_at"`
	StartedAt                *time.Time `json:"started_at"`
	EndedAt                  *time.Time `json:"ended_at"`
	SideARatingDelta         *int       `json:"side_a_rating_delta"`
	SideBRatingDelta         *int       `json:"side_b_rating_delta"`
	ModerationPending        bool       `json:"moderation_pending"`
	Hidden                   bool       `json:"hidden"`
}

type TurnBrief struct {
	TurnNumber  int       `json:"turn_number"`
	Side        string    `json:"side"`
	AnonymousID string    `json:"anonymous_id"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}

type DebateTurn struct {
	ID          uuid.UUID `json:"id"`
	TurnNumber  int       `json:"turn_number"`
	Side        string    `json:"side"`
	AnonymousID string    `json:"anonymous_id"`
	Content     string    `json:"content"`
	Hidden      bool      `json:"hidden"`
	AIAssisted  bool      `json:"ai_assisted"`
	AINote      *string   `json:"ai_note"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

type DebateEvent struct {
	ID          uuid.UUID       `json:"id"`
	EventType   string          `json:"event_type"`
	Side        *string         `json:"side"`
	ActorUserID *uuid.UUID      `json:"-"`
	Payload     json.RawMessage `json:"payload_json"`
	CreatedAt   time.Time       `json:"created_at"`
}

type DebateSeatStint struct {
	ID                uuid.UUID  `json:"id"`
	Side              string     `json:"side"`
	UserID            uuid.UUID  `json:"-"`
	AnonymousID       string     `json:"anonymous_id"`
	StintIndex        int        `json:"stint_index"`
	JoinedAt          time.Time  `json:"joined_at"`
	LeftAt            *time.Time `json:"left_at"`
	LeftReason        *string    `json:"left_reason"`
	ReplacedByStintID *uuid.UUID `json:"replaced_by_stint_id"`
}

type DebateTimelineItem struct {
	Type      string       `json:"type"`
	CreatedAt time.Time    `json:"created_at"`
	Turn      *DebateTurn  `json:"turn,omitempty"`
	Event     *DebateEvent `json:"event,omitempty"`
}

type DebateParticipantHistory struct {
	SideA []DebateSeatStint `json:"side_a"`
	SideB []DebateSeatStint `json:"side_b"`
}

type DebateViewer struct {
	IsParticipant bool    `json:"is_participant"`
	Side          *string `json:"side"`
	HasUpvoted    bool    `json:"has_upvoted"`
	IsFollowing   bool    `json:"is_following"`
	RevealChoice  *bool   `json:"reveal_choice"`
}

type DebateDetail struct {
	DebateFeedItem
	Turns              []DebateTurn             `json:"turns"`
	Timeline           []DebateTimelineItem     `json:"timeline"`
	ParticipantHistory DebateParticipantHistory `json:"participant_history"`
	Viewer             *DebateViewer            `json:"viewer"`
}
