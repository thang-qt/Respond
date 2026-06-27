CREATE TYPE notification_type AS ENUM (
  'your_turn',
  'debate_joined',
  'debate_ended',
  'turn_expiring',
  'seat_open',
  'draw_proposed',
  'comment_on_reflection',
  'replacement_joined'
);

CREATE TABLE notifications (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type        notification_type NOT NULL,
  message     TEXT NOT NULL,
  debate_id   UUID REFERENCES debates(id) ON DELETE CASCADE,
  is_read     BOOLEAN NOT NULL DEFAULT false,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user_id
  ON notifications (user_id, created_at DESC);

CREATE INDEX idx_notifications_unread
  ON notifications (user_id)
  WHERE is_read = false;
