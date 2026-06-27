CREATE TYPE invite_status AS ENUM ('pending', 'accepted', 'revoked', 'expired');

ALTER TABLE users
  ADD COLUMN invited_by_user_id UUID REFERENCES users(id);

CREATE INDEX idx_users_invited_by_user_id ON users (invited_by_user_id);

CREATE TABLE invites (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  inviter_user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  invited_email       TEXT NOT NULL,
  token_hash          TEXT NOT NULL,
  status              invite_status NOT NULL DEFAULT 'pending',
  expires_at          TIMESTAMPTZ NOT NULL,
  accepted_by_user_id UUID REFERENCES users(id),
  accepted_at         TIMESTAMPTZ,
  revoked_at          TIMESTAMPTZ,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_invites_accept_pair CHECK (
    (accepted_by_user_id IS NULL AND accepted_at IS NULL)
    OR
    (accepted_by_user_id IS NOT NULL AND accepted_at IS NOT NULL)
  ),
  CONSTRAINT chk_invites_revoke_pair CHECK (
    (status <> 'revoked' AND revoked_at IS NULL)
    OR
    (status = 'revoked' AND revoked_at IS NOT NULL)
  )
);

CREATE UNIQUE INDEX idx_invites_token_hash ON invites (token_hash);

CREATE INDEX idx_invites_inviter_created
  ON invites (inviter_user_id, created_at DESC);

CREATE INDEX idx_invites_email_status
  ON invites (LOWER(invited_email), status, created_at DESC);

CREATE UNIQUE INDEX idx_invites_pending_email_unique
  ON invites (LOWER(invited_email))
  WHERE status = 'pending';
