# API Client Contract: GRIT Frontend

The frontend consumes the Go backend API. All endpoints are relative to the same origin (no CORS). During development, Vite proxies `/api/*` to `http://localhost:8080`.

## Base Configuration

- **Base URL**: `/api` (same origin, proxied in dev)
- **Content-Type**: `application/json` for all JSON endpoints
- **Error Format**: `{ "error": "<code>", "message": "<human-readable>" }`
- **Cache Header**: `X-Cache: HIT | MISS` on all analysis responses

## Endpoints

### Core Analysis

| Method | Path | Response | Notes |
|--------|------|----------|-------|
| GET | `/api/{owner}/{repo}` | `AnalysisResult` (200) or `{ job_id, status, poll_url }` (202) | Returns cached data or triggers async job |
| GET | `/api/{owner}/{repo}/status` | `AnalysisJob` | Polled at 3s interval until `completed` or `failed` |
| DELETE | `/api/{owner}/{repo}/cache` | 204 No Content | Busts all caches for the repo |

### Pillar Endpoints

| Method | Path | Response | Notes |
|--------|------|----------|-------|
| GET | `/api/{owner}/{repo}/complexity` | `ComplexityResult` | 409 if core analysis not yet complete |
| GET | `/api/{owner}/{repo}/churn-matrix` | `ChurnMatrixResult` | 409 if core analysis not yet complete |
| GET | `/api/{owner}/{repo}/contributors` | `ContributorResult` | 409 if core analysis not yet complete |
| GET | `/api/{owner}/{repo}/contributors/files` | `FileContributors[]` | Per-file ownership data |
| GET | `/api/{owner}/{repo}/temporal` | `TemporalResult` | Query param `?period=3y` (default) |

### AI Endpoints

| Method | Path | Response | Notes |
|--------|------|----------|-------|
| POST | `/api/{owner}/{repo}/ai/summary` | SSE stream or cached `AISummary` JSON | 503 if no API key, 409 if no core analysis |
| POST | `/api/{owner}/{repo}/ai/chat` | SSE stream | 429 if rate limited, body: `ChatRequest` |
| GET | `/api/{owner}/{repo}/ai/health` | `HealthScore` JSON | 503 if no API key, 409 if no core analysis |

### Badge

| Method | Path | Response | Notes |
|--------|------|----------|-------|
| GET | `/api/{owner}/{repo}/badge` | SVG redirect | shields.io badge |

## SSE Event Format

AI streaming endpoints use Server-Sent Events:

```
data: <text chunk>\n\n
data: <text chunk>\n\n
event: done
data: {}\n\n
```

- **`data:` events** — text chunks to append to display
- **`event: done`** — signals stream end
- **`data: [ERROR] ...`** — inline error during streaming

## Error Codes

| HTTP Status | Error Code | UI Action |
|-------------|------------|-----------|
| 400 | `bad_request` | Show validation error message |
| 403 | `private_repo` | Show "Private repository" banner |
| 404 | `not_found` | Show "Repository not found" page |
| 409 | `analysis_pending` | Show "Analysis in progress" banner |
| 429 | `rate_limited` | Show "Rate limited" with retry-after countdown |
| 500 | `internal_error` | Show "Something went wrong" generic error |
| 503 | `ai_unavailable` | Show "AI features not available" notice |

## Polling Behavior

The `useStatus` hook uses TanStack Query with:
- `refetchInterval: 3000` — poll every 3 seconds
- Auto-disable polling when `status === 'completed' || status === 'failed'`
- On `completed`: invalidate analysis query to trigger data refetch
- On `failed`: show error banner with `error` message
