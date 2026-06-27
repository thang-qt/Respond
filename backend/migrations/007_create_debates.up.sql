CREATE TABLE debates (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  topic                TEXT NOT NULL,
  category_id          UUID NOT NULL REFERENCES categories(id),
  time_mode            time_mode NOT NULL,
  turn_limit           INTEGER NOT NULL,
  context              TEXT,
  status               debate_status NOT NULL DEFAULT 'waiting',
  side_a_user_id       UUID NOT NULL REFERENCES users(id),
  side_b_user_id       UUID REFERENCES users(id),
  side_a_anonymous_id  TEXT NOT NULL,
  side_b_anonymous_id  TEXT,
  side_a_revealed      BOOLEAN,
  side_b_revealed      BOOLEAN,
  outcome              debate_outcome,
  winner_side          debate_side,
  turn_count           INTEGER NOT NULL DEFAULT 0,
  current_turn_side    debate_side NOT NULL DEFAULT 'a',
  turn_deadline        TIMESTAMPTZ,
  extension_deadline   TIMESTAMPTZ,
  extension_a_accepted BOOLEAN,
  extension_b_accepted BOOLEAN,
  draw_proposed_by     debate_side,
  draw_proposed_at     TIMESTAMPTZ,
  draw_turn_number     INTEGER,
  open_side            debate_side,
  resigned_user_id     UUID REFERENCES users(id),
  prompt_id            UUID REFERENCES prompts(id),
  upvote_count         INTEGER NOT NULL DEFAULT 0,
  comment_count        INTEGER NOT NULL DEFAULT 0,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  started_at           TIMESTAMPTZ,
  ended_at             TIMESTAMPTZ
);

CREATE INDEX idx_debates_status ON debates (status);
CREATE INDEX idx_debates_created_at ON debates (created_at DESC);
CREATE INDEX idx_debates_started_at ON debates (started_at DESC);
CREATE INDEX idx_debates_ended_at ON debates (ended_at DESC);
CREATE INDEX idx_debates_upvote_count ON debates (upvote_count DESC);
CREATE INDEX idx_debates_category_id ON debates (category_id);
CREATE INDEX idx_debates_side_a_user_id ON debates (side_a_user_id);
CREATE INDEX idx_debates_side_b_user_id ON debates (side_b_user_id);
CREATE INDEX idx_debates_waiting_expiry
  ON debates (created_at) WHERE status = 'waiting';
CREATE INDEX idx_debates_active_turn_deadline
  ON debates (turn_deadline) WHERE status = 'active';
CREATE INDEX idx_debates_extension_deadline
  ON debates (extension_deadline) WHERE status = 'pending_extension';
CREATE INDEX idx_debates_replacement_expiry
  ON debates (turn_deadline) WHERE status = 'waiting_replacement';
CREATE INDEX idx_debates_prompt_id ON debates (prompt_id)
  WHERE prompt_id IS NOT NULL;
