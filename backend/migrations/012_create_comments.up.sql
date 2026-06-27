CREATE TABLE comments (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  debate_id          UUID NOT NULL REFERENCES debates(id) ON DELETE CASCADE,
  parent_id          UUID REFERENCES comments(id) ON DELETE CASCADE,
  user_id            UUID NOT NULL REFERENCES users(id),
  content            TEXT NOT NULL,
  is_reflection      BOOLEAN NOT NULL DEFAULT false,
  is_deleted         BOOLEAN NOT NULL DEFAULT false,
  upvote_count       INTEGER NOT NULL DEFAULT 0,
  moderation_pending BOOLEAN NOT NULL DEFAULT false,
  hidden             BOOLEAN NOT NULL DEFAULT false,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ
);

CREATE INDEX idx_comments_debate_id ON comments (debate_id, created_at);
CREATE INDEX idx_comments_parent_id ON comments (parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_comments_user_id ON comments (user_id);
CREATE INDEX idx_comments_reflections ON comments (debate_id) WHERE is_reflection = true;
CREATE UNIQUE INDEX idx_comments_reflection_unique ON comments (debate_id, user_id) WHERE is_reflection = true;
CREATE INDEX idx_comments_upvotes ON comments (debate_id, upvote_count DESC);
