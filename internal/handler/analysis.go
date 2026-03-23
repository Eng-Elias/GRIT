package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"

	"github.com/grit-app/grit/internal/cache"
	ghclient "github.com/grit-app/grit/internal/github"
	"github.com/grit-app/grit/internal/job"
	gritmetrics "github.com/grit-app/grit/internal/metrics"
	"github.com/grit-app/grit/internal/models"
)

var ownerRepoRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

type AnalysisHandler struct {
	cache     *cache.Cache
	publisher *job.Publisher
	ghToken   string
}

func NewAnalysisHandler(c *cache.Cache, pub *job.Publisher, ghToken string) *AnalysisHandler {
	return &AnalysisHandler{
		cache:     c,
		publisher: pub,
		ghToken:   ghToken,
	}
}

func (h *AnalysisHandler) HandleAnalysis(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	if !ownerRepoRegex.MatchString(owner) || !ownerRepoRegex.MatchString(repo) {
		WriteBadRequest(w, fmt.Sprintf("Invalid repository identifier: %s/%s", owner, repo))
		return
	}

	if len(owner)+len(repo) > 256 {
		WriteBadRequest(w, "Repository identifier too long")
		return
	}

	token := extractToken(r)

	sha, cachedData, err := h.tryCache(r.Context(), owner, repo)
	if err == nil && cachedData != nil {
		gritmetrics.CacheHitTotal.Inc()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(http.StatusOK)
		w.Write(cachedData)
		return
	}

	gritmetrics.CacheMissTotal.Inc()

	if sha == "" {
		ghClient := ghclient.NewClient(token)
		_, _, defaultBranch, err := ghClient.FetchMetadata(r.Context(), owner, repo)
		if err != nil {
			handleGitHubError(w, err, owner, repo)
			return
		}
		sha = defaultBranch
	}

	if err := h.cache.Ping(r.Context()); err != nil {
		slog.Warn("redis unavailable, cannot enqueue job", "error", err)
		WriteServiceUnavailable(w, "Cache service unavailable")
		return
	}

	jobID, err := h.publisher.Publish(r.Context(), owner, repo, sha, token)
	if err != nil {
		slog.Error("failed to publish job", "error", err)
		WriteServiceUnavailable(w, "Job queue unavailable")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"job_id":   jobID,
		"status":   string(models.JobStatusQueued),
		"poll_url": fmt.Sprintf("/api/%s/%s/status", owner, repo),
	})
}

func (h *AnalysisHandler) tryCache(ctx context.Context, owner, repo string) (string, []byte, error) {
	data, err := h.cache.GetAnalysis(ctx, owner, repo, "")
	if err != nil {
		return "", nil, err
	}

	var result models.AnalysisResult
	if err := json.Unmarshal(data, &result); err != nil {
		return "", nil, cache.ErrCacheMiss
	}

	result.Cached = true
	now := result.AnalyzedAt
	result.CachedAt = &now

	enriched, err := json.Marshal(result)
	if err != nil {
		return result.Repository.LatestSHA, data, nil
	}

	return result.Repository.LatestSHA, enriched, nil
}

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}

func handleGitHubError(w http.ResponseWriter, err error, owner, repo string) {
	var rateLimitErr *ghclient.RateLimitError
	var notFoundErr *ghclient.NotFoundError
	var privateErr *ghclient.PrivateRepoError

	switch {
	case errors.As(err, &rateLimitErr):
		WriteRateLimited(w, rateLimitErr.RetryAfter)
	case errors.As(err, &notFoundErr):
		WriteNotFound(w, fmt.Sprintf("Repository %s/%s does not exist.", owner, repo))
	case errors.As(err, &privateErr):
		WritePrivateRepo(w)
	default:
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
	}
}
