WITH ranked AS (
  SELECT
    id,
    ROW_NUMBER() OVER (
      PARTITION BY to_email, template
      ORDER BY created_at DESC, id DESC
    ) AS rn
  FROM email_jobs
  WHERE template = 'verify_email'
    AND status IN ('pending', 'processing')
)
UPDATE email_jobs ej
SET
  status = 'failed',
  last_error = 'deduped by migration 031',
  updated_at = now()
FROM ranked r
WHERE ej.id = r.id
  AND r.rn > 1;

CREATE UNIQUE INDEX idx_email_jobs_verify_pending_unique
  ON email_jobs (to_email)
  WHERE template = 'verify_email'
    AND status IN ('pending', 'processing');
