DROP INDEX IF EXISTS idx_debates_invited_user_id;

ALTER TABLE debates
  DROP COLUMN IF EXISTS challenge_expires_at,
  DROP COLUMN IF EXISTS invited_user_id;

-- Cannot remove enum values from notification_type.
