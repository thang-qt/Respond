DROP INDEX IF EXISTS idx_notifications_debate_turn;

ALTER TABLE notifications
  DROP COLUMN IF EXISTS turn_number;
