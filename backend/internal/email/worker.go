package email

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"respond/internal/model"
	"respond/internal/store"
)

var retryBackoff = []time.Duration{
	time.Minute,
	5 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
}

type Worker struct {
	Store       *store.Store
	Mailer      Mailer
	Logger      *slog.Logger
	FrontendURL string
	BatchSize   int
}

func NewWorker(st *store.Store, mailer Mailer, logger *slog.Logger, frontendURL string) *Worker {
	return &Worker{
		Store:       st,
		Mailer:      mailer,
		Logger:      logger,
		FrontendURL: frontendURL,
		BatchSize:   20,
	}
}

func (w *Worker) Process(ctx context.Context) {
	if w == nil || w.Mailer == nil || w.Store == nil {
		return
	}

	jobs, err := w.Store.ClaimReadyEmailJobs(ctx, w.BatchSize)
	if err != nil {
		w.Logger.Error("claim email jobs failed", "error", err)
		return
	}
	for _, job := range jobs {
		w.processJob(ctx, job)
	}
}

func (w *Worker) processJob(ctx context.Context, job model.EmailJob) {
	subject, body, err := Render(job.Template, job.PayloadJSON, w.FrontendURL)
	if err != nil {
		w.Logger.Error("render email template failed", "job_id", job.ID, "template", job.Template, "error", err)
		_ = w.Store.MarkEmailJobFailed(ctx, job.ID, truncateErr(err))
		return
	}

	if err := w.Mailer.Send(ctx, Message{To: job.ToEmail, Subject: subject, Body: body}); err != nil {
		if job.Attempts >= job.MaxAttempts {
			if markErr := w.Store.MarkEmailJobFailed(ctx, job.ID, truncateErr(err)); markErr != nil {
				w.Logger.Error("mark email job failed failed", "job_id", job.ID, "error", markErr)
			}
			w.Logger.Error("email send failed permanently", "job_id", job.ID, "attempts", job.Attempts, "error", err)
			return
		}

		delay := backoffForAttempt(job.Attempts)
		next := time.Now().Add(delay)
		if markErr := w.Store.RescheduleEmailJob(ctx, job.ID, next, truncateErr(err)); markErr != nil {
			w.Logger.Error("reschedule email job failed", "job_id", job.ID, "error", markErr)
			return
		}
		w.Logger.Warn("email send failed, rescheduled", "job_id", job.ID, "attempts", job.Attempts, "next_attempt_at", next.UTC(), "error", err)
		return
	}

	if err := w.Store.MarkEmailJobSent(ctx, job.ID); err != nil {
		w.Logger.Error("mark email job sent failed", "job_id", job.ID, "error", err)
		return
	}
	w.Logger.Info("email sent", "job_id", job.ID, "template", job.Template)
}

func backoffForAttempt(attempt int) time.Duration {
	if attempt <= 0 {
		return retryBackoff[0]
	}
	idx := attempt - 1
	if idx >= len(retryBackoff) {
		idx = len(retryBackoff) - 1
	}
	return retryBackoff[idx]
}

func truncateErr(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if len(msg) > 1000 {
		return fmt.Sprintf("%s...", msg[:1000])
	}
	return msg
}
