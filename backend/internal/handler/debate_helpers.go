package handler

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"respond/internal/model"
	"respond/internal/realtime"
)

func (h Handler) broadcastDebateEvents(debateID uuid.UUID, events []model.DebateEvent) {
	for _, event := range events {
		h.Hub.Broadcast(debateID, realtime.EventDebateEvent, event)
	}
}

const (
	minTopicLength       = 10
	maxTopicLength       = 200
	maxContextLength     = 500
	minOpeningTurnLength = 100
	maxOpeningTurnLength = 5000
	maxAINoteLength      = 300
)

type CreateDebateRequest struct {
	Topic                 string   `json:"topic"`
	TagIDs                []string `json:"tag_ids"`
	TimeMode              string   `json:"time_mode"`
	TurnLimit             int      `json:"turn_limit"`
	Context               *string  `json:"context"`
	OpeningTurn           string   `json:"opening_turn"`
	OpeningTurnAIAssisted bool     `json:"opening_turn_ai_assisted"`
	OpeningTurnAINote     *string  `json:"opening_turn_ai_note"`
}

type CreateChallengeRequest struct {
	CreateDebateRequest
	InvitedUsername string `json:"invited_username"`
}

type CreateRechallengeRequest struct {
	CreateDebateRequest
}

func (h Handler) validateCreateDebateRequest(ctx context.Context, req *CreateDebateRequest) ([]uuid.UUID, string, error) {
	req.Topic = strings.TrimSpace(req.Topic)
	req.OpeningTurn = strings.TrimSpace(req.OpeningTurn)
	if req.Context != nil {
		trimmed := strings.TrimSpace(*req.Context)
		req.Context = &trimmed
		if trimmed == "" {
			req.Context = nil
		}
	}
	if req.OpeningTurnAINote != nil {
		trimmed := strings.TrimSpace(*req.OpeningTurnAINote)
		if trimmed == "" {
			req.OpeningTurnAINote = nil
		} else {
			req.OpeningTurnAINote = &trimmed
		}
	}

	if length := utf8.RuneCountInString(req.Topic); length < minTopicLength || length > maxTopicLength {
		return nil, "Topic must be 10–200 characters.", nil
	}

	if req.Context != nil {
		if length := utf8.RuneCountInString(*req.Context); length > maxContextLength {
			return nil, "Context must be at most 500 characters.", nil
		}
	}

	if length := utf8.RuneCountInString(req.OpeningTurn); length < minOpeningTurnLength || length > maxOpeningTurnLength {
		return nil, "Opening argument must be 100–5,000 characters.", nil
	}

	if req.OpeningTurnAINote != nil {
		if !req.OpeningTurnAIAssisted {
			return nil, "opening_turn_ai_note requires opening_turn_ai_assisted=true.", nil
		}
		if length := utf8.RuneCountInString(*req.OpeningTurnAINote); length > maxAINoteLength {
			return nil, "AI note must be at most 300 characters.", nil
		}
	}

	if len(req.TagIDs) < 1 || len(req.TagIDs) > 3 {
		return nil, "tag_ids must contain 1-3 tags.", nil
	}
	parsedTagIDs := make([]uuid.UUID, 0, len(req.TagIDs))
	seenTagIDs := make(map[uuid.UUID]struct{}, len(req.TagIDs))
	for _, rawTagID := range req.TagIDs {
		tagID, err := uuid.Parse(strings.TrimSpace(rawTagID))
		if err != nil {
			return nil, "Invalid tag_ids.", nil
		}
		if _, exists := seenTagIDs[tagID]; exists {
			return nil, "tag_ids must be unique.", nil
		}
		seenTagIDs[tagID] = struct{}{}
		parsedTagIDs = append(parsedTagIDs, tagID)
	}
	tagCount, err := h.Store.CountTagsByIDs(ctx, parsedTagIDs)
	if err != nil {
		return nil, "", err
	}
	if tagCount != len(parsedTagIDs) {
		return nil, "Invalid tag_ids.", nil
	}

	switch req.TimeMode {
	case "marathon", "standard", "rapid", "blitz":
	default:
		return nil, "Invalid time mode.", nil
	}

	switch req.TurnLimit {
	case 10, 20, 30, 40:
	default:
		return nil, "Invalid turn limit.", nil
	}

	return parsedTagIDs, "", nil
}
