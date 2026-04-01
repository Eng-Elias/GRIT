package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/nats-io/nats.go"

	"github.com/grit-app/grit/internal/analysis/churn"
	"github.com/grit-app/grit/internal/cache"
	gritmetrics "github.com/grit-app/grit/internal/metrics"
	"github.com/grit-app/grit/internal/models"
)

// ChurnWorker processes churn analysis jobs from the NATS queue.
type ChurnWorker struct {
	js       nats.JetStreamContext
	analyzer *churn.Analyzer
	cache    *cache.Cache
	cloneDir string
}

// NewChurnWorker creates a new churn worker.
func NewChurnWorker(js nats.JetStreamContext, analyzer *churn.Analyzer, c *cache.Cache, cloneDir string) *ChurnWorker {
	return &ChurnWorker{
		js:       js,
		analyzer: analyzer,
		cache:    c,
		cloneDir: cloneDir,
	}
}

// Start begins consuming churn jobs from NATS.
func (cw *ChurnWorker) Start(ctx context.Context) error {
	sub, err := cw.js.PullSubscribe(ChurnSubject, "grit-churn-worker", nats.AckWait(6*time.Minute), nats.MaxDeliver(3))
	if err != nil {
		return fmt.Errorf("churn worker: subscribe: %w", err)
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

func (cw *ChurnWorker) processMessage(ctx context.Context, msg *nats.Msg) {
	var payload JobPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		slog.Error("churn worker: unmarshal payload", "error", err)
		msg.Nak()
		return
	}

	slog.Info("churn worker: processing job",
		"job_id", payload.JobID,
		"owner", payload.Owner,
		"repo", payload.Repo,
	)

	cw.updateJobStatus(ctx, payload.JobID, models.JobStatusRunning)

	// Open the cloned repository.
	repoCloneDir := fmt.Sprintf("%s/%s/%s", cw.cloneDir, payload.Owner, payload.Repo)
	repo, err := git.PlainOpen(repoCloneDir)
	if err != nil {
		slog.Error("churn worker: open repo",
			"job_id", payload.JobID,
			"path", repoCloneDir,
			"error", err,
		)
		now := time.Now().UTC()
		cw.updateJobCompletion(ctx, payload.JobID, models.JobStatusFailed, &now, "failed to open cloned repo")
		gritmetrics.JobsFailedTotal.Inc()
		msg.Ack()
		return
	}

	// Fetch complexity data from cache (optional — analysis proceeds without it).
	complexityData, _ := cw.cache.GetComplexity(ctx, payload.Owner, payload.Repo, payload.SHA)

	// Get repository metadata from core analysis.
	repoMeta := models.Repository{
		Owner:    payload.Owner,
		Name:     payload.Repo,
		FullName: fmt.Sprintf("%s/%s", payload.Owner, payload.Repo),
	}
	coreData, err := cw.cache.GetAnalysis(ctx, payload.Owner, payload.Repo, payload.SHA)
	if err == nil {
		var coreResult models.AnalysisResult
		if json.Unmarshal(coreData, &coreResult) == nil {
			repoMeta = coreResult.Repository
		}
	}

	start := time.Now()
	result, err := cw.analyzer.Analyze(repo, complexityData, repoMeta)
	duration := time.Since(start)

	gritmetrics.ChurnAnalysisDuration.Observe(duration.Seconds())

	if err != nil {
		slog.Error("churn worker: analysis failed",
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
		slog.Error("churn worker: marshal result", "job_id", payload.JobID, "error", err)
		msg.Nak()
		return
	}

	if err := cw.cache.SetChurn(ctx, payload.Owner, payload.Repo, payload.SHA, data); err != nil {
		slog.Error("churn worker: cache result", "job_id", payload.JobID, "error", err)
	}

	now := time.Now().UTC()
	cw.updateJobCompletion(ctx, payload.JobID, models.JobStatusCompleted, &now, "")
	cw.cache.DeleteActiveChurnJob(ctx, payload.Owner, payload.Repo, payload.SHA)

	gritmetrics.JobsCompletedTotal.Inc()

	slog.Info("churn worker: job completed",
		"job_id", payload.JobID,
		"owner", payload.Owner,
		"repo", payload.Repo,
		"total_commits", result.TotalCommits,
		"files_churned", result.TotalFilesChurned,
		"critical_count", result.CriticalCount,
		"stale_count", result.StaleCount,
		"duration", duration,
	)

	msg.Ack()
}

func (cw *ChurnWorker) updateJobStatus(ctx context.Context, jobID string, status models.JobStatus) {
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

func (cw *ChurnWorker) updateJobCompletion(ctx context.Context, jobID string, status models.JobStatus, completedAt *time.Time, errMsg string) {
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
		job.ResultURL = fmt.Sprintf("/api/%s/%s/churn-matrix", job.Owner, job.Repo)
	}

	cw.cache.SetJob(ctx, jobID, &job)
}
