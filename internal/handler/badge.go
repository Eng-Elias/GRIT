package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/grit-app/grit/internal/cache"
	"github.com/grit-app/grit/internal/models"
)

type BadgeHandler struct {
	cache *cache.Cache
}

func NewBadgeHandler(c *cache.Cache) *BadgeHandler {
	return &BadgeHandler{cache: c}
}

type BadgeResponse struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

func (h *BadgeHandler) HandleBadge(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	if !ownerRepoRegex.MatchString(owner) || !ownerRepoRegex.MatchString(repo) {
		WriteBadRequest(w, fmt.Sprintf("Invalid repository identifier: %s/%s", owner, repo))
		return
	}

	data, err := h.cache.GetAnalysis(r.Context(), owner, repo, "")
	if err != nil {
		WriteJSON(w, http.StatusOK, BadgeResponse{
			SchemaVersion: 1,
			Label:         "lines of code",
			Message:       "analyzing...",
			Color:         "lightgrey",
		})
		return
	}

	var result models.AnalysisResult
	if err := json.Unmarshal(data, &result); err != nil {
		WriteJSON(w, http.StatusOK, BadgeResponse{
			SchemaVersion: 1,
			Label:         "lines of code",
			Message:       "error",
			Color:         "red",
		})
		return
	}

	WriteJSON(w, http.StatusOK, BadgeResponse{
		SchemaVersion: 1,
		Label:         "lines of code",
		Message:       FormatLineCount(result.TotalLines),
		Color:         ColorForLineCount(result.TotalLines),
	})
}

func FormatLineCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func ColorForLineCount(n int) string {
	switch {
	case n < 1_000:
		return "brightgreen"
	case n < 10_000:
		return "green"
	case n < 100_000:
		return "yellowgreen"
	case n < 500_000:
		return "yellow"
	case n < 1_000_000:
		return "orange"
	default:
		return "blue"
	}
}
