DROP INDEX IF EXISTS idx_invites_pending_email_unique;
DROP INDEX IF EXISTS idx_invites_email_status;
DROP INDEX IF EXISTS idx_invites_inviter_created;
DROP INDEX IF EXISTS idx_invites_token_hash;

DROP TABLE IF EXISTS invites;

DROP INDEX IF EXISTS idx_users_invited_by_user_id;

ALTER TABLE users
  DROP COLUMN IF EXISTS invited_by_user_id;

DROP TYPE IF EXISTS invite_status;
