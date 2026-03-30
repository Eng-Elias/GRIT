package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/grit-app/grit/internal/cache"
	"github.com/grit-app/grit/internal/models"
)

// ComplexityHandler serves complexity analysis results.
type ComplexityHandler struct {
	cache *cache.Cache
}

// NewComplexityHandler creates a new complexity handler.
func NewComplexityHandler(c *cache.Cache) *ComplexityHandler {
	return &ComplexityHandler{cache: c}
}

// HandleComplexity handles GET /api/:owner/:repo/complexity.
func (h *ComplexityHandler) HandleComplexity(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	if !ownerRepoRegex.MatchString(owner) || !ownerRepoRegex.MatchString(repo) {
		WriteBadRequest(w, fmt.Sprintf("Invalid repository identifier: %s/%s", owner, repo))
		return
	}

	// Check complexity cache.
	_, complexityData, err := h.tryComplexityCache(r, owner, repo)
	if err == nil && complexityData != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(http.StatusOK)
		w.Write(complexityData)
		return
	}

	// Check if there's an active complexity job.
	jobID, err := h.cache.GetActiveComplexityJob(r.Context(), owner, repo, "")
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
	_, coreData, coreErr := h.tryCoreCache(r, owner, repo)
	if coreErr != nil || coreData == nil {
		WriteNotFound(w, "Core analysis must complete before complexity data is available.")
		return
	}

	// Core exists but no complexity — return 404.
	WriteNotFound(w, fmt.Sprintf("Complexity analysis not yet available for %s/%s.", owner, repo))
}

func (h *ComplexityHandler) tryComplexityCache(r *http.Request, owner, repo string) (string, []byte, error) {
	sha := r.URL.Query().Get("sha")
	data, err := h.cache.GetComplexity(r.Context(), owner, repo, sha)
	if err != nil {
		return "", nil, err
	}
	return sha, data, nil
}

func (h *ComplexityHandler) tryCoreCache(r *http.Request, owner, repo string) (string, []byte, error) {
	sha := r.URL.Query().Get("sha")
	data, err := h.cache.GetAnalysis(r.Context(), owner, repo, sha)
	if err != nil {
		return "", nil, err
	}
	return sha, data, nil
}
