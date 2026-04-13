package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/nats-io/nats.go"

	"github.com/grit-app/grit/internal/analysis/blame"
	"github.com/grit-app/grit/internal/cache"
	gritmetrics "github.com/grit-app/grit/internal/metrics"
	"github.com/grit-app/grit/internal/models"
)

// BlameWorker processes blame/contributor analysis jobs from the NATS queue.
type BlameWorker struct {
	js       nats.JetStreamContext
	analyzer *blame.Analyzer
	cache    *cache.Cache
	cloneDir string
}

// NewBlameWorker creates a new blame worker.
func NewBlameWorker(js nats.JetStreamContext, analyzer *blame.Analyzer, c *cache.Cache, cloneDir string) *BlameWorker {
	return &BlameWorker{
		js:       js,
		analyzer: analyzer,
		cache:    c,
		cloneDir: cloneDir,
	}
}

// Start begins consuming blame jobs from NATS.
func (bw *BlameWorker) Start(ctx context.Context) error {
	sub, err := bw.js.PullSubscribe(BlameSubject, "grit-blame-worker", nats.AckWait(12*time.Minute), nats.MaxDeliver(3))
	if err != nil {
		return fmt.Errorf("blame worker: subscribe: %w", err)
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
					bw.processMessage(ctx, msg)
				}
			}
		}
	}()

	return nil
}

func (bw *BlameWorker) processMessage(ctx context.Context, msg *nats.Msg) {
	var payload JobPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		slog.Error("blame worker: unmarshal payload", "error", err)
		msg.Nak()
		return
	}

	slog.Info("blame worker: processing job",
		"job_id", payload.JobID,
		"owner", payload.Owner,
		"repo", payload.Repo,
	)

	bw.updateJobStatus(ctx, payload.JobID, models.JobStatusRunning)

	// Open the cloned repository.
	repoCloneDir := fmt.Sprintf("%s/%s/%s", bw.cloneDir, payload.Owner, payload.Repo)
	repo, err := git.PlainOpen(repoCloneDir)
	if err != nil {
		slog.Error("blame worker: open repo",
			"job_id", payload.JobID,
			"path", repoCloneDir,
			"error", err,
		)
		now := time.Now().UTC()
		bw.updateJobCompletion(ctx, payload.JobID, models.JobStatusFailed, &now, "failed to open cloned repo")
		gritmetrics.JobsFailedTotal.Inc()
		msg.Ack()
		return
	}

	// Resolve the commit hash.
	commitHash, err := bw.resolveCommitHash(repo, payload.SHA)
	if err != nil {
		slog.Error("blame worker: resolve commit",
			"job_id", payload.JobID,
			"error", err,
		)
		now := time.Now().UTC()
		bw.updateJobCompletion(ctx, payload.JobID, models.JobStatusFailed, &now, "failed to resolve commit")
		gritmetrics.JobsFailedTotal.Inc()
		msg.Ack()
		return
	}

	// Walk the commit tree to find source files.
	sourceFiles, err := bw.listSourceFiles(repo, commitHash)
	if err != nil {
		slog.Error("blame worker: list source files",
			"job_id", payload.JobID,
			"error", err,
		)
		now := time.Now().UTC()
		bw.updateJobCompletion(ctx, payload.JobID, models.JobStatusFailed, &now, "failed to list source files")
		gritmetrics.JobsFailedTotal.Inc()
		msg.Ack()
		return
	}

	// Get repository metadata from core analysis.
	repoMeta := models.Repository{
		Owner:     payload.Owner,
		Name:      payload.Repo,
		FullName:  fmt.Sprintf("%s/%s", payload.Owner, payload.Repo),
		LatestSHA: commitHash.String(),
	}
	coreData, err := bw.cache.GetAnalysis(ctx, payload.Owner, payload.Repo, payload.SHA)
	if err == nil {
		var coreResult models.AnalysisResult
		if json.Unmarshal(coreData, &coreResult) == nil {
			repoMeta = coreResult.Repository
		}
	}

	start := time.Now()
	result, err := bw.analyzer.Analyze(ctx, repo, commitHash, sourceFiles)
	duration := time.Since(start)

	gritmetrics.BlameAnalysisDuration.Observe(duration.Seconds())

	if err != nil {
		slog.Error("blame worker: analysis failed",
			"job_id", payload.JobID,
			"error", err,
		)
		now := time.Now().UTC()
		bw.updateJobCompletion(ctx, payload.JobID, models.JobStatusFailed, &now, err.Error())
		gritmetrics.JobsFailedTotal.Inc()
		msg.Ack()
		return
	}

	result.Repository = repoMeta

	data, err := json.Marshal(result)
	if err != nil {
		slog.Error("blame worker: marshal result", "job_id", payload.JobID, "error", err)
		msg.Nak()
		return
	}

	if err := bw.cache.SetContributors(ctx, payload.Owner, payload.Repo, commitHash.String(), data); err != nil {
		slog.Error("blame worker: cache result", "job_id", payload.JobID, "error", err)
	}

	now := time.Now().UTC()
	bw.updateJobCompletion(ctx, payload.JobID, models.JobStatusCompleted, &now, "")
	bw.cache.DeleteActiveBlameJob(ctx, payload.Owner, payload.Repo, payload.SHA)

	gritmetrics.BlameJobsCompletedTotal.Inc()

	slog.Info("blame worker: job completed",
		"job_id", payload.JobID,
		"owner", payload.Owner,
		"repo", payload.Repo,
		"total_authors", len(result.Authors),
		"bus_factor", result.BusFactor,
		"total_files", result.TotalFilesAnalyzed,
		"total_lines", result.TotalLinesAnalyzed,
		"partial", result.Partial,
		"duration", duration,
	)

	msg.Ack()
}

// resolveCommitHash resolves a SHA string to a plumbing.Hash.
// If SHA is empty, it resolves HEAD.
func (bw *BlameWorker) resolveCommitHash(repo *git.Repository, sha string) (plumbing.Hash, error) {
	if sha != "" {
		return plumbing.NewHash(sha), nil
	}

	ref, err := repo.Head()
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("blame worker: resolve HEAD: %w", err)
	}
	return ref.Hash(), nil
}

// listSourceFiles walks the commit tree and returns paths of recognized source files.
func (bw *BlameWorker) listSourceFiles(repo *git.Repository, commitHash plumbing.Hash) ([]string, error) {
	commit, err := repo.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("blame worker: get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("blame worker: get tree: %w", err)
	}

	var files []string
	err = tree.Files().ForEach(func(f *object.File) error {
		if blame.IsSourceFile(f.Name) {
			files = append(files, f.Name)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("blame worker: walk tree: %w", err)
	}

	return files, nil
}

func (bw *BlameWorker) updateJobStatus(ctx context.Context, jobID string, status models.JobStatus) {
	data, err := bw.cache.GetJob(ctx, jobID)
	if err != nil {
		return
	}

	var job models.AnalysisJob
	if err := json.Unmarshal(data, &job); err != nil {
		return
	}

	job.Status = status
	bw.cache.SetJob(ctx, jobID, &job)
}

func (bw *BlameWorker) updateJobCompletion(ctx context.Context, jobID string, status models.JobStatus, completedAt *time.Time, errMsg string) {
	data, err := bw.cache.GetJob(ctx, jobID)
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
		job.ResultURL = fmt.Sprintf("/api/%s/%s/contributors", job.Owner, job.Repo)
	}

	bw.cache.SetJob(ctx, jobID, &job)
}
