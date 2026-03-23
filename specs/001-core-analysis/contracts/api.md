# API Contracts: Core Analysis Engine

**Branch**: `001-core-analysis`
**Date**: 2026-03-23
**Base URL**: `/api`

## Common Headers

### Request Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | No | `Bearer <github_pat>` for higher rate limits |
| `Accept` | No | `application/json` (default) |

### Response Headers

| Header | Present | Description |
|--------|---------|-------------|
| `Content-Type` | Always | `application/json` |
| `X-Cache` | Analysis endpoints | `HIT` or `MISS` |
| `Retry-After` | On 429 | Seconds until rate limit resets |

## Error Envelope

All error responses use this format:

```json
{
  "error": "<error_code>",
  "message": "<human-readable description>"
}
```

### Error Codes

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `bad_request` | Malformed `owner/repo` identifier |
| 403 | `private_repo` | Private repo accessed without token |
| 404 | `not_found` | Repository does not exist on GitHub |
| 429 | `rate_limited` | GitHub API rate limit exceeded |
| 503 | `service_unavailable` | NATS or Redis unavailable |
| 504 | `analysis_timeout` | Analysis exceeded 5-minute timeout |

---

## 1. GET /api/:owner/:repo

**Description**: Returns the full core analysis result for a repository.

### Request

```
GET /api/facebook/react
Authorization: Bearer ghp_xxxx  (optional)
```

### Response — Cache HIT (200)

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Cache: HIT
```

```json
{
  "repository": {
    "owner": "facebook",
    "name": "react",
    "full_name": "facebook/react",
    "default_branch": "main",
    "latest_sha": "abc123..."
  },
  "metadata": {
    "name": "react",
    "description": "The library for web and native user interfaces.",
    "homepage_url": "https://react.dev",
    "stars": 230000,
    "forks": 47000,
    "watchers": 6700,
    "open_issues": 900,
    "size_kb": 350000,
    "default_branch": "main",
    "primary_language": "JavaScript",
    "license_name": "MIT License",
    "license_spdx": "MIT",
    "topics": ["react", "javascript", "frontend", "ui"],
    "created_at": "2013-05-24T16:15:54Z",
    "updated_at": "2026-03-23T10:00:00Z",
    "pushed_at": "2026-03-23T09:45:00Z",
    "has_wiki": true,
    "has_projects": true,
    "has_discussions": true,
    "community_health": 100
  },
  "commit_activity": {
    "weekly_counts": [12, 15, 8, 20, 18, "...(52 entries)"],
    "total_commits": 18500
  },
  "files": [
    {
      "path": "packages/react/src/React.js",
      "language": "JavaScript",
      "total_lines": 150,
      "code_lines": 120,
      "comment_lines": 20,
      "blank_lines": 10,
      "byte_size": 4500
    }
  ],
  "languages": [
    {
      "language": "JavaScript",
      "file_count": 2500,
      "total_lines": 450000,
      "code_lines": 380000,
      "comment_lines": 40000,
      "blank_lines": 30000,
      "percentage": 65.5
    },
    {
      "language": "TypeScript",
      "file_count": 800,
      "total_lines": 150000,
      "code_lines": 130000,
      "comment_lines": 10000,
      "blank_lines": 10000,
      "percentage": 21.8
    }
  ],
  "total_files": 5200,
  "total_lines": 687000,
  "total_code_lines": 580000,
  "total_comment_lines": 60000,
  "total_blank_lines": 47000,
  "analyzed_at": "2026-03-23T08:00:00Z",
  "cached": true,
  "cached_at": "2026-03-23T08:00:00Z"
}
```

### Response — Cache MISS (202)

```
HTTP/1.1 202 Accepted
Content-Type: application/json
X-Cache: MISS
```

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "queued",
  "poll_url": "/api/facebook/react/status"
}
```

### Error Responses

- `400` — Invalid `owner/repo` format
- `403` — Private repo, no token provided
- `404` — Repository not found on GitHub
- `429` — GitHub rate limit exceeded (includes `Retry-After` header)
- `503` — NATS or Redis unavailable

---

## 2. GET /api/:owner/:repo/status

