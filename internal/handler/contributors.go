package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/grit-app/grit/internal/cache"
	"github.com/grit-app/grit/internal/models"
)

// ContributorHandler serves contributor/blame analysis results.
type ContributorHandler struct {
	cache *cache.Cache
}

// NewContributorHandler creates a new contributor handler.
func NewContributorHandler(c *cache.Cache) *ContributorHandler {
	return &ContributorHandler{cache: c}
}

// HandleContributors handles GET /api/:owner/:repo/contributors.
func (h *ContributorHandler) HandleContributors(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	if !ownerRepoRegex.MatchString(owner) || !ownerRepoRegex.MatchString(repo) {
		WriteBadRequest(w, fmt.Sprintf("Invalid repository identifier: %s/%s", owner, repo))
		return
	}

	// Check contributor cache.
	sha := r.URL.Query().Get("sha")
	data, err := h.cache.GetContributors(r.Context(), owner, repo, sha)
	if err == nil && data != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	// Check if there's an active blame job.
	jobID, err := h.cache.GetActiveBlameJob(r.Context(), owner, repo, "")
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
		WriteNotFound(w, "Core analysis must complete before contributor data is available.")
		return
	}

	// Core exists but no contributor data — return 404.
	WriteNotFound(w, fmt.Sprintf("Contributor analysis not yet available for %s/%s.", owner, repo))
}

// HandleContributorFiles handles GET /api/:owner/:repo/contributors/files.
func (h *ContributorHandler) HandleContributorFiles(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	if !ownerRepoRegex.MatchString(owner) || !ownerRepoRegex.MatchString(repo) {
		WriteBadRequest(w, fmt.Sprintf("Invalid repository identifier: %s/%s", owner, repo))
		return
	}

	// Check contributor cache.
	sha := r.URL.Query().Get("sha")
	data, err := h.cache.GetContributors(r.Context(), owner, repo, sha)
	if err == nil && data != nil {
		// Extract only the file_contributors from the full result.
		var result models.ContributorResult
		if json.Unmarshal(data, &result) == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"repository":        result.Repository,
				"file_contributors": result.FileContributors,
				"total_files":       result.TotalFilesAnalyzed,
				"partial":           result.Partial,
			})
			return
		}
	}

	// Check if there's an active blame job.
	jobID, err := h.cache.GetActiveBlameJob(r.Context(), owner, repo, "")
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
		WriteNotFound(w, "Core analysis must complete before contributor file data is available.")
		return
	}

	WriteNotFound(w, fmt.Sprintf("Contributor file analysis not yet available for %s/%s.", owner, repo))
}
