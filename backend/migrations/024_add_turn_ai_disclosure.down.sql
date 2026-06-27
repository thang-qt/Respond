ALTER TABLE turns
  DROP CONSTRAINT IF EXISTS chk_turns_ai_note_length;

ALTER TABLE turns
  DROP COLUMN IF EXISTS ai_note,
  DROP COLUMN IF EXISTS ai_assisted;
