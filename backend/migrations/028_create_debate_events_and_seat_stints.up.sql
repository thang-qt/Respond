CREATE TABLE debate_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  debate_id UUID NOT NULL REFERENCES debates(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  side debate_side,
  actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_debate_events_type CHECK (
    event_type IN (
      'seat_opened',
      'replacement_joined',
      'conceded',
      'draw_proposed',
      'draw_declined',
      'draw_accepted',
      'extension_proposed',
      'extension_accepted',
      'extension_declined',
      'walkover',
      'replacement_expired',
      'extension_expired'
    )
  )
);

CREATE INDEX idx_debate_events_debate_created_at ON debate_events (debate_id, created_at, id);

CREATE TABLE debate_seat_stints (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  debate_id UUID NOT NULL REFERENCES debates(id) ON DELETE CASCADE,
  side debate_side NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id),
  anonymous_id TEXT NOT NULL,
  stint_index INTEGER NOT NULL,
  joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  left_at TIMESTAMPTZ,
  left_reason TEXT,
  replaced_by_stint_id UUID REFERENCES debate_seat_stints(id) ON DELETE SET NULL,
  CONSTRAINT chk_debate_seat_stints_index CHECK (stint_index > 0),
  CONSTRAINT chk_debate_seat_stints_left_reason CHECK (
    left_reason IS NULL OR left_reason IN ('resigned', 'finished', 'walkover')
  ),
  CONSTRAINT chk_debate_seat_stints_time CHECK (left_at IS NULL OR left_at >= joined_at),
  CONSTRAINT uq_debate_seat_stints_index UNIQUE (debate_id, side, stint_index)
);

CREATE UNIQUE INDEX idx_debate_seat_stints_active
  ON debate_seat_stints (debate_id, side)
  WHERE left_at IS NULL;

CREATE INDEX idx_debate_seat_stints_debate_side_joined
  ON debate_seat_stints (debate_id, side, joined_at, id);
