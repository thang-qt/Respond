DROP INDEX IF EXISTS idx_user_enforcement_actions_active;
DROP INDEX IF EXISTS idx_user_enforcement_actions_target_created;
DROP TABLE IF EXISTS user_enforcement_actions;

DROP TYPE IF EXISTS user_capability;
DROP TYPE IF EXISTS user_enforcement_action_type;

DROP INDEX IF EXISTS idx_users_account_status;
ALTER TABLE users DROP COLUMN IF EXISTS account_status;
DROP TYPE IF EXISTS account_status;

-- Cannot remove enum values from moderation_action_type and notification_type.
