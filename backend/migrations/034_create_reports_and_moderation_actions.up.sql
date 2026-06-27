CREATE TYPE report_target AS ENUM ('turn', 'comment');

CREATE TYPE report_reason AS ENUM ('hate', 'harassment', 'spam', 'off_topic', 'other');

CREATE TYPE report_status AS ENUM ('open', 'dismissed', 'actioned');

CREATE TYPE moderation_action_type AS ENUM (
  'hide_content',
  'restore_content',
  'dismiss_report',
  'change_user_role'
);

CREATE TABLE reports (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  reporter_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  target_type         report_target NOT NULL,
  target_id           UUID NOT NULL,
  reason              report_reason NOT NULL,
  details             TEXT,
  status              report_status NOT NULL DEFAULT 'open',
  trusted_report      BOOLEAN NOT NULL DEFAULT false,
  reviewed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at         TIMESTAMPTZ,
  resolution_note     TEXT,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_reports_unique_reporter_target
  ON reports (reporter_id, target_type, target_id);

CREATE INDEX idx_reports_target_created
  ON reports (target_type, target_id, created_at DESC);

CREATE INDEX idx_reports_status_created
  ON reports (status, created_at DESC);

CREATE TABLE moderation_actions (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  action_type   moderation_action_type NOT NULL,
  target_type   TEXT NOT NULL,
  target_id     UUID NOT NULL,
  report_id     UUID REFERENCES reports(id) ON DELETE SET NULL,
  payload_json  JSONB NOT NULL DEFAULT '{}'::jsonb,
  reason        TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_moderation_actions_target
  ON moderation_actions (target_type, target_id, created_at DESC);

CREATE INDEX idx_moderation_actions_actor
  ON moderation_actions (actor_user_id, created_at DESC);
