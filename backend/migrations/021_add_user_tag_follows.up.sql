CREATE TABLE user_tag_follows (
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tag_id     UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, tag_id)
);

CREATE INDEX idx_user_tag_follows_tag_id ON user_tag_follows (tag_id);
