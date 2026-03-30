package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/grit-app/grit/internal/analysis/complexity"
	"github.com/grit-app/grit/internal/cache"
	gritmetrics "github.com/grit-app/grit/internal/metrics"
	"github.com/grit-app/grit/internal/models"
)

// ComplexityWorker processes complexity analysis jobs from the NATS queue.
type ComplexityWorker struct {
	js       nats.JetStreamContext
	analyzer *complexity.Analyzer
	cache    *cache.Cache
	cloneDir string
}

// NewComplexityWorker creates a new complexity worker.
func NewComplexityWorker(js nats.JetStreamContext, analyzer *complexity.Analyzer, c *cache.Cache, cloneDir string) *ComplexityWorker {
	return &ComplexityWorker{
		js:       js,
		analyzer: analyzer,
		cache:    c,
		cloneDir: cloneDir,
	}
}

// Start begins consuming complexity jobs from NATS.
func (cw *ComplexityWorker) Start(ctx context.Context) error {
	sub, err := cw.js.PullSubscribe(ComplexitySubject, "grit-complexity-worker", nats.AckWait(6*time.Minute), nats.MaxDeliver(3))
	if err != nil {
		return fmt.Errorf("complexity worker: subscribe: %w", err)
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
					cw.processMessage(ctx, msg)
				}
			}
		}
	}()

	return nil
}

func (cw *ComplexityWorker) processMessage(ctx context.Context, msg *nats.Msg) {
	var payload JobPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		slog.Error("complexity worker: unmarshal payload", "error", err)
		msg.Nak()
		return
	}

	slog.Info("complexity worker: processing job",
		"job_id", payload.JobID,
		"owner", payload.Owner,
		"repo", payload.Repo,
	)

	cw.updateJobStatus(ctx, payload.JobID, models.JobStatusRunning)

	analysisCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Resolve clone directory for this repo.
	repoCloneDir := fmt.Sprintf("%s/%s/%s", cw.cloneDir, payload.Owner, payload.Repo)

	// Get core analysis result to obtain file list.
	coreData, err := cw.cache.GetAnalysis(ctx, payload.Owner, payload.Repo, payload.SHA)
	if err != nil {
		slog.Error("complexity worker: core analysis not found",
			"job_id", payload.JobID,
			"error", err,
		)
		now := time.Now().UTC()
		cw.updateJobCompletion(ctx, payload.JobID, models.JobStatusFailed, &now, "core analysis result not found")
		msg.Ack()
		return
	}

	var coreResult models.AnalysisResult
	if err := json.Unmarshal(coreData, &coreResult); err != nil {
		slog.Error("complexity worker: unmarshal core result", "job_id", payload.JobID, "error", err)
		now := time.Now().UTC()
		cw.updateJobCompletion(ctx, payload.JobID, models.JobStatusFailed, &now, "failed to parse core result")
		msg.Ack()
		return
	}

	start := time.Now()
	result, err := cw.analyzer.Analyze(analysisCtx, repoCloneDir, coreResult.Files, coreResult.Repository)
	duration := time.Since(start)

	gritmetrics.ComplexityAnalysisDuration.Observe(duration.Seconds())

	if err != nil {
		slog.Error("complexity worker: analysis failed",
			"job_id", payload.JobID,
			"error", err,
		)
		now := time.Now().UTC()
		cw.updateJobCompletion(ctx, payload.JobID, models.JobStatusFailed, &now, err.Error())
		gritmetrics.JobsFailedTotal.Inc()
		msg.Ack()
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		slog.Error("complexity worker: marshal result", "job_id", payload.JobID, "error", err)
		msg.Nak()
		return
	}

	if err := cw.cache.SetComplexity(ctx, payload.Owner, payload.Repo, payload.SHA, data); err != nil {
		slog.Error("complexity worker: cache result", "job_id", payload.JobID, "error", err)
	}

	now := time.Now().UTC()
	cw.updateJobCompletion(ctx, payload.JobID, models.JobStatusCompleted, &now, "")
	cw.cache.DeleteActiveComplexityJob(ctx, payload.Owner, payload.Repo, payload.SHA)

	gritmetrics.JobsCompletedTotal.Inc()

	slog.Info("complexity worker: job completed",
		"job_id", payload.JobID,
		"owner", payload.Owner,
		"repo", payload.Repo,
		"files_analyzed", result.TotalFilesAnalyzed,
		"total_functions", result.TotalFunctionCount,
		"duration", duration,
	)

	msg.Ack()
}

func (cw *ComplexityWorker) updateJobStatus(ctx context.Context, jobID string, status models.JobStatus) {
	data, err := cw.cache.GetJob(ctx, jobID)
	if err != nil {
		return
	}

	var job models.AnalysisJob
	if err := json.Unmarshal(data, &job); err != nil {
		return
	}

	job.Status = status
	cw.cache.SetJob(ctx, jobID, &job)
}

func (cw *ComplexityWorker) updateJobCompletion(ctx context.Context, jobID string, status models.JobStatus, completedAt *time.Time, errMsg string) {
	data, err := cw.cache.GetJob(ctx, jobID)
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
		job.ResultURL = fmt.Sprintf("/api/%s/%s/complexity", job.Owner, job.Repo)
	}

	cw.cache.SetJob(ctx, jobID, &job)
}
