package store

import (
	"fmt"
	"strings"
	"time"

	"respond/internal/model"
)

func moderationAuditActionForUserEnforcement(actionType model.UserEnforcementActionType) string {
	switch actionType {
	case model.UserEnforcementActionWarning:
		return "warn_user"
	case model.UserEnforcementActionRestriction:
		return "restrict_user"
	case model.UserEnforcementActionSuspension:
		return "suspend_user"
	case model.UserEnforcementActionBan:
		return "ban_user"
	default:
		return "warn_user"
	}
}

func parseUserCapabilities(values []string) []model.UserCapability {
	if len(values) == 0 {
		return []model.UserCapability{}
	}
	out := make([]model.UserCapability, 0, len(values))
	for _, value := range values {
		out = append(out, model.UserCapability(value))
	}
	return out
}

func userEnforcementNotificationMessage(actionType model.UserEnforcementActionType, capabilities []model.UserCapability, expiresAt *time.Time, note string) string {
	message := fmt.Sprintf("A moderator issued an account %s action.", actionType)
	if actionType == model.UserEnforcementActionRestriction && len(capabilities) > 0 {
		names := make([]string, 0, len(capabilities))
		for _, capability := range capabilities {
			names = append(names, string(capability))
		}
		message = fmt.Sprintf("A moderator restricted your account actions: %s.", strings.Join(names, ", "))
	}
	if expiresAt != nil {
		message += fmt.Sprintf(" Effective until %s.", expiresAt.UTC().Format(time.RFC3339))
	}
	message += fmt.Sprintf(" Moderator note: %s", note)
	return message
}

func userEnforcementStatus(item model.UserEnforcementAction, now time.Time) string {
	if item.RevokedAt != nil {
		return "revoked"
	}
	if item.ExpiresAt != nil && !item.ExpiresAt.After(now) {
		return "expired"
	}
	return "active"
}

func validateUserEnforcementActionInput(actionType model.UserEnforcementActionType, capabilities []model.UserCapability, expiresAt *time.Time) error {
	switch actionType {
	case model.UserEnforcementActionWarning, model.UserEnforcementActionRestriction, model.UserEnforcementActionSuspension, model.UserEnforcementActionBan:
	default:
		return ErrUserEnforcementActionInvalid
	}

	for _, capability := range capabilities {
		switch capability {
		case model.UserCapabilityCreateDebate,
			model.UserCapabilityComment,
			model.UserCapabilityVote,
			model.UserCapabilityFollow,
			model.UserCapabilityReport,
			model.UserCapabilityInvite:
		default:
			return ErrUserEnforcementCapabilityInvalid
		}
	}

	if actionType == model.UserEnforcementActionRestriction {
		if len(capabilities) == 0 {
			return ErrUserEnforcementCapabilityInvalid
		}
	} else if len(capabilities) > 0 {
		return ErrUserEnforcementCapabilityInvalid
	}

	if actionType == model.UserEnforcementActionBan || actionType == model.UserEnforcementActionWarning {
		if expiresAt != nil {
			return ErrUserEnforcementActionInvalid
		}
	}

	if expiresAt != nil && !expiresAt.After(time.Now()) {
		return ErrUserEnforcementActionInvalid
	}

	return nil
}
