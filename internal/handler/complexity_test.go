package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grit-app/grit/internal/cache"
	"github.com/grit-app/grit/internal/models"
)

func skipIfNoRedisComplexity(t *testing.T) *cache.Cache {
	t.Helper()
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}
	c, err := cache.New(redisURL)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	return c
}

func TestHandleComplexity_Returns200WithCachedData(t *testing.T) {
	c := skipIfNoRedisComplexity(t)
	ctx := context.Background()

	result := models.ComplexityResult{
		Repository:         models.Repository{Owner: "test-owner", Name: "test-repo"},
		TotalFilesAnalyzed: 5,
		TotalFunctionCount: 20,
		MeanComplexity:     3.5,
	}
	data, err := json.Marshal(result)
	require.NoError(t, err)
	require.NoError(t, c.SetComplexity(ctx, "test-owner", "test-repo", "abc123", data))
	defer c.DeleteComplexity(ctx, "test-owner", "test-repo")

	h := NewComplexityHandler(c)
	r := chi.NewRouter()
	r.Get("/api/{owner}/{repo}/complexity", h.HandleComplexity)

	req := httptest.NewRequest(http.MethodGet, "/api/test-owner/test-repo/complexity", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "HIT", w.Header().Get("X-Cache"))
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestHandleComplexity_Returns404WhenNoCoreAnalysis(t *testing.T) {
	c := skipIfNoRedisComplexity(t)
	ctx := context.Background()
	// Ensure no data exists.
	c.DeleteAnalysis(ctx, "no-core-owner", "no-core-repo")
	c.DeleteComplexity(ctx, "no-core-owner", "no-core-repo")

	h := NewComplexityHandler(c)
	r := chi.NewRouter()
	r.Get("/api/{owner}/{repo}/complexity", h.HandleComplexity)

	req := httptest.NewRequest(http.MethodGet, "/api/no-core-owner/no-core-repo/complexity", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleComplexity_Returns400ForInvalidParams(t *testing.T) {
	c := skipIfNoRedisComplexity(t)

	h := NewComplexityHandler(c)
	r := chi.NewRouter()
	r.Get("/api/{owner}/{repo}/complexity", h.HandleComplexity)

	req := httptest.NewRequest(http.MethodGet, "/api/bad%20owner/repo/complexity", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleComplexity_Returns202WhenJobInProgress(t *testing.T) {
	c := skipIfNoRedisComplexity(t)
	ctx := context.Background()

	// Set up an active complexity job.
	jobID := "test-complexity-job-123"
	job := models.AnalysisJob{
		JobID:  jobID,
		Owner:  "progress-owner",
		Repo:   "progress-repo",
		Status: models.JobStatusRunning,
	}
	jobData, _ := json.Marshal(job)
	require.NoError(t, c.SetJob(ctx, jobID, &job))
	require.NoError(t, c.SetActiveComplexityJob(ctx, "progress-owner", "progress-repo", "", jobID))
	defer func() {
		c.DeleteComplexity(ctx, "progress-owner", "progress-repo")
		c.DeleteActiveComplexityJob(ctx, "progress-owner", "progress-repo", "")
	}()
	_ = jobData

	h := NewComplexityHandler(c)
	r := chi.NewRouter()
	r.Get("/api/{owner}/{repo}/complexity", h.HandleComplexity)

	req := httptest.NewRequest(http.MethodGet, "/api/progress-owner/progress-repo/complexity", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
}
