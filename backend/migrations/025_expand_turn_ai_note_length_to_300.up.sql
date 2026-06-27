ALTER TABLE turns
  DROP CONSTRAINT IF EXISTS chk_turns_ai_note_length;

ALTER TABLE turns
  ADD CONSTRAINT chk_turns_ai_note_length
  CHECK (ai_note IS NULL OR char_length(ai_note) <= 300);
