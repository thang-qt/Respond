ALTER TYPE report_target ADD VALUE IF NOT EXISTS 'debate';

ALTER TYPE report_reason ADD VALUE IF NOT EXISTS 'illegal';

ALTER TABLE debates
  ADD COLUMN IF NOT EXISTS moderation_pending BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS hidden BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_debates_hidden ON debates (hidden);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_reports_resolution_note_length'
  ) THEN
    ALTER TABLE reports
      ADD CONSTRAINT chk_reports_resolution_note_length
      CHECK (resolution_note IS NULL OR char_length(resolution_note) <= 500);
  END IF;
END $$;
