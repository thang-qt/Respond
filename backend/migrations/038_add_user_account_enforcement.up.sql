CREATE TYPE account_status AS ENUM ('active', 'suspended', 'banned');

ALTER TABLE users
  ADD COLUMN account_status account_status NOT NULL DEFAULT 'active';

CREATE INDEX idx_users_account_status ON users (account_status);

CREATE TYPE user_enforcement_action_type AS ENUM (
  'warning',
  'restriction',
  'suspension',
  'ban',
  'revoke'
);

CREATE TYPE user_capability AS ENUM (
  'create_debate',
  'comment',
  'vote',
  'follow',
  'report'
);

CREATE TABLE user_enforcement_actions (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  target_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  action_type   user_enforcement_action_type NOT NULL,
  capabilities  user_capability[] NOT NULL DEFAULT '{}',
  expires_at    TIMESTAMPTZ,
  revoked_at    TIMESTAMPTZ,
  note          TEXT NOT NULL,
  payload_json  JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_user_enforcement_note_length
    CHECK (char_length(btrim(note)) BETWEEN 1 AND 500),
  CONSTRAINT chk_user_enforcement_capabilities_required
    CHECK (
      (action_type = 'restriction' AND cardinality(capabilities) >= 1)
      OR (action_type <> 'restriction' AND cardinality(capabilities) = 0)
    ),
  CONSTRAINT chk_user_enforcement_expires_action
    CHECK (
      (action_type IN ('restriction', 'suspension') OR expires_at IS NULL)
    )
);

CREATE INDEX idx_user_enforcement_actions_target_created
  ON user_enforcement_actions (target_user_id, created_at DESC);

CREATE INDEX idx_user_enforcement_actions_active
  ON user_enforcement_actions (target_user_id, action_type, expires_at)
  WHERE revoked_at IS NULL;

ALTER TYPE moderation_action_type ADD VALUE IF NOT EXISTS 'warn_user';
ALTER TYPE moderation_action_type ADD VALUE IF NOT EXISTS 'restrict_user';
ALTER TYPE moderation_action_type ADD VALUE IF NOT EXISTS 'suspend_user';
ALTER TYPE moderation_action_type ADD VALUE IF NOT EXISTS 'ban_user';
ALTER TYPE moderation_action_type ADD VALUE IF NOT EXISTS 'revoke_user_enforcement';

ALTER TYPE notification_type ADD VALUE IF NOT EXISTS 'account_enforcement';
ALTER TYPE notification_type ADD VALUE IF NOT EXISTS 'account_enforcement_revoked';