**Description**: Poll job progress for an in-progress analysis.

### Request

```
GET /api/facebook/react/status
```

### Response — Job Running (200)

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "progress": {
    "clone": "completed",
    "file_walk": "running",
    "metadata_fetch": "completed",
    "commit_activity_fetch": "running"
  },
  "created_at": "2026-03-23T08:00:00Z"
}
```

### Response — Job Completed (200)

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "progress": {
    "clone": "completed",
    "file_walk": "completed",
    "metadata_fetch": "completed",
    "commit_activity_fetch": "completed"
  },
  "created_at": "2026-03-23T08:00:00Z",
  "completed_at": "2026-03-23T08:01:30Z",
  "result_url": "/api/facebook/react"
}
```

### Response — Job Failed (200)

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "failed",
  "progress": {
    "clone": "completed",
    "file_walk": "failed",
    "metadata_fetch": "completed",
    "commit_activity_fetch": "completed"
  },
  "created_at": "2026-03-23T08:00:00Z",
  "completed_at": "2026-03-23T08:05:00Z",
  "error": "analysis_timeout"
}
```

### Response — No Active Job (200)

If no job exists and no cached result, returns:

```json
{
  "status": "not_found",
  "message": "No analysis in progress. Request GET /api/:owner/:repo to start one."
}
```

---

## 3. GET /api/:owner/:repo/badge

**Description**: Returns a shields.io-compatible JSON badge.

### Request

```
GET /api/facebook/react/badge
```

### Response — Cached Result Available (200)

```json
{
  "schemaVersion": 1,
  "label": "lines of code",
  "message": "687k",
  "color": "blue"
}
```

### Response — Analysis In Progress (200)

```json
{
  "schemaVersion": 1,
  "label": "lines of code",
  "message": "analyzing...",
  "color": "lightgrey"
}
```

### Badge Color Rules

| Total Lines | Color |
|-------------|-------|
| < 1,000 | `brightgreen` |
| < 10,000 | `green` |
| < 100,000 | `yellowgreen` |
| < 500,000 | `yellow` |
| < 1,000,000 | `orange` |
| ≥ 1,000,000 | `blue` |

### Message Formatting

| Value | Formatted |
|-------|-----------|
| 500 | `500` |
| 1,500 | `1.5k` |
| 25,000 | `25k` |
| 1,500,000 | `1.5M` |

---

## 4. DELETE /api/:owner/:repo/cache

**Description**: Manually invalidate cached analysis for a repository.

### Request

```
DELETE /api/facebook/react/cache
```

### Response — Cache Cleared (204)

```
HTTP/1.1 204 No Content
```

### Response — No Cache Entry (204)

Returns 204 even if no cache entry existed (idempotent).

---

## 5. GET /metrics

**Description**: Prometheus-compatible metrics endpoint.

### Request

```
GET /metrics
```

### Response (200)

Standard Prometheus text format with the following metrics:

```
# HELP grit_jobs_completed_total Total analysis jobs completed
# TYPE grit_jobs_completed_total counter
grit_jobs_completed_total 42

# HELP grit_jobs_failed_total Total analysis jobs failed
# TYPE grit_jobs_failed_total counter
grit_jobs_failed_total 3

# HELP grit_cache_hit_total Total cache hits
# TYPE grit_cache_hit_total counter
grit_cache_hit_total 150

# HELP grit_cache_miss_total Total cache misses
# TYPE grit_cache_miss_total counter
grit_cache_miss_total 45

# HELP grit_clone_duration_seconds Time to clone repository
# TYPE grit_clone_duration_seconds histogram
grit_clone_duration_seconds_bucket{le="1"} 10
...

# HELP grit_analysis_duration_seconds End-to-end analysis duration
# TYPE grit_analysis_duration_seconds histogram
grit_analysis_duration_seconds_bucket{le="30"} 20
...

# HELP grit_github_api_requests_total GitHub API requests by endpoint
# TYPE grit_github_api_requests_total counter
grit_github_api_requests_total{endpoint="repos"} 90
grit_github_api_requests_total{endpoint="stats"} 45
grit_github_api_requests_total{endpoint="community"} 45
```
