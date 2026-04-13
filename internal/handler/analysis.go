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

		// Embed complexity_summary and churn_summary into core response.
		cachedData = h.embedComplexitySummary(r.Context(), owner, repo, cachedData)
		cachedData = h.embedChurnSummary(r.Context(), owner, repo, cachedData)
		cachedData = h.embedContributorSummary(r.Context(), owner, repo, cachedData)

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

func (h *AnalysisHandler) embedComplexitySummary(ctx context.Context, owner, repo string, coreData []byte) []byte {
	summary := models.ComplexitySummary{
		Status:        "pending",
		ComplexityURL: fmt.Sprintf("/api/%s/%s/complexity", owner, repo),
	}

	complexityData, err := h.cache.GetComplexity(ctx, owner, repo, "")
	if err == nil && complexityData != nil {
		var cr models.ComplexityResult
		if json.Unmarshal(complexityData, &cr) == nil {
			summary.Status = "complete"
			summary.MeanComplexity = cr.MeanComplexity
			summary.TotalFunctionCount = cr.TotalFunctionCount
			summary.HotFileCount = len(cr.HotFiles)
		}
	}

	// Merge complexity_summary into the core JSON object.
	var raw map[string]json.RawMessage
	if json.Unmarshal(coreData, &raw) != nil {
		return coreData
	}
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return coreData
	}
	raw["complexity_summary"] = summaryJSON
	merged, err := json.Marshal(raw)
	if err != nil {
		return coreData
	}
	return merged
}

func (h *AnalysisHandler) embedChurnSummary(ctx context.Context, owner, repo string, coreData []byte) []byte {
	summary := models.ChurnSummary{
		Status:         "pending",
		ChurnMatrixURL: fmt.Sprintf("/api/%s/%s/churn-matrix", owner, repo),
	}

	churnData, err := h.cache.GetChurn(ctx, owner, repo, "")
	if err == nil && churnData != nil {
		var cr models.ChurnMatrixResult
		if json.Unmarshal(churnData, &cr) == nil {
			summary.Status = "complete"
			summary.TotalFiles = cr.TotalFilesChurned
			summary.CriticalCount = cr.CriticalCount
			summary.StaleCount = cr.StaleCount
		}
	}

	var raw map[string]json.RawMessage
	if json.Unmarshal(coreData, &raw) != nil {
		return coreData
	}
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return coreData
	}
	raw["churn_summary"] = summaryJSON
	merged, err := json.Marshal(raw)
	if err != nil {
		return coreData
	}
	return merged
}

func (h *AnalysisHandler) embedContributorSummary(ctx context.Context, owner, repo string, coreData []byte) []byte {
	summary := models.ContributorSummary{
		Status:          "pending",
		ContributorsURL: fmt.Sprintf("/api/%s/%s/contributors", owner, repo),
	}

	contributorData, err := h.cache.GetContributors(ctx, owner, repo, "")
	if err == nil && contributorData != nil {
		var cr models.ContributorResult
		if json.Unmarshal(contributorData, &cr) == nil {
			summary.Status = "complete"
			summary.BusFactor = cr.BusFactor
			summary.TotalAuthors = len(cr.Authors)

			// Count active authors.
			activeCount := 0
			for _, a := range cr.Authors {
				if a.IsActive {
					activeCount++
				}
			}
			summary.ActiveAuthors = activeCount

			// Top 3 contributors as AuthorBrief.
			top := 3
			if len(cr.Authors) < top {
				top = len(cr.Authors)
			}
			briefs := make([]models.AuthorBrief, top)
			for i := 0; i < top; i++ {
				briefs[i] = models.AuthorBrief{
					Name:       cr.Authors[i].Name,
					Email:      cr.Authors[i].Email,
					LinesOwned: cr.Authors[i].TotalLinesOwned,
				}
			}
			summary.TopContributors = briefs
		}
	}

	var raw map[string]json.RawMessage
	if json.Unmarshal(coreData, &raw) != nil {
		return coreData
	}
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return coreData
	}
	raw["contributor_summary"] = summaryJSON
	merged, err := json.Marshal(raw)
	if err != nil {
		return coreData
	}
	return merged
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
