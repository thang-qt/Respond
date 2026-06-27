DROP INDEX IF EXISTS idx_moderation_actions_actor;
DROP INDEX IF EXISTS idx_moderation_actions_target;
DROP TABLE IF EXISTS moderation_actions;

DROP INDEX IF EXISTS idx_reports_status_created;
DROP INDEX IF EXISTS idx_reports_target_created;
DROP INDEX IF EXISTS idx_reports_unique_reporter_target;
DROP TABLE IF EXISTS reports;

DROP TYPE IF EXISTS moderation_action_type;
DROP TYPE IF EXISTS report_status;
DROP TYPE IF EXISTS report_reason;
DROP TYPE IF EXISTS report_target;
