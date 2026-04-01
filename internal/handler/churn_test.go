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

func skipIfNoRedisChurn(t *testing.T) *cache.Cache {
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

func TestHandleChurnMatrix_Returns200WithCachedData(t *testing.T) {
	c := skipIfNoRedisChurn(t)
	ctx := context.Background()

	result := models.ChurnMatrixResult{
		Repository:    models.Repository{Owner: "churn-owner", Name: "churn-repo"},
		TotalCommits:  100,
		CriticalCount: 3,
		StaleCount:    2,
	}
	data, err := json.Marshal(result)
	require.NoError(t, err)
	require.NoError(t, c.SetChurn(ctx, "churn-owner", "churn-repo", "sha123", data))
	defer c.DeleteChurn(ctx, "churn-owner", "churn-repo")

	h := NewChurnHandler(c)
	r := chi.NewRouter()
	r.Get("/api/{owner}/{repo}/churn-matrix", h.HandleChurnMatrix)

	req := httptest.NewRequest(http.MethodGet, "/api/churn-owner/churn-repo/churn-matrix", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "HIT", w.Header().Get("X-Cache"))
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestHandleChurnMatrix_Returns404WhenNoCoreAnalysis(t *testing.T) {
	c := skipIfNoRedisChurn(t)
	ctx := context.Background()
	c.DeleteAnalysis(ctx, "no-churn-owner", "no-churn-repo")
	c.DeleteChurn(ctx, "no-churn-owner", "no-churn-repo")

	h := NewChurnHandler(c)
	r := chi.NewRouter()
	r.Get("/api/{owner}/{repo}/churn-matrix", h.HandleChurnMatrix)

	req := httptest.NewRequest(http.MethodGet, "/api/no-churn-owner/no-churn-repo/churn-matrix", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleChurnMatrix_Returns400ForInvalidParams(t *testing.T) {
	c := skipIfNoRedisChurn(t)

	h := NewChurnHandler(c)
	r := chi.NewRouter()
	r.Get("/api/{owner}/{repo}/churn-matrix", h.HandleChurnMatrix)

	req := httptest.NewRequest(http.MethodGet, "/api/bad%20owner/repo/churn-matrix", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleChurnMatrix_Returns202WhenJobInProgress(t *testing.T) {
	c := skipIfNoRedisChurn(t)
	ctx := context.Background()

	jobID := "test-churn-job-456"
	job := models.AnalysisJob{
		JobID:  jobID,
		Owner:  "churn-progress-owner",
		Repo:   "churn-progress-repo",
		Status: models.JobStatusRunning,
	}
	require.NoError(t, c.SetJob(ctx, jobID, &job))
	require.NoError(t, c.SetActiveChurnJob(ctx, "churn-progress-owner", "churn-progress-repo", "", jobID))
	defer func() {
		c.DeleteChurn(ctx, "churn-progress-owner", "churn-progress-repo")
		c.DeleteActiveChurnJob(ctx, "churn-progress-owner", "churn-progress-repo", "")
	}()

	h := NewChurnHandler(c)
	r := chi.NewRouter()
	r.Get("/api/{owner}/{repo}/churn-matrix", h.HandleChurnMatrix)

	req := httptest.NewRequest(http.MethodGet, "/api/churn-progress-owner/churn-progress-repo/churn-matrix", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
}
