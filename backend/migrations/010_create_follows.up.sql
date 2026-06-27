CREATE TABLE follows (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  debate_id  UUID NOT NULL REFERENCES debates(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_follows_user_debate ON follows (user_id, debate_id);
CREATE INDEX idx_follows_debate_id ON follows (debate_id);
