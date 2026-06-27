ALTER TABLE turns
  ADD COLUMN ai_assisted BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN ai_note TEXT;

ALTER TABLE turns
  ADD CONSTRAINT chk_turns_ai_note_length
  CHECK (ai_note IS NULL OR char_length(ai_note) <= 200);
