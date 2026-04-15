package handler

import (
	"encoding/json"
	"errors"
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

// AIChatHandler serves the AI chat endpoint.
type AIChatHandler struct {
	cache       *cache.Cache
	client      *ai.Client
	rateLimiter *ai.RateLimiter
	cloneDir    string
}

// NewAIChatHandler creates a handler for the AI chat endpoint. client may be
// nil if GEMINI_API_KEY was not configured.
func NewAIChatHandler(c *cache.Cache, client *ai.Client, rl *ai.RateLimiter, cloneDir string) *AIChatHandler {
	return &AIChatHandler{cache: c, client: client, rateLimiter: rl, cloneDir: cloneDir}
}

// HandleAIChat streams an AI chat response via SSE with per-IP rate limiting.
func (h *AIChatHandler) HandleAIChat(w http.ResponseWriter, r *http.Request) {
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

	// Check core analysis cache (409 if missing).
	coreData, err := h.cache.GetAnalysis(r.Context(), owner, repo, "")
	if err != nil {
		WriteError(w, http.StatusConflict, "analysis_pending",
			"Core analysis must complete before AI chat is available")
		return
	}
	var coreResult models.AnalysisResult
	if err := json.Unmarshal(coreData, &coreResult); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to read analysis data")
		return
	}
	sha := coreResult.Repository.LatestSHA

	// 429 — rate limit check.
	ip := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip = fwd
	}
	if !h.rateLimiter.Allow(ip) {
		w.Header().Set("Retry-After", "60")
		WriteError(w, http.StatusTooManyRequests, "rate_limited",
			"Chat rate limit exceeded (10 requests per minute). Try again later.")
		gritmetrics.AIRequestTotal.WithLabelValues("chat", "rate_limited").Inc()
		return
	}

	// Decode request body.
	var chatReq models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		WriteBadRequest(w, "Invalid request body: "+err.Error())
		return
	}

	// Validate.
	if err := ai.ValidateChatRequest(&chatReq); err != nil {
		status := http.StatusBadRequest
		code := "bad_request"
		msg := err.Error()
		if errors.Is(err, ai.ErrEmptyMessages) {
			msg = "Messages array must not be empty"
		} else if errors.Is(err, ai.ErrLastRoleNotUser) {
			msg = "Last message role must be 'user'"
		} else if errors.Is(err, ai.ErrEmptyContent) {
			msg = "Message content must not be empty or whitespace"
		}
		WriteError(w, status, code, msg)
		return
	}

	// Assemble context.
	cloneDir := fmt.Sprintf("%s/%s/%s", h.cloneDir, owner, repo)
	contextParts, err := ai.AssembleContext(r.Context(), h.cache, owner, repo, sha, cloneDir)
	if err != nil {
		slog.Error("ai: assemble context failed", "owner", owner, "repo", repo, "error", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to build AI context")
		return
	}

	// Stream SSE.
	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	err = ai.GenerateChat(r.Context(), h.client, contextParts, &chatReq, func(chunk string) {
		fmt.Fprint(w, ai.FormatSSEEvent(chunk))
		flusher.Flush()
	})
	if err != nil {
		slog.Error("ai: chat generation failed", "owner", owner, "repo", repo, "error", err)
		fmt.Fprint(w, ai.FormatSSEEvent("[ERROR] Chat generation failed"))
		flusher.Flush()
		gritmetrics.AIRequestTotal.WithLabelValues("chat", "error").Inc()
		return
	}

	// Send done event.
	fmt.Fprint(w, "event: done\ndata: {}\n\n")
	flusher.Flush()

	gritmetrics.AIRequestTotal.WithLabelValues("chat", "success").Inc()
	gritmetrics.AIRequestDuration.WithLabelValues("chat").Observe(time.Since(start).Seconds())
}
