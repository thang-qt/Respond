CREATE TYPE vote_target_type AS ENUM ('debate', 'comment');

CREATE TABLE votes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  target_type vote_target_type NOT NULL,
  target_id   UUID NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_votes_user_id ON votes (user_id);
CREATE INDEX idx_votes_target ON votes (target_type, target_id);
CREATE UNIQUE INDEX idx_votes_unique ON votes (user_id, target_type, target_id);
