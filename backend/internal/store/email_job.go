package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"respond/internal/model"
)

type CreateEmailJobParams struct {
	ToEmail     string
	Template    string
	PayloadJSON json.RawMessage
}

var ErrEmailJobAlreadyQueued = errors.New("email job already queued")

func (s *Store) CreateEmailJob(ctx context.Context, params CreateEmailJobParams) (model.EmailJob, error) {
	payload := params.PayloadJSON
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}

	const query = `
		INSERT INTO email_jobs (to_email, template, payload_json)
		VALUES ($1, $2, $3)
		RETURNING id, to_email, template, payload_json, status, attempts,
			max_attempts, next_attempt_at, last_error, sent_at, created_at, updated_at
	`

	var job model.EmailJob
	err := s.DB.QueryRowContext(ctx, query, params.ToEmail, params.Template, []byte(payload)).Scan(
		&job.ID,
		&job.ToEmail,
		&job.Template,
		&job.PayloadJSON,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.NextAttemptAt,
		&job.LastError,
		&job.SentAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "idx_email_jobs_verify_pending_unique" {
			return model.EmailJob{}, ErrEmailJobAlreadyQueued
		}
		return model.EmailJob{}, fmt.Errorf("create email job: %w", err)
	}
	return job, nil
}

func (s *Store) UpsertPendingVerificationEmailJob(ctx context.Context, toEmail string, payload json.RawMessage) (model.EmailJob, error) {
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}

	const updateQuery = `
		UPDATE email_jobs
		SET payload_json = $2,
			status = 'pending',
			next_attempt_at = now(),
			last_error = NULL,
			updated_at = now()
		WHERE template = 'verify_email'
		  AND to_email = $1
		  AND status IN ('pending', 'processing')
		RETURNING id, to_email, template, payload_json, status, attempts,
			max_attempts, next_attempt_at, last_error, sent_at, created_at, updated_at
	`

	var job model.EmailJob
	err := s.DB.QueryRowContext(ctx, updateQuery, toEmail, []byte(payload)).Scan(
		&job.ID,
		&job.ToEmail,
		&job.Template,
		&job.PayloadJSON,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.NextAttemptAt,
		&job.LastError,
		&job.SentAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err == nil {
		return job, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return model.EmailJob{}, fmt.Errorf("upsert verification email job update: %w", err)
	}

	job, err = s.CreateEmailJob(ctx, CreateEmailJobParams{
		ToEmail:     toEmail,
		Template:    "verify_email",
		PayloadJSON: payload,
	})
	if err == nil {
		return job, nil
	}
	if !errors.Is(err, ErrEmailJobAlreadyQueued) {
		return model.EmailJob{}, fmt.Errorf("upsert verification email job create: %w", err)
	}

	err = s.DB.QueryRowContext(ctx, updateQuery, toEmail, []byte(payload)).Scan(
		&job.ID,
		&job.ToEmail,
		&job.Template,
		&job.PayloadJSON,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.NextAttemptAt,
		&job.LastError,
		&job.SentAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return model.EmailJob{}, fmt.Errorf("upsert verification email job retry update: %w", err)
	}
	return job, nil
}

func (s *Store) ClaimReadyEmailJobs(ctx context.Context, limit int) ([]model.EmailJob, error) {
	if limit <= 0 {
		limit = 10
	}

	const query = `
		WITH claim AS (
			SELECT id
			FROM email_jobs
			WHERE status IN ('pending', 'processing')
			  AND next_attempt_at <= now()
			ORDER BY created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE email_jobs e
		SET status = 'processing',
			attempts = e.attempts + 1,
			updated_at = now()
		FROM claim
		WHERE e.id = claim.id
		RETURNING e.id, e.to_email, e.template, e.payload_json, e.status,
			e.attempts, e.max_attempts, e.next_attempt_at, e.last_error,
			e.sent_at, e.created_at, e.updated_at
	`

	rows, err := s.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("claim email jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]model.EmailJob, 0, limit)
	for rows.Next() {
		var job model.EmailJob
		if err := rows.Scan(
			&job.ID,
			&job.ToEmail,
			&job.Template,
			&job.PayloadJSON,
			&job.Status,
			&job.Attempts,
			&job.MaxAttempts,
			&job.NextAttemptAt,
			&job.LastError,
			&job.SentAt,
			&job.CreatedAt,
			&job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan claimed email job: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate claimed email jobs: %w", err)
	}

	return jobs, nil
}

func (s *Store) MarkEmailJobSent(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE email_jobs
		SET status = 'sent',
			sent_at = now(),
			last_error = NULL,
			updated_at = now()
		WHERE id = $1
	`
	if _, err := s.DB.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("mark email job sent: %w", err)
	}
	return nil
}

func (s *Store) RescheduleEmailJob(ctx context.Context, id uuid.UUID, nextAttemptAt time.Time, lastError string) error {
	const query = `
		UPDATE email_jobs
		SET status = 'pending',
			next_attempt_at = $2,
			last_error = $3,
			updated_at = now()
		WHERE id = $1
	`
	if _, err := s.DB.ExecContext(ctx, query, id, nextAttemptAt, lastError); err != nil {
		return fmt.Errorf("reschedule email job: %w", err)
	}
	return nil
}

func (s *Store) MarkEmailJobFailed(ctx context.Context, id uuid.UUID, lastError string) error {
	const query = `
		UPDATE email_jobs
		SET status = 'failed',
			last_error = $2,
			updated_at = now()
		WHERE id = $1
	`
	if _, err := s.DB.ExecContext(ctx, query, id, lastError); err != nil {
		return fmt.Errorf("mark email job failed: %w", err)
	}
	return nil
}
