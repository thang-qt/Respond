CREATE TABLE email_jobs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  to_email        TEXT NOT NULL,
  template        TEXT NOT NULL,
  payload_json    JSONB NOT NULL DEFAULT '{}'::jsonb,
  status          TEXT NOT NULL DEFAULT 'pending',
  attempts        INTEGER NOT NULL DEFAULT 0,
  max_attempts    INTEGER NOT NULL DEFAULT 5,
  next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_error      TEXT,
  sent_at         TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_email_jobs_status
    CHECK (status IN ('pending', 'processing', 'sent', 'failed'))
);

CREATE INDEX idx_email_jobs_ready
  ON email_jobs (status, next_attempt_at, created_at)
  WHERE status IN ('pending', 'processing');

CREATE INDEX idx_email_jobs_template
  ON email_jobs (template, created_at DESC);
