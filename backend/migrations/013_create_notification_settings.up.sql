CREATE TABLE notification_settings (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id               UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  email_your_turn       BOOLEAN NOT NULL DEFAULT true,
  email_debate_joined   BOOLEAN NOT NULL DEFAULT true,
  email_debate_ended    BOOLEAN NOT NULL DEFAULT false,
  email_turn_expiring   BOOLEAN NOT NULL DEFAULT true,
  email_seat_open       BOOLEAN NOT NULL DEFAULT false,
  email_draw_proposed   BOOLEAN NOT NULL DEFAULT true,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_notification_settings_user_id
  ON notification_settings (user_id);
