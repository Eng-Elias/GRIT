# Quickstart: AI Layer (Gemini)

## Prerequisites

- GRIT backend running with core analysis functional
- Redis and NATS running
- Google Gemini API key (free tier: https://aistudio.google.com/apikey)

## Configuration

Add to your `.env` file:

```
GEMINI_API_KEY=your_api_key_here
```

The backend reads this at startup. If not set, AI endpoints return 503.

## Usage

### 1. Ensure Core Analysis Exists

First, trigger a core analysis (AI features require cached results):

```bash
curl http://localhost:8080/api/golang/go
# Wait for job to complete (poll /api/golang/go/status)
```

### 2. AI Codebase Summary (Streaming)

```bash
curl -N http://localhost:8080/api/golang/go/ai/summary
```

Response streams as SSE:
```
data: {"chunk": "The Go programming language..."}
data: {"chunk": " is a statically typed..."}
data: {"done": true, "summary": { ... }}
```

Second request returns cached JSON instantly.

### 3. Ask This Repo (Chat)

```bash
curl -X POST http://localhost:8080/api/golang/go/ai/chat \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "What testing framework does this project use?"}
    ]
  }'
```

Response streams as SSE. Send full conversation history each time:

```bash
curl -X POST http://localhost:8080/api/golang/go/ai/chat \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "What testing framework does this project use?"},
      {"role": "model", "content": "The Go standard library testing package..."},
      {"role": "user", "content": "Are there any third-party test helpers?"}
    ]
  }'
```

Rate limit: 10 requests per minute per IP.

### 4. AI Health Score

```bash
curl http://localhost:8080/api/golang/go/ai/health
```

Returns structured JSON (non-streaming):
```json
{
  "overall_score": 72,
  "categories": {
    "readme_quality": { "score": 85, "notes": "..." },
    "contributing_guide": { "score": 40, "notes": "..." },
    "code_documentation": { "score": 70, "notes": "..." },
    "test_coverage_signals": { "score": 80, "notes": "..." },
    "project_hygiene": { "score": 85, "notes": "..." }
  },
  "top_improvements": ["...", "...", "..."]
}
```

## Error Handling

| Scenario | Response |
|----------|----------|
| No GEMINI_API_KEY | 503 `{"error": "ai_not_configured"}` |
| Core analysis not cached | 409 `{"error": "analysis_pending"}` |
| Gemini rate limited | 429 `{"error": "rate_limited", "retry_after": 60}` |
| Chat IP rate limited | 429 `{"error": "rate_limited", "retry_after": 60}` |
| Gemini unreachable | 503 `{"error": "ai_unavailable"}` |

## Verification

```bash
# Check AI endpoints are registered
curl -s http://localhost:8080/api/test-owner/test-repo/ai/summary | jq .error
# Expected: "analysis_pending" (if no core analysis) or "ai_not_configured" (if no key)

# Check metrics
curl -s http://localhost:8080/metrics | grep ai_
```
