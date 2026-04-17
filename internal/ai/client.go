package ai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"
)

// Sentinel errors returned by the AI client. These are the ONLY errors
// that handlers should ever surface to clients — raw provider errors must
// never be exposed (Constitution Principle V).
var (
	ErrNotConfigured = errors.New("ai: not configured")
	ErrRateLimited   = errors.New("ai: rate limited")
	ErrUnavailable   = errors.New("ai: unavailable")
)

const (
	PrimaryModel  = "gemini-2.5-flash"
	FallbackModel = "gemini-2.5-flash-lite"

	maxRetryAttempts = 3
	baseBackoff      = 1 * time.Second
	maxBackoff       = 30 * time.Second
)

// Client wraps the Google Generative AI SDK with retry logic and model fallback.
type Client struct {
	inner *genai.Client
}

// NewClient initializes the Gemini client once at startup. Returns ErrNotConfigured
// if apiKey is empty so the caller can surface a 503.
func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, ErrNotConfigured
	}
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("ai: init client: %w", err)
	}
	return &Client{inner: c}, nil
}

// GenerateStream calls the streaming API, retrying with backoff and falling
// back to FallbackModel on the final attempt. The returned iterator yields
// genai.GenerateContentResponse values. The caller is responsible for
// iterating the result.
func (c *Client) GenerateStream(ctx context.Context, contents []*genai.Content, cfg *genai.GenerateContentConfig) *streamResult {
	return &streamResult{client: c, contents: contents, cfg: cfg, ctx: ctx}
}

// streamResult holds the parameters for a lazy-evaluated streaming call.
type streamResult struct {
	client   *Client
	contents []*genai.Content
	cfg      *genai.GenerateContentConfig
	ctx      context.Context
}

// Iter returns an iterator over the stream, performing retry with backoff.
func (sr *streamResult) Iter() (iter func(yield func(*genai.GenerateContentResponse, error) bool), model string, err error) {
	var lastErr error
	for attempt := range maxRetryAttempts {
		model := PrimaryModel
		if attempt == maxRetryAttempts-1 {
			model = FallbackModel
		}

		stream := sr.client.inner.Models.GenerateContentStream(sr.ctx, model, sr.contents, sr.cfg)
		if stream != nil {
			return stream, model, nil
		}

		lastErr = fmt.Errorf("ai: nil stream on attempt %d", attempt+1)
		slog.Warn("ai stream nil, retrying", "attempt", attempt+1, "model", model, "error", lastErr)

		if attempt < maxRetryAttempts-1 {
			sleepWithBackoff(sr.ctx, attempt)
		}
	}
	return nil, FallbackModel, maskError(lastErr)
}

// Generate calls the non-streaming API with retry and model fallback.
func (c *Client) Generate(ctx context.Context, contents []*genai.Content, cfg *genai.GenerateContentConfig) (*genai.GenerateContentResponse, string, error) {
	var lastErr error
	for attempt := range maxRetryAttempts {
		model := PrimaryModel
		if attempt == maxRetryAttempts-1 {
			model = FallbackModel
		}

		resp, err := c.inner.Models.GenerateContent(ctx, model, contents, cfg)
		if err == nil {
			return resp, model, nil
		}
		lastErr = err
		slog.Warn("ai generate failed, retrying", "attempt", attempt+1, "model", model, "error", err)

		if isRetryable(err) && attempt < maxRetryAttempts-1 {
			sleepWithBackoff(ctx, attempt)
			continue
		}
		if !isRetryable(err) {
			break
		}
	}
	return nil, FallbackModel, maskError(lastErr)
}

// sleepWithBackoff sleeps for min(base * 2^attempt + jitter, max).
func sleepWithBackoff(ctx context.Context, attempt int) {
	delay := time.Duration(float64(baseBackoff) * math.Pow(2, float64(attempt)))
	jitter := time.Duration(rand.Int64N(int64(baseBackoff)))
	delay += jitter
	if delay > maxBackoff {
		delay = maxBackoff
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// isRetryable returns true for 429 rate limit and 5xx server errors.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	if strings.Contains(s, "429") || strings.Contains(s, "RESOURCE_EXHAUSTED") {
		return true
	}
	for code := http.StatusInternalServerError; code <= http.StatusNetworkAuthenticationRequired; code++ {
		if strings.Contains(s, fmt.Sprintf("%d", code)) {
			return true
		}
	}
	return false
}

// maskError translates raw provider errors into our sentinel errors.
func maskError(err error) error {
	if err == nil {
		return nil
	}
	s := err.Error()
	if strings.Contains(s, "429") || strings.Contains(s, "RESOURCE_EXHAUSTED") {
		return ErrRateLimited
	}
	return ErrUnavailable
}
