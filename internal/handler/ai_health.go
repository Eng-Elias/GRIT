package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/grit-app/grit/internal/ai"
	"github.com/grit-app/grit/internal/cache"
	gritmetrics "github.com/grit-app/grit/internal/metrics"
	"github.com/grit-app/grit/internal/models"
)

// AIHealthHandler serves the AI health score endpoint.
type AIHealthHandler struct {
	cache    *cache.Cache
	client   *ai.Client
	cloneDir string
}

// NewAIHealthHandler creates a handler for the AI health endpoint.
func NewAIHealthHandler(c *cache.Cache, client *ai.Client, cloneDir string) *AIHealthHandler {
	return &AIHealthHandler{cache: c, client: client, cloneDir: cloneDir}
}

// HandleAIHealth returns a structured JSON health assessment, cached for 6h.
func (h *AIHealthHandler) HandleAIHealth(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	if !ownerRepoRegex.MatchString(owner) || !ownerRepoRegex.MatchString(repo) {
		WriteBadRequest(w, fmt.Sprintf("Invalid repository identifier: %s/%s", owner, repo))
		return
	}

	// 503 — AI not configured.
	if h.client == nil {
		WriteServiceUnavailable(w, "AI features are not available: GEMINI_API_KEY not configured")
		return
	}

	// Check core analysis cache (409 if missing).
	coreData, err := h.cache.GetAnalysis(r.Context(), owner, repo, "")
	if err != nil {
		WriteError(w, http.StatusConflict, "analysis_pending",
			"Core analysis must complete before AI health score is available")
		return
	}
	var coreResult models.AnalysisResult
	if err := json.Unmarshal(coreData, &coreResult); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to read analysis data")
		return
	}
	sha := coreResult.Repository.LatestSHA

	// Check AI health cache — return cached if present.
	cachedData, err := h.cache.GetAIHealth(r.Context(), owner, repo, sha)
	if err == nil {
		var hs models.HealthScore
		if json.Unmarshal(cachedData, &hs) == nil {
			now := time.Now().UTC()
			hs.Cached = true
			hs.CachedAt = &now
			w.Header().Set("X-Cache", "HIT")
			gritmetrics.AIRequestTotal.WithLabelValues("health", "cache_hit").Inc()
			WriteJSON(w, http.StatusOK, hs)
			return
		}
	}

	// Generate new health score.
	cloneDir := fmt.Sprintf("%s/%s/%s", h.cloneDir, owner, repo)
	contextParts, err := ai.AssembleContext(r.Context(), h.cache, owner, repo, sha, cloneDir)
	if err != nil {
		slog.Error("ai: assemble context failed", "owner", owner, "repo", repo, "error", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to build AI context")
		return
	}

	hs, err := ai.GenerateHealth(r.Context(), h.client, contextParts, coreResult.Repository)
	if err != nil {
		slog.Error("ai: health generation failed", "owner", owner, "repo", repo, "error", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "AI health score generation failed")
		gritmetrics.AIRequestTotal.WithLabelValues("health", "error").Inc()
		return
	}

	// Cache the result.
	resultJSON, _ := json.Marshal(hs)
	if err := h.cache.SetAIHealth(r.Context(), owner, repo, sha, resultJSON); err != nil {
		slog.Error("ai: failed to cache health score", "owner", owner, "repo", repo, "error", err)
	}

	w.Header().Set("X-Cache", "MISS")
	gritmetrics.AIRequestTotal.WithLabelValues("health", "success").Inc()
	gritmetrics.AIRequestDuration.WithLabelValues("health").Observe(time.Since(start).Seconds())
	WriteJSON(w, http.StatusOK, hs)
}
