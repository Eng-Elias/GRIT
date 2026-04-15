package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/nats-io/nats.go"

	"github.com/grit-app/grit/internal/analysis/temporal"
	"github.com/grit-app/grit/internal/cache"
	gritmetrics "github.com/grit-app/grit/internal/metrics"
	"github.com/grit-app/grit/internal/models"
)

// TemporalWorker processes temporal analysis jobs from the NATS queue.
type TemporalWorker struct {
	js       nats.JetStreamContext
	cache    *cache.Cache
	cloneDir string
}

// NewTemporalWorker creates a new temporal worker.
func NewTemporalWorker(js nats.JetStreamContext, c *cache.Cache, cloneDir string) *TemporalWorker {
	return &TemporalWorker{
		js:       js,
		cache:    c,
		cloneDir: cloneDir,
	}
}

// Start begins consuming temporal jobs from NATS.
func (tw *TemporalWorker) Start(ctx context.Context) error {
	sub, err := tw.js.PullSubscribe(TemporalSubject, "grit-temporal-worker", nats.AckWait(12*time.Minute), nats.MaxDeliver(3))
	if err != nil {
		return fmt.Errorf("temporal worker: subscribe: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msgs, err := sub.Fetch(1, nats.MaxWait(5*time.Second))
				if err != nil {
					continue
				}
				for _, msg := range msgs {
					tw.processMessage(ctx, msg)
				}
			}
		}
	}()

	return nil
}

func (tw *TemporalWorker) processMessage(ctx context.Context, msg *nats.Msg) {
	var payload JobPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		slog.Error("temporal worker: unmarshal payload", "error", err)
		msg.Nak()
		return
	}

	slog.Info("temporal worker: processing job",
		"job_id", payload.JobID,
		"owner", payload.Owner,
		"repo", payload.Repo,
	)

	tw.updateJobStatus(ctx, payload.JobID, models.JobStatusRunning)

	// Open the cloned repository.
	repoCloneDir := fmt.Sprintf("%s/%s/%s", tw.cloneDir, payload.Owner, payload.Repo)
	repo, err := git.PlainOpen(repoCloneDir)
	if err != nil {
		slog.Error("temporal worker: open repo",
			"job_id", payload.JobID,
			"path", repoCloneDir,
			"error", err,
		)
		now := time.Now().UTC()
		tw.updateJobCompletion(ctx, payload.JobID, models.JobStatusFailed, &now, "failed to open cloned repo")
		gritmetrics.JobsFailedTotal.Inc()
		msg.Ack()
		return
	}

	start := time.Now()
	result, err := temporal.Analyze(ctx, repo, temporal.AnalyzeOptions{
		Period: "3y",
		Owner:  payload.Owner,
		Repo:   payload.Repo,
		Token:  payload.Token,
	})
	duration := time.Since(start)

	gritmetrics.TemporalAnalysisDuration.Observe(duration.Seconds())

	if err != nil {
		slog.Error("temporal worker: analysis failed",
			"job_id", payload.JobID,
			"error", err,
		)
		now := time.Now().UTC()
		tw.updateJobCompletion(ctx, payload.JobID, models.JobStatusFailed, &now, err.Error())
		gritmetrics.JobsFailedTotal.Inc()
		msg.Ack()
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		slog.Error("temporal worker: marshal result", "job_id", payload.JobID, "error", err)
		msg.Nak()
		return
	}

	if err := tw.cache.SetTemporal(ctx, payload.Owner, payload.Repo, payload.SHA, "3y", data); err != nil {
		slog.Error("temporal worker: cache result", "job_id", payload.JobID, "error", err)
	}

	now := time.Now().UTC()
	tw.updateJobCompletion(ctx, payload.JobID, models.JobStatusCompleted, &now, "")
	tw.cache.DeleteActiveTemporalJob(ctx, payload.Owner, payload.Repo, payload.SHA)

	gritmetrics.TemporalJobsCompletedTotal.Inc()

	slog.Info("temporal worker: job completed",
		"job_id", payload.JobID,
		"owner", payload.Owner,
		"repo", payload.Repo,
		"total_months", result.TotalMonths,
		"total_weeks", result.TotalWeeks,
		"duration", duration,
	)

	msg.Ack()
}

func (tw *TemporalWorker) updateJobStatus(ctx context.Context, jobID string, status models.JobStatus) {
	data, err := tw.cache.GetJob(ctx, jobID)
	if err != nil {
		return
	}

	var job models.AnalysisJob
	if err := json.Unmarshal(data, &job); err != nil {
		return
	}

	job.Status = status
	tw.cache.SetJob(ctx, jobID, &job)
}

func (tw *TemporalWorker) updateJobCompletion(ctx context.Context, jobID string, status models.JobStatus, completedAt *time.Time, errMsg string) {
	data, err := tw.cache.GetJob(ctx, jobID)
	if err != nil {
		return
	}

	var job models.AnalysisJob
	if err := json.Unmarshal(data, &job); err != nil {
		return
	}

	job.Status = status
	job.CompletedAt = completedAt
	job.Error = errMsg

	if status == models.JobStatusCompleted {
		job.ResultURL = fmt.Sprintf("/api/%s/%s/temporal", job.Owner, job.Repo)
	}

	tw.cache.SetJob(ctx, jobID, &job)
}
