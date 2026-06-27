ALTER TABLE reports
  DROP CONSTRAINT IF EXISTS chk_reports_resolution_note_length;

DROP INDEX IF EXISTS idx_debates_hidden;

ALTER TABLE debates
  DROP COLUMN IF EXISTS hidden,
  DROP COLUMN IF EXISTS moderation_pending;

-- Enum values added in up migration are intentionally kept.
