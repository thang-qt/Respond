ALTER TYPE notification_type ADD VALUE IF NOT EXISTS 'challenge_received';
ALTER TYPE notification_type ADD VALUE IF NOT EXISTS 'challenge_accepted';
ALTER TYPE notification_type ADD VALUE IF NOT EXISTS 'challenge_declined';
ALTER TYPE notification_type ADD VALUE IF NOT EXISTS 'challenge_expired';

ALTER TABLE debates
  ADD COLUMN invited_user_id UUID REFERENCES users(id),
  ADD COLUMN challenge_expires_at TIMESTAMPTZ;

CREATE INDEX idx_debates_invited_user_id
  ON debates (invited_user_id)
  WHERE invited_user_id IS NOT NULL;
