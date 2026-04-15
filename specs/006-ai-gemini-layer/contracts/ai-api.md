# API Contract: AI Layer (Gemini)

## POST /api/:owner/:repo/ai/summary

Generate a streaming AI codebase summary.

### Request

No body required.

### Response — Streaming (200 OK)

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

data: {"chunk": "This project is a..."}

data: {"chunk": " repository analysis tool..."}

data: {"done": true, "summary": { ... full AISummary object ... }}

```

### Response — Cached (200 OK)

```
Content-Type: application/json
X-Cache: HIT

{
  "repository": { "owner": "...", "name": "...", "full_name": "...", "latest_sha": "..." },
  "description": "2-3 paragraphs...",
  "architecture": "Monolith with modular packages",
  "tech_stack": ["Go", "Redis", "NATS"],
  "red_flags": ["No CI configuration detected", "..."],
  "entry_points": ["Start with cmd/main.go", "..."],
  "generated_at": "2026-04-15T20:00:00Z",
  "model": "gemini-2.5-flash",
  "cached": true,
  "cached_at": "2026-04-15T20:00:00Z"
}
```

### Error Responses

| Status | Body | Condition |
|--------|------|-----------|
| 400 | `{ "error": "bad_request", "message": "Invalid repository identifier" }` | Invalid owner/repo |
| 409 | `{ "error": "analysis_pending", "message": "Core analysis must complete first" }` | No cached core analysis |
| 429 | `{ "error": "rate_limited", "retry_after": 60 }` | Gemini rate limit after retries |
| 503 | `{ "error": "ai_not_configured", "message": "AI features require API key" }` | No GEMINI_API_KEY |
| 503 | `{ "error": "ai_unavailable", "message": "AI service temporarily unavailable" }` | Gemini unreachable |

---

## POST /api/:owner/:repo/ai/chat

Send a chat message about the repository.

### Request

```json
{
  "messages": [
    { "role": "user", "content": "What does this project do?" },
    { "role": "model", "content": "This project is..." },
    { "role": "user", "content": "How is the code organized?" }
  ]
}
```

### Response — Streaming (200 OK)

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

data: {"chunk": "The code is organized..."}

data: {"chunk": " into several packages..."}

data: {"done": true}

```

### Error Responses

| Status | Body | Condition |
|--------|------|-----------|
| 400 | `{ "error": "bad_request", "message": "Messages array is required" }` | Empty or missing messages |
| 400 | `{ "error": "bad_request", "message": "Last message must be from user" }` | Last role != "user" |
| 409 | `{ "error": "analysis_pending", "message": "Core analysis must complete first" }` | No cached core analysis |
| 429 | `{ "error": "rate_limited", "retry_after": 60 }` | IP rate limit (10/min) or Gemini rate limit |
| 503 | `{ "error": "ai_not_configured", "message": "AI features require API key" }` | No GEMINI_API_KEY |
| 503 | `{ "error": "ai_unavailable", "message": "AI service temporarily unavailable" }` | Gemini unreachable |

---

## GET /api/:owner/:repo/ai/health

Get a structured AI health assessment of the repository.

### Request

No body. No query parameters.

### Response — Success (200 OK)

```json
{
  "repository": { "owner": "...", "name": "...", "full_name": "...", "latest_sha": "..." },
  "overall_score": 72,
  "categories": {
    "readme_quality":        { "score": 85, "notes": "Comprehensive README with..." },
    "contributing_guide":    { "score": 40, "notes": "No CONTRIBUTING.md found" },
    "code_documentation":    { "score": 70, "notes": "Moderate inline documentation..." },
    "test_coverage_signals": { "score": 80, "notes": "Test files found in most packages..." },
    "project_hygiene":       { "score": 85, "notes": "License present, .gitignore..." }
  },
  "top_improvements": [
    "Add a CONTRIBUTING.md guide for new contributors",
    "Add godoc comments to exported functions",
    "Set up a CI pipeline configuration"
  ],
  "generated_at": "2026-04-15T20:00:00Z",
  "model": "gemini-2.5-flash",
  "cached": false,
  "cached_at": null
}
```

### Response — Cached (200 OK)

Same structure with `"cached": true` and `X-Cache: HIT` header.

### Error Responses

| Status | Body | Condition |
|--------|------|-----------|
| 400 | `{ "error": "bad_request", "message": "Invalid repository identifier" }` | Invalid owner/repo |
| 409 | `{ "error": "analysis_pending", "message": "Core analysis must complete first" }` | No cached core analysis |
| 429 | `{ "error": "rate_limited", "retry_after": 60 }` | Gemini rate limit after retries |
| 503 | `{ "error": "ai_not_configured", "message": "AI features require API key" }` | No GEMINI_API_KEY |
| 503 | `{ "error": "ai_unavailable", "message": "AI service temporarily unavailable" }` | Gemini unreachable |

---

## SSE Protocol

All streaming endpoints follow this protocol:

- Content-Type: `text/event-stream`
- Each event: `data: <JSON>\n\n`
- Chunk event: `data: {"chunk": "<text>"}\n\n`
- Final event: `data: {"done": true, ...optional full result...}\n\n`
- Error during stream: `data: {"error": "<code>", "message": "<msg>"}\n\n`
- Client should close connection after receiving `{"done": true}` or `{"error": ...}`
