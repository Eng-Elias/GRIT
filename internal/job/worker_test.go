package job

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grit-app/grit/internal/analysis/core"
	"github.com/grit-app/grit/internal/cache"
	"github.com/grit-app/grit/internal/models"
)

func redisURL() string {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "redis://localhost:6379"
	}
	return url
}

func natsURL() string {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "nats://localhost:4222"
	}
	return url
}

func skipIfNoRedis(t *testing.T) *cache.Cache {
	t.Helper()
	c, err := cache.New(redisURL())
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func skipIfNoNATS(t *testing.T) nats.JetStreamContext {
	t.Helper()
	nc, err := nats.Connect(natsURL())
	if err != nil {
		t.Skipf("NATS not available: %v", err)
	}
	t.Cleanup(func() { nc.Close() })
	js, err := nc.JetStream()
	if err != nil {
		t.Skipf("JetStream not available: %v", err)
	}
	return js
}

func TestWorker_UpdateJobStatus(t *testing.T) {
	c := skipIfNoRedis(t)
	ctx := context.Background()

	jobID := "test-job-status-" + time.Now().Format("150405")
	initialJob := models.AnalysisJob{
		JobID:     jobID,
		Owner:     "test-owner",
		Repo:      "test-repo",
		SHA:       "abc123",
		Status:    models.JobStatusQueued,
		Progress:  models.NewJobProgress(),
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, c.SetJob(ctx, jobID, &initialJob))

	js := skipIfNoNATS(t)
	analyzer := core.NewAnalyzer("/tmp/grit-test", 51200)
	w := NewWorker(js, analyzer, c, nil)

	w.updateJobStatus(ctx, jobID, models.JobStatusRunning, nil)

	data, err := c.GetJob(ctx, jobID)
	require.NoError(t, err)

	var job models.AnalysisJob
	require.NoError(t, json.Unmarshal(data, &job))

	assert.Equal(t, models.JobStatusRunning, job.Status)
	assert.Equal(t, models.SubJobPending, job.Progress.Clone)
}

func TestWorker_UpdateJobProgress(t *testing.T) {
	c := skipIfNoRedis(t)
	ctx := context.Background()

	jobID := "test-job-progress-" + time.Now().Format("150405")
	initialJob := models.AnalysisJob{
		JobID:     jobID,
		Owner:     "test-owner",
		Repo:      "test-repo",
		SHA:       "abc123",
		Status:    models.JobStatusRunning,
		Progress:  models.NewJobProgress(),
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, c.SetJob(ctx, jobID, &initialJob))

	js := skipIfNoNATS(t)
	analyzer := core.NewAnalyzer("/tmp/grit-test", 51200)
	w := NewWorker(js, analyzer, c, nil)

	w.updateJobProgress(ctx, jobID, "clone", models.SubJobRunning)

	data, err := c.GetJob(ctx, jobID)
	require.NoError(t, err)
	var job1 models.AnalysisJob
	require.NoError(t, json.Unmarshal(data, &job1))
	assert.Equal(t, models.SubJobRunning, job1.Progress.Clone)

	w.updateJobProgress(ctx, jobID, "clone", models.SubJobCompleted)
	w.updateJobProgress(ctx, jobID, "file_walk", models.SubJobRunning)

	data, err = c.GetJob(ctx, jobID)
	require.NoError(t, err)
	var job2 models.AnalysisJob
	require.NoError(t, json.Unmarshal(data, &job2))
	assert.Equal(t, models.SubJobCompleted, job2.Progress.Clone)
	assert.Equal(t, models.SubJobRunning, job2.Progress.FileWalk)
}

func TestWorker_UpdateJobCompletion_Success(t *testing.T) {
	c := skipIfNoRedis(t)
	ctx := context.Background()

	jobID := "test-job-complete-" + time.Now().Format("150405")
	initialJob := models.AnalysisJob{
		JobID:     jobID,
		Owner:     "test-owner",
		Repo:      "test-repo",
		SHA:       "abc123",
		Status:    models.JobStatusRunning,
		Progress:  models.NewJobProgress(),
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, c.SetJob(ctx, jobID, &initialJob))

	js := skipIfNoNATS(t)
	analyzer := core.NewAnalyzer("/tmp/grit-test", 51200)
	w := NewWorker(js, analyzer, c, nil)

	now := time.Now().UTC()
	w.updateJobCompletion(ctx, jobID, models.JobStatusCompleted, &now, "")

	data, err := c.GetJob(ctx, jobID)
	require.NoError(t, err)
	var job models.AnalysisJob
	require.NoError(t, json.Unmarshal(data, &job))

	assert.Equal(t, models.JobStatusCompleted, job.Status)
	assert.NotNil(t, job.CompletedAt)
	assert.Empty(t, job.Error)
	assert.Equal(t, "/api/test-owner/test-repo", job.ResultURL)
	assert.Equal(t, models.SubJobCompleted, job.Progress.Clone)
	assert.Equal(t, models.SubJobCompleted, job.Progress.FileWalk)
	assert.Equal(t, models.SubJobCompleted, job.Progress.MetadataFetch)
	assert.Equal(t, models.SubJobCompleted, job.Progress.CommitActivityFetch)
}

func TestWorker_TimeoutSetsFailedStatus(t *testing.T) {
	c := skipIfNoRedis(t)
	ctx := context.Background()

	jobID := "test-job-timeout-" + time.Now().Format("150405")
	initialJob := models.AnalysisJob{
		JobID:     jobID,
		Owner:     "test-owner",
		Repo:      "test-repo",
		SHA:       "abc123",
		Status:    models.JobStatusRunning,
		Progress:  models.NewJobProgress(),
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, c.SetJob(ctx, jobID, &initialJob))

	js := skipIfNoNATS(t)
	analyzer := core.NewAnalyzer("/tmp/grit-test", 51200)
	w := NewWorker(js, analyzer, c, nil)

	now := time.Now().UTC()
	w.updateJobCompletion(ctx, jobID, models.JobStatusFailed, &now, "context deadline exceeded")

	data, err := c.GetJob(ctx, jobID)
	require.NoError(t, err)
	var job models.AnalysisJob
	require.NoError(t, json.Unmarshal(data, &job))

	assert.Equal(t, models.JobStatusFailed, job.Status)
	assert.Contains(t, job.Error, "context deadline exceeded")
	assert.NotNil(t, job.CompletedAt)
	assert.Empty(t, job.ResultURL, "timed out jobs should not have result_url")
}

func TestWorker_UpdateJobCompletion_Failed(t *testing.T) {
	c := skipIfNoRedis(t)
	ctx := context.Background()

	jobID := "test-job-failed-" + time.Now().Format("150405")
	initialJob := models.AnalysisJob{
		JobID:     jobID,
		Owner:     "test-owner",
		Repo:      "test-repo",
		SHA:       "abc123",
		Status:    models.JobStatusRunning,
		Progress:  models.NewJobProgress(),
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, c.SetJob(ctx, jobID, &initialJob))

	js := skipIfNoNATS(t)
	analyzer := core.NewAnalyzer("/tmp/grit-test", 51200)
	w := NewWorker(js, analyzer, c, nil)

	now := time.Now().UTC()
	w.updateJobCompletion(ctx, jobID, models.JobStatusFailed, &now, "clone failed: timeout")

	data, err := c.GetJob(ctx, jobID)
	require.NoError(t, err)
	var job models.AnalysisJob
	require.NoError(t, json.Unmarshal(data, &job))

	assert.Equal(t, models.JobStatusFailed, job.Status)
	assert.NotNil(t, job.CompletedAt)
	assert.Equal(t, "clone failed: timeout", job.Error)
	assert.Empty(t, job.ResultURL, "failed jobs should not have result_url")
}

// --- Complexity Worker Tests ---

func TestComplexityWorker_UpdateJobStatus(t *testing.T) {
	c := skipIfNoRedis(t)
	ctx := context.Background()

	jobID := "test-cw-status-" + time.Now().Format("150405")
	initialJob := models.AnalysisJob{
		JobID:     jobID,
		Owner:     "test-owner",
		Repo:      "test-repo",
		SHA:       "abc123",
		Status:    models.JobStatusQueued,
		Progress:  models.NewJobProgress(),
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, c.SetJob(ctx, jobID, &initialJob))

	js := skipIfNoNATS(t)
	cw := NewComplexityWorker(js, nil, c, "/tmp/grit-test", nil)

	cw.updateJobStatus(ctx, jobID, models.JobStatusRunning)

	data, err := c.GetJob(ctx, jobID)
	require.NoError(t, err)

	var job models.AnalysisJob
	require.NoError(t, json.Unmarshal(data, &job))
	assert.Equal(t, models.JobStatusRunning, job.Status)
}

func TestComplexityWorker_UpdateJobCompletion_Success(t *testing.T) {
	c := skipIfNoRedis(t)
	ctx := context.Background()

	jobID := "test-cw-complete-" + time.Now().Format("150405")
	initialJob := models.AnalysisJob{
		JobID:     jobID,
		Owner:     "test-owner",
		Repo:      "test-repo",
		SHA:       "abc123",
		Status:    models.JobStatusRunning,
		Progress:  models.NewJobProgress(),
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, c.SetJob(ctx, jobID, &initialJob))

	js := skipIfNoNATS(t)
	cw := NewComplexityWorker(js, nil, c, "/tmp/grit-test", nil)

	now := time.Now().UTC()
	cw.updateJobCompletion(ctx, jobID, models.JobStatusCompleted, &now, "")

	data, err := c.GetJob(ctx, jobID)
	require.NoError(t, err)
	var job models.AnalysisJob
	require.NoError(t, json.Unmarshal(data, &job))

	assert.Equal(t, models.JobStatusCompleted, job.Status)
	assert.NotNil(t, job.CompletedAt)
	assert.Empty(t, job.Error)
	assert.Equal(t, "/api/test-owner/test-repo/complexity", job.ResultURL)
}

func TestComplexityWorker_UpdateJobCompletion_Failed(t *testing.T) {
	c := skipIfNoRedis(t)
	ctx := context.Background()

	jobID := "test-cw-failed-" + time.Now().Format("150405")
	initialJob := models.AnalysisJob{
		JobID:     jobID,
		Owner:     "test-owner",
		Repo:      "test-repo",
		SHA:       "abc123",
		Status:    models.JobStatusRunning,
		Progress:  models.NewJobProgress(),
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, c.SetJob(ctx, jobID, &initialJob))

	js := skipIfNoNATS(t)
	cw := NewComplexityWorker(js, nil, c, "/tmp/grit-test", nil)

	now := time.Now().UTC()
	cw.updateJobCompletion(ctx, jobID, models.JobStatusFailed, &now, "core analysis result not found")

	data, err := c.GetJob(ctx, jobID)
	require.NoError(t, err)
	var job models.AnalysisJob
	require.NoError(t, json.Unmarshal(data, &job))

	assert.Equal(t, models.JobStatusFailed, job.Status)
	assert.NotNil(t, job.CompletedAt)
	assert.Equal(t, "core analysis result not found", job.Error)
	assert.Empty(t, job.ResultURL, "failed complexity jobs should not have result_url")
}

func TestPublisher_PublishComplexity_DeduplicatesActiveJob(t *testing.T) {
	c := skipIfNoRedis(t)
	ctx := context.Background()

	js := skipIfNoNATS(t)
	_ = EnsureStream(js)
	pub := NewPublisher(js, c)

	// First publish should create a new job.
	jobID1, err := pub.PublishComplexity(ctx, "dedup-owner", "dedup-repo", "sha1", "")
	require.NoError(t, err)
	assert.NotEmpty(t, jobID1)

	// Second publish with same params should return existing job.
	jobID2, err := pub.PublishComplexity(ctx, "dedup-owner", "dedup-repo", "sha1", "")
	require.NoError(t, err)
	assert.Equal(t, jobID1, jobID2)

	// Cleanup.
	c.DeleteActiveComplexityJob(ctx, "dedup-owner", "dedup-repo", "sha1")
}
