ALTER TABLE reports
  ADD COLUMN resolution TEXT;

ALTER TABLE reports
  ADD CONSTRAINT chk_reports_resolution_value
  CHECK (resolution IS NULL OR resolution IN ('dismiss', 'hide', 'restore'));

UPDATE reports
SET resolution = CASE
  WHEN status = 'dismissed' THEN 'dismiss'
  WHEN status = 'actioned' AND EXISTS (
    SELECT 1
    FROM moderation_actions ma
    WHERE ma.report_id = reports.id
      AND ma.action_type = 'hide_content'
  ) THEN 'hide'
  WHEN status = 'actioned' AND EXISTS (
    SELECT 1
    FROM moderation_actions ma
    WHERE ma.report_id = reports.id
      AND ma.action_type = 'restore_content'
  ) THEN 'restore'
  ELSE NULL
END;

CREATE INDEX idx_reports_resolution_created
  ON reports (resolution, created_at DESC)
  WHERE resolution IS NOT NULL;
