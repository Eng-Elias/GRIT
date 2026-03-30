# API Contract: Complexity Analysis

**Feature**: 002-complexity-analysis  
**Date**: 2026-03-30  
**Base URL**: `http://localhost:8080`

## Endpoints

### GET /api/:owner/:repo/complexity

Returns full complexity analysis for a repository.

**Path Parameters**:

| Parameter | Type   | Constraints                       |
|-----------|--------|-----------------------------------|
| `owner`   | string | Must match `^[a-zA-Z0-9._-]+$`   |
| `repo`    | string | Must match `^[a-zA-Z0-9._-]+$`   |

**Headers**:

| Header          | Required | Description                          |
|-----------------|----------|--------------------------------------|
| Authorization   | No       | `Bearer <token>` for GitHub API auth |

**Response Headers**:

| Header   | Values          | Description                   |
|----------|-----------------|-------------------------------|
| X-Cache  | `HIT` or `MISS` | Whether result was from cache |

---

#### 200 OK — Cached complexity result available

```json
{
  "repository": {
    "owner": "sindresorhus",
    "name": "is",
    "full_name": "sindresorhus/is",
    "default_branch": "main",
    "latest_sha": "eff8e6b318d098317d5673b9d391f8e72cb15363",
    "size_kb": 1192
  },
  "files": [
    {
      "path": "source/index.ts",
      "language": "TypeScript",
      "cyclomatic": 245,
      "cognitive": 312,
      "function_count": 58,
      "avg_function_complexity": 4.22,
      "max_function_complexity": 18,
      "loc": 1459,
      "complexity_density": 0.168,
      "functions": [
        {
          "name": "isPlainObject",
          "start_line": 42,
          "end_line": 68,
          "cyclomatic": 18,
          "cognitive": 24
        },
        {
          "name": "isArray",
          "start_line": 70,
          "end_line": 75,
          "cyclomatic": 1,
          "cognitive": 0
        }
      ]
    }
  ],
  "hot_files": [
    {
      "path": "source/index.ts",
      "language": "TypeScript",
      "cyclomatic": 245,
      "cognitive": 312,
      "function_count": 58,
      "avg_function_complexity": 4.22,
      "max_function_complexity": 18,
      "loc": 1459,
      "complexity_density": 0.168,
      "functions": []
    }
  ],
  "total_files_analyzed": 4,
  "total_function_count": 72,
  "mean_complexity": 65.5,
  "median_complexity": 12.0,
  "p90_complexity": 245.0,
  "distribution": {
    "low": 2,
    "medium": 1,
    "high": 0,
    "critical": 1
  },
  "analyzed_at": "2026-03-30T12:00:00Z",
  "cached": true,
  "cached_at": "2026-03-30T12:00:00Z"
}
```

**Notes**:
- `hot_files` entries omit the `functions` array (empty) to reduce payload size
- `hot_files` is sorted by `complexity_density` descending, max 20 entries
- Files with 0 functions are excluded from `mean_complexity`, `median_complexity`, `p90_complexity`

---

#### 202 Accepted — Complexity analysis in progress

Returned when complexity analysis has not yet completed (either pending or running).

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "poll_url": "/api/sindresorhus/is/status"
}
```

---

#### 400 Bad Request — Invalid owner/repo

```json
{
  "error": "bad_request",
  "message": "Invalid repository identifier: bad owner/repo"
}
```

---

#### 404 Not Found — Core analysis not yet available

Returned when complexity is requested but core analysis hasn't completed.

```json
{
  "error": "not_found",
  "message": "Core analysis must complete before complexity data is available."
}
```

---

#### 404 Not Found — Repository does not exist

```json
{
  "error": "not_found",
  "message": "Repository nonexistent/repo does not exist."
}
```

---

### GET /api/:owner/:repo (Modified)

The existing core analysis endpoint is extended with an optional `complexity_summary` field.

#### Complexity summary when available

```json
{
  "repository": { "..." : "..." },
  "metadata": { "..." : "..." },
  "files": [],
  "languages": [],
  "total_files": 14,
  "total_lines": 5261,
  "analyzed_at": "2026-03-30T12:00:00Z",
  "cached": true,
  "cached_at": "2026-03-30T12:00:00Z",
  "complexity_summary": {
    "status": "complete",
    "mean_complexity": 65.5,
    "total_function_count": 72,
    "hot_file_count": 4,
    "complexity_url": "/api/sindresorhus/is/complexity"
  }
}
```

#### Complexity summary when pending

```json
{
  "...": "core analysis fields...",
  "complexity_summary": {
    "status": "pending",
    "mean_complexity": 0,
    "total_function_count": 0,
    "hot_file_count": 0,
    "complexity_url": "/api/sindresorhus/is/complexity"
  }
}
```

---

### DELETE /api/:owner/:repo/cache (Modified)

The existing cache deletion endpoint is extended to also delete complexity cache entries.

**Behavior change**: In addition to deleting `{owner}/{repo}:*:core` keys, also deletes `{owner}/{repo}:*:complexity` keys.

**Response**: 204 No Content (unchanged)

---

## Error Envelope

All error responses use the existing consistent JSON envelope:

```json
{
  "error": "<error_code>",
  "message": "<human-readable description>"
}
```

| HTTP Status | Error Code          | When                                    |
|-------------|---------------------|-----------------------------------------|
| 400         | `bad_request`       | Invalid owner/repo format               |
| 404         | `not_found`         | Repo not found or core analysis missing |
| 429         | `rate_limited`      | GitHub API rate limit hit               |
| 503         | `service_unavailable`| Redis or NATS down                     |
