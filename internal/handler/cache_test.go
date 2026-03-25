package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grit-app/grit/internal/cache"
)

func redisURL() string {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "redis://localhost:6379"
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

func TestHandleDeleteCache_Returns204(t *testing.T) {
	c := skipIfNoRedis(t)
	ctx := context.Background()

	require.NoError(t, c.SetAnalysis(ctx, "test-owner", "test-repo", "abc123", []byte(`{"total_lines":42}`)))

	h := NewCacheHandler(c)
	r := chi.NewRouter()
	r.Delete("/api/{owner}/{repo}/cache", h.HandleDeleteCache)

	req := httptest.NewRequest(http.MethodDelete, "/api/test-owner/test-repo/cache", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	_, err := c.GetAnalysis(ctx, "test-owner", "test-repo", "abc123")
	assert.ErrorIs(t, err, cache.ErrCacheMiss, "cache should be cleared after delete")
}

func TestHandleDeleteCache_Idempotent(t *testing.T) {
	c := skipIfNoRedis(t)

	h := NewCacheHandler(c)
	r := chi.NewRouter()
	r.Delete("/api/{owner}/{repo}/cache", h.HandleDeleteCache)

	req := httptest.NewRequest(http.MethodDelete, "/api/nonexistent-owner/nonexistent-repo/cache", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code, "DELETE on missing cache should still return 204")
}

func TestHandleDeleteCache_InvalidOwner(t *testing.T) {
	c := skipIfNoRedis(t)

	h := NewCacheHandler(c)
	r := chi.NewRouter()
	r.Delete("/api/{owner}/{repo}/cache", h.HandleDeleteCache)

	req := httptest.NewRequest(http.MethodDelete, "/api/bad%20owner/repo/cache", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalysisHandler_CacheHit_XCacheHeader(t *testing.T) {
	c := skipIfNoRedis(t)
	ctx := context.Background()

	analysisJSON := `{"repository":{"owner":"cache-hit-owner","name":"cache-hit-repo","latest_sha":"abc123","default_branch":"main"},"analyzed_at":"2026-01-01T00:00:00Z","total_lines":100,"total_files":5,"cached":false}`
	require.NoError(t, c.SetAnalysis(ctx, "cache-hit-owner", "cache-hit-repo", "abc123", analysisJSON_bytes(analysisJSON)))

	h := NewAnalysisHandler(c, nil, "")
	r := chi.NewRouter()
	r.Get("/api/{owner}/{repo}", h.HandleAnalysis)

	req := httptest.NewRequest(http.MethodGet, "/api/cache-hit-owner/cache-hit-repo", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "HIT", w.Header().Get("X-Cache"))
	assert.Contains(t, w.Body.String(), `"cached":true`)
	assert.Contains(t, w.Body.String(), `"cached_at"`)
}

func analysisJSON_bytes(s string) []byte {
	return []byte(s)
}
