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

// AISummaryHandler serves the AI codebase summary endpoint.
type AISummaryHandler struct {
	cache    *cache.Cache
	client   *ai.Client
	cloneDir string
}

// NewAISummaryHandler creates a handler for the AI summary endpoint. client
// may be nil if GEMINI_API_KEY was not configured; the handler will return
// 503 in that case.
func NewAISummaryHandler(c *cache.Cache, client *ai.Client, cloneDir string) *AISummaryHandler {
	return &AISummaryHandler{cache: c, client: client, cloneDir: cloneDir}
}

// HandleAISummary streams an AI-generated codebase summary via SSE, or
// returns a cached result with X-Cache: HIT.
func (h *AISummaryHandler) HandleAISummary(w http.ResponseWriter, r *http.Request) {
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

	// Check core analysis cache to get SHA (409 if missing).
	coreData, err := h.cache.GetAnalysis(r.Context(), owner, repo, "")
	if err != nil {
		WriteError(w, http.StatusConflict, "analysis_pending",
			"Core analysis must complete before AI summary is available")
		return
	}
	var coreResult models.AnalysisResult
	if err := json.Unmarshal(coreData, &coreResult); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to read analysis data")
		return
	}
	sha := coreResult.Repository.LatestSHA

	// Check AI summary cache — return cached if present.
	cachedData, err := h.cache.GetAISummary(r.Context(), owner, repo, sha)
	if err == nil {
		var summary models.AISummary
		if json.Unmarshal(cachedData, &summary) == nil {
			now := time.Now().UTC()
			summary.Cached = true
			summary.CachedAt = &now
			w.Header().Set("X-Cache", "HIT")
			gritmetrics.AIRequestTotal.WithLabelValues("summary", "cache_hit").Inc()
			WriteJSON(w, http.StatusOK, summary)
			return
		}
	}

	// Stream new summary via SSE.
	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Streaming not supported")
		return
	}

	cloneDir := fmt.Sprintf("%s/%s/%s", h.cloneDir, owner, repo)
	contextParts, err := ai.AssembleContext(r.Context(), h.cache, owner, repo, sha, cloneDir)
	if err != nil {
		slog.Error("ai: assemble context failed", "owner", owner, "repo", repo, "error", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to build AI context")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Cache", "MISS")

	summary, err := ai.GenerateSummary(r.Context(), h.client, contextParts, coreResult.Repository, func(chunk string) {
		fmt.Fprint(w, ai.FormatSSEEvent(chunk))
		flusher.Flush()
	})
	if err != nil {
		slog.Error("ai: summary generation failed", "owner", owner, "repo", repo, "error", err)
		fmt.Fprint(w, ai.FormatSSEEvent("[ERROR] AI summary generation failed"))
		flusher.Flush()
		gritmetrics.AIRequestTotal.WithLabelValues("summary", "error").Inc()
		return
	}

	// Send final structured result.
	finalJSON, _ := json.Marshal(summary)
	fmt.Fprintf(w, "event: done\ndata: %s\n\n", string(finalJSON))
	flusher.Flush()

	// Cache the result.
	if err := h.cache.SetAISummary(r.Context(), owner, repo, sha, finalJSON); err != nil {
		slog.Error("ai: failed to cache summary", "owner", owner, "repo", repo, "error", err)
	}

	gritmetrics.AIRequestTotal.WithLabelValues("summary", "success").Inc()
	gritmetrics.AIRequestDuration.WithLabelValues("summary").Observe(time.Since(start).Seconds())
}
