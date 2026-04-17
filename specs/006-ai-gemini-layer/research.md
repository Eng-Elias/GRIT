# Research: AI Layer (Gemini)

## R1: Gemini Go SDK Initialization

**Decision**: Use `google.golang.org/genai` with `genai.NewClient(ctx, &genai.ClientConfig{APIKey: key, Backend: genai.BackendGeminiAPI})`. Initialize once at startup in `main.go` and inject into handlers.

**Rationale**: The `genai` SDK is the official Google Go client. Single initialization avoids repeated auth handshakes and allows the client to manage connection pooling internally. `BackendGeminiAPI` targets the Gemini REST API directly (not Vertex AI).

**Alternatives considered**:
- Raw HTTP calls to Gemini REST API — rejected; SDK handles auth, serialization, streaming protocol
- Vertex AI backend — rejected; requires GCP project setup, not self-hostable friendly

## R2: Streaming via SSE

**Decision**: Use `client.Models.GenerateContentStream(ctx, model, contents, config)` which returns an iterator. Iterate in the handler goroutine, writing each chunk to `http.ResponseWriter` via `http.Flusher` with `text/event-stream` content type.

**Rationale**: SSE is natively supported by browsers and the `http.Flusher` interface in Go's stdlib. The genai SDK's streaming API returns chunks as they arrive from Gemini, enabling low-latency first-token delivery.

**Alternatives considered**:
- WebSocket — rejected; more complex, bidirectional not needed for AI responses
- Long polling — rejected; higher latency, worse UX for streaming content
- Buffered response — rejected; users must wait for full generation (10-30s)

**Implementation pattern**:
```
w.Header().Set("Content-Type", "text/event-stream")
w.Header().Set("Cache-Control", "no-cache")
w.Header().Set("Connection", "keep-alive")
flusher := w.(http.Flusher)
for result, err := range stream {
    // write "data: <chunk>\n\n"
    flusher.Flush()
}
// write "data: [DONE]\n\n"
```

## R3: Retry with Model Fallback

**Decision**: Implement `retryWithBackoff(ctx, fn, maxAttempts=3)` that catches HTTP 429 / server errors. On attempt 3, switch model from `gemini-2.5-flash` to `gemini-2.5-flash-lite`. Use exponential backoff with jitter: `min(baseDelay * 2^attempt + jitter, maxDelay)` where base=1s, max=30s.

**Rationale**: Gemini free tier is rate-limited. Flash-Lite has a higher free quota and lower latency, making it a natural fallback. Jitter prevents thundering herd on retry.

**Alternatives considered**:
- No fallback (just retry same model) — rejected; if Flash is rate-limited, retrying same model wastes attempts
- Immediate fallback to Flash-Lite on first failure — rejected; Flash-Lite has lower quality; prefer Flash when possible
- Queue-based retry — rejected; overkill for 3 attempts, adds latency

## R4: Rate Limiting (Chat)

**Decision**: Use `golang.org/x/time/rate.NewLimiter(rate.Every(6*time.Second), 10)` per IP, stored in a `sync.Map`. The limiter allows 10 events with a refill rate of ~10/minute.

**Rationale**: `x/time/rate` is the standard Go rate limiter. `sync.Map` is lock-free for concurrent reads (common case). Per-IP isolation prevents one user from exhausting the limit for others.

**Alternatives considered**:
- Redis-based rate limiting — rejected; adds latency for every chat request, constitution says Redis is only for caching
- Fixed-window counter — rejected; allows bursts at window boundaries
- Third-party middleware (e.g., chi rate limiter) — rejected; less control over per-IP granularity and custom 429 response format

**Cleanup**: A background goroutine periodically evicts stale entries from the `sync.Map` (e.g., every 5 minutes, remove IPs not seen in the last 10 minutes).

## R5: Context Construction

**Decision**: Pure function `BuildContext(metadata, fileTree, readme, manifests, complexFiles, dirSummary) []genai.Part` in `internal/ai/context.go`. Assembles sections with token budget enforcement using 4-chars-per-token approximation.

**Rationale**: A pure function is easily testable with no side effects. Returning `[]genai.Part` integrates directly with the genai SDK's content format.

**Token budget allocation** (800K total):
1. Metadata: ~500 tokens (always included)
2. README: up to 8,000 tokens (32,000 chars)
3. Manifests: up to 2,000 tokens each × 7 = 14,000 tokens max
4. Top 5 complex files: ~150 lines × 5 ≈ 15,000 tokens
5. Directory summary: ~2,000 tokens
6. File tree: remaining budget (up to ~760,000 tokens)

**Truncation priority** (lowest priority truncated first):
1. File tree paths (truncated first if oversized)
2. Complex file content
3. Manifest content
4. README content (truncated last)

## R6: Health Score JSON Parsing

**Decision**: Use `GenerationConfig{ResponseMIMEType: "application/json"}` to instruct Gemini to return structured JSON. Parse with `json.Unmarshal` into `models.HealthScore`. On parse failure, retry once with a stricter prompt that includes the exact JSON schema.

**Rationale**: Gemini supports constrained JSON output via ResponseMIMEType. This dramatically reduces parse failures compared to free-form text extraction.

**Alternatives considered**:
- Regex extraction from markdown — rejected; fragile, error-prone
- Function calling — rejected; adds complexity, JSON mode is simpler for this use case
- Always send schema in prompt — rejected; wastes tokens on happy path; save strict prompt for retry

## R7: Error Masking

**Decision**: All Gemini errors are caught at the client layer and mapped to one of:
- `{ "error": "rate_limited", "retry_after": 60 }` — after exhausting retries on 429
- `{ "error": "ai_unavailable" }` — network/server errors after retries
- `{ "error": "analysis_pending" }` — core analysis not cached
- `{ "error": "ai_not_configured" }` — no GEMINI_API_KEY

Raw Gemini error messages, stack traces, and internal details are never included in client responses.

**Rationale**: Constitution Principle V explicitly requires this. Exposing raw errors leaks implementation details and confuses users.

## R8: Cache Key Patterns

**Decision**: Follow existing `{owner}/{repo}:{sha}:{pillar}` pattern:
- AI Summary: `{owner}/{repo}:{sha}:ai_summary` — TTL 1h
- AI Health: `{owner}/{repo}:{sha}:ai_health` — TTL 6h
- Chat: not cached (per spec FR-015)

**Rationale**: Matches Constitution Principle IV key format and TTL requirements.
