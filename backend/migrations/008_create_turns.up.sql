CREATE TABLE turns (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  debate_id          UUID NOT NULL REFERENCES debates(id) ON DELETE CASCADE,
  turn_number        INTEGER NOT NULL,
  side               debate_side NOT NULL,
  user_id            UUID NOT NULL REFERENCES users(id),
  anonymous_id       TEXT NOT NULL,
  content            TEXT NOT NULL,
  is_system          BOOLEAN NOT NULL DEFAULT false,
  moderation_pending BOOLEAN NOT NULL DEFAULT false,
  hidden             BOOLEAN NOT NULL DEFAULT false,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_turns_debate_id_number
  ON turns (debate_id, turn_number);
CREATE UNIQUE INDEX idx_turns_debate_turn_unique
  ON turns (debate_id, turn_number);
CREATE INDEX idx_turns_user_id ON turns (user_id);
