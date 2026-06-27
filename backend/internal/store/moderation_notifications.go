package store

import (
	"fmt"
)

func moderationNotificationMessage(resolution, targetType string, turnNumber *int, debateTopic *string, note *string) string {
	targetLabel := "your content"

	switch targetType {
	case "turn":
		targetLabel = "your turn"
		if turnNumber != nil {
			targetLabel = fmt.Sprintf("your turn #%d", *turnNumber)
		}
	case "comment":
		targetLabel = "your comment"
	case "debate":
		targetLabel = "your debate"
	}

	debateContext := ""
	if debateTopic != nil && *debateTopic != "" {
		debateContext = fmt.Sprintf(" in \"%s\"", *debateTopic)
	}

	message := ""
	if resolution == "hide" {
		message = fmt.Sprintf("A moderator hid %s%s.", targetLabel, debateContext)
	} else {
		message = fmt.Sprintf("A moderator restored %s%s.", targetLabel, debateContext)
	}

	if note != nil && *note != "" {
		message += fmt.Sprintf(" Moderator note: %s", *note)
	}

	return message
}
