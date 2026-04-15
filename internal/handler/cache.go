package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/grit-app/grit/internal/cache"
)

type CacheHandler struct {
	cache *cache.Cache
}

func NewCacheHandler(c *cache.Cache) *CacheHandler {
	return &CacheHandler{cache: c}
}

func (h *CacheHandler) HandleDeleteCache(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	if !ownerRepoRegex.MatchString(owner) || !ownerRepoRegex.MatchString(repo) {
		WriteBadRequest(w, fmt.Sprintf("Invalid repository identifier: %s/%s", owner, repo))
		return
	}

	_ = h.cache.DeleteAnalysis(r.Context(), owner, repo)
	_ = h.cache.DeleteComplexity(r.Context(), owner, repo)
	_ = h.cache.DeleteChurn(r.Context(), owner, repo)
	_ = h.cache.DeleteContributors(r.Context(), owner, repo)
	_ = h.cache.DeleteTemporal(r.Context(), owner, repo)

	w.WriteHeader(http.StatusNoContent)
}
