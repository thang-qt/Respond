ALTER TABLE notifications
  ADD COLUMN turn_number INT;

CREATE INDEX idx_notifications_debate_turn
  ON notifications (debate_id, turn_number)
  WHERE turn_number IS NOT NULL;
