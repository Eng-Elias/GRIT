package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/grit-app/grit/internal/cache"
	"github.com/grit-app/grit/internal/models"
)

// ChurnHandler serves churn analysis results.
type ChurnHandler struct {
	cache *cache.Cache
}

// NewChurnHandler creates a new churn handler.
func NewChurnHandler(c *cache.Cache) *ChurnHandler {
	return &ChurnHandler{cache: c}
}

// HandleChurnMatrix handles GET /api/:owner/:repo/churn-matrix.
func (h *ChurnHandler) HandleChurnMatrix(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	if !ownerRepoRegex.MatchString(owner) || !ownerRepoRegex.MatchString(repo) {
		WriteBadRequest(w, fmt.Sprintf("Invalid repository identifier: %s/%s", owner, repo))
		return
	}

	// Check churn cache.
	sha := r.URL.Query().Get("sha")
	data, err := h.cache.GetChurn(r.Context(), owner, repo, sha)
	if err == nil && data != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	// Check if there's an active churn job.
	jobID, err := h.cache.GetActiveChurnJob(r.Context(), owner, repo, "")
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
		WriteNotFound(w, "Core analysis must complete before churn data is available.")
		return
	}

	// Core exists but no churn — return 404.
	WriteNotFound(w, fmt.Sprintf("Churn analysis not yet available for %s/%s.", owner, repo))
}
