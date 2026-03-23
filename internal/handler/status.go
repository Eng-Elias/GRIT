package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/grit-app/grit/internal/cache"
	"github.com/grit-app/grit/internal/models"
)

type StatusHandler struct {
	cache *cache.Cache
}

func NewStatusHandler(c *cache.Cache) *StatusHandler {
	return &StatusHandler{cache: c}
}

func (h *StatusHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	if !ownerRepoRegex.MatchString(owner) || !ownerRepoRegex.MatchString(repo) {
		WriteBadRequest(w, fmt.Sprintf("Invalid repository identifier: %s/%s", owner, repo))
		return
	}

	jobID, err := h.findActiveJobID(r.Context(), owner, repo)
	if err != nil {
		WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "not_found",
			"message": fmt.Sprintf("No analysis in progress. Request GET /api/%s/%s to start one.", owner, repo),
		})
		return
	}

	data, err := h.cache.GetJob(r.Context(), jobID)
	if err != nil {
		WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "not_found",
			"message": fmt.Sprintf("No analysis in progress. Request GET /api/%s/%s to start one.", owner, repo),
		})
		return
	}

	var job models.AnalysisJob
	if err := json.Unmarshal(data, &job); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to parse job state")
		return
	}

	WriteJSON(w, http.StatusOK, &job)
}

func (h *StatusHandler) findActiveJobID(ctx context.Context, owner, repo string) (string, error) {
	jobID, err := h.cache.GetActiveJob(ctx, owner, repo, "")
	if err != nil {
		return "", err
	}
	return jobID, nil
}
