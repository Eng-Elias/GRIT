package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/grit-app/grit/internal/cache"
	"github.com/grit-app/grit/internal/models"
)

// TemporalHandler serves temporal analysis results.
type TemporalHandler struct {
	cache *cache.Cache
}

// NewTemporalHandler creates a new temporal handler.
func NewTemporalHandler(c *cache.Cache) *TemporalHandler {
	return &TemporalHandler{cache: c}
}

// HandleTemporal handles GET /api/:owner/:repo/temporal.
func (h *TemporalHandler) HandleTemporal(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	if !ownerRepoRegex.MatchString(owner) || !ownerRepoRegex.MatchString(repo) {
		WriteBadRequest(w, fmt.Sprintf("Invalid repository identifier: %s/%s", owner, repo))
		return
	}

	sha := r.URL.Query().Get("sha")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "3y"
	}

	// Validate period
	switch period {
	case "1y", "2y", "3y":
		// valid
	default:
		WriteBadRequest(w, fmt.Sprintf("Invalid period: %s. Must be 1y, 2y, or 3y.", period))
		return
	}

	// Check temporal cache.
	data, err := h.cache.GetTemporal(r.Context(), owner, repo, sha, period)
	if err == nil && data != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	// Check if there's an active temporal job.
	jobID, err := h.cache.GetActiveTemporalJob(r.Context(), owner, repo, "")
	if err == nil && jobID != "" {
		jobData, err := h.cache.GetJob(r.Context(), jobID)
		if err == nil {
			var job models.AnalysisJob
			if json.Unmarshal(jobData, &job) == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				json.NewEncoder(w).Encode(map[string]string{
					"job_id":   jobID,
					"status":   string(job.Status),
					"poll_url": fmt.Sprintf("/api/%s/%s/status", owner, repo),
				})
				return
			}
		}
	}

	// Check if core analysis exists — if not, return 404.
	coreData, coreErr := h.cache.GetAnalysis(r.Context(), owner, repo, sha)
	if coreErr != nil || coreData == nil {
		WriteNotFound(w, "Core analysis must complete before temporal data is available.")
		return
	}

	// Core exists but no temporal — return 404.
	WriteNotFound(w, fmt.Sprintf("Temporal analysis not yet available for %s/%s.", owner, repo))
}
