package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/grit-app/grit/internal/analysis/core"
	"github.com/grit-app/grit/internal/cache"
	gritmetrics "github.com/grit-app/grit/internal/metrics"
	"github.com/grit-app/grit/internal/models"
)

type Worker struct {
	js        nats.JetStreamContext
	analyzer  *core.Analyzer
	cache     *cache.Cache
	publisher *Publisher
}

func NewWorker(js nats.JetStreamContext, analyzer *core.Analyzer, c *cache.Cache, publisher *Publisher) *Worker {
	return &Worker{
		js:        js,
		analyzer:  analyzer,
		cache:     c,
		publisher: publisher,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	sub, err := w.js.PullSubscribe(Subject, "grit-worker", nats.AckWait(6*time.Minute), nats.MaxDeliver(3))
	if err != nil {
		return fmt.Errorf("worker: subscribe: %w", err)
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
					w.processMessage(ctx, msg)
				}
			}
		}
	}()

	return nil
}

func (w *Worker) processMessage(ctx context.Context, msg *nats.Msg) {
	var payload JobPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		slog.Error("worker: unmarshal payload", "error", err)
		msg.Nak()
		return
	}

	slog.Info("worker: processing job",
		"job_id", payload.JobID,
		"owner", payload.Owner,
		"repo", payload.Repo,
	)

	w.updateJobStatus(ctx, payload.JobID, models.JobStatusRunning, nil)

	analysisCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	result, err := w.analyzer.Analyze(analysisCtx, payload.Owner, payload.Repo, payload.Token,
		func(step string, status models.SubJobStatus) {
			w.updateJobProgress(ctx, payload.JobID, step, status)
		},
	)

	if err != nil {
		slog.Error("worker: analysis failed",
			"job_id", payload.JobID,
			"error", err,
		)
		now := time.Now().UTC()
		w.updateJobCompletion(ctx, payload.JobID, models.JobStatusFailed, &now, err.Error())
		gritmetrics.JobsFailedTotal.Inc()
		msg.Ack()
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		slog.Error("worker: marshal result", "job_id", payload.JobID, "error", err)
		msg.Nak()
		return
	}

	if err := w.cache.SetAnalysis(ctx, payload.Owner, payload.Repo, result.Repository.LatestSHA, data); err != nil {
		slog.Error("worker: cache result", "job_id", payload.JobID, "error", err)
	}

	now := time.Now().UTC()
	w.updateJobCompletion(ctx, payload.JobID, models.JobStatusCompleted, &now, "")
	w.cache.DeleteActiveJob(ctx, payload.Owner, payload.Repo, result.Repository.LatestSHA)

	gritmetrics.JobsCompletedTotal.Inc()

	slog.Info("worker: job completed",
		"job_id", payload.JobID,
		"owner", payload.Owner,
		"repo", payload.Repo,
		"total_files", result.TotalFiles,
		"total_lines", result.TotalLines,
	)

	// Auto-trigger downstream analysis jobs after core completion.
	if w.publisher != nil {
		compJobID, err := w.publisher.PublishComplexity(ctx, payload.Owner, payload.Repo, result.Repository.LatestSHA, payload.Token)
		if err != nil {
			slog.Error("worker: failed to publish complexity job", "job_id", payload.JobID, "error", err)
		} else {
			slog.Info("worker: complexity job enqueued", "core_job_id", payload.JobID, "complexity_job_id", compJobID)
		}

		blameJobID, err := w.publisher.PublishBlame(ctx, payload.Owner, payload.Repo, result.Repository.LatestSHA, payload.Token)
		if err != nil {
			slog.Error("worker: failed to publish blame job", "job_id", payload.JobID, "error", err)
		} else {
			slog.Info("worker: blame job enqueued", "core_job_id", payload.JobID, "blame_job_id", blameJobID)
		}
	}

	msg.Ack()
}

func (w *Worker) updateJobStatus(ctx context.Context, jobID string, status models.JobStatus, progress *models.JobProgress) {
	data, err := w.cache.GetJob(ctx, jobID)
	if err != nil {
		return
	}

	var job models.AnalysisJob
	if err := json.Unmarshal(data, &job); err != nil {
		return
	}

	job.Status = status
	if progress != nil {
		job.Progress = *progress
	}

	w.cache.SetJob(ctx, jobID, &job)
}

func (w *Worker) updateJobProgress(ctx context.Context, jobID string, step string, status models.SubJobStatus) {
	data, err := w.cache.GetJob(ctx, jobID)
	if err != nil {
		return
	}

	var job models.AnalysisJob
	if err := json.Unmarshal(data, &job); err != nil {
		return
	}

	switch step {
	case "clone":
		job.Progress.Clone = status
	case "file_walk":
		job.Progress.FileWalk = status
	case "metadata_fetch":
		job.Progress.MetadataFetch = status
	case "commit_activity_fetch":
		job.Progress.CommitActivityFetch = status
	}

	w.cache.SetJob(ctx, jobID, &job)
}

func (w *Worker) updateJobCompletion(ctx context.Context, jobID string, status models.JobStatus, completedAt *time.Time, errMsg string) {
	data, err := w.cache.GetJob(ctx, jobID)
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
		job.ResultURL = fmt.Sprintf("/api/%s/%s", job.Owner, job.Repo)
		job.Progress.Clone = models.SubJobCompleted
		job.Progress.FileWalk = models.SubJobCompleted
		job.Progress.MetadataFetch = models.SubJobCompleted
		job.Progress.CommitActivityFetch = models.SubJobCompleted
	}

	w.cache.SetJob(ctx, jobID, &job)
}
