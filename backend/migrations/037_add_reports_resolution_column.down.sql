DROP INDEX IF EXISTS idx_reports_resolution_created;

ALTER TABLE reports
  DROP CONSTRAINT IF EXISTS chk_reports_resolution_value;

ALTER TABLE reports
  DROP COLUMN IF EXISTS resolution;
