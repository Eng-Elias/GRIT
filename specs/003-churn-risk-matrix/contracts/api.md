# API Contract: Churn Matrix Endpoint

**Feature**: 003-churn-risk-matrix
**Date**: 2026-04-01

## GET /api/:owner/:repo/churn-matrix

Retrieve churn analysis, risk matrix, and dead code estimation for a repository.

### Request

```
GET /api/{owner}/{repo}/churn-matrix
```

**Path Parameters**:

| Parameter | Type   | Validation            | Description        |
|-----------|--------|-----------------------|--------------------|
| owner     | string | `^[a-zA-Z0-9_.-]+$`  | Repository owner   |
| repo      | string | `^[a-zA-Z0-9_.-]+$`  | Repository name    |

### Responses

#### 200 OK — Cached churn data available

**Headers**:
- `Content-Type: application/json`
- `X-Cache: HIT`

**Body**:

```json
{
  "repository": {
    "owner": "facebook",
    "name": "react",
    "full_name": "facebook/react"
  },
  "churn": [
    { "path": "src/ReactFiber.js", "churn": 342, "last_modified": "2026-03-28T14:22:00Z" },
    { "path": "src/ReactDOM.js", "churn": 289, "last_modified": "2026-03-30T09:11:00Z" }
  ],
  "risk_matrix": [
    {
      "path": "src/ReactFiber.js",
      "churn": 342,
      "complexity_cyclomatic": 47,
      "language": "JavaScript",
      "loc": 1200,
      "risk_level": "critical"
    },
    {
      "path": "src/ReactDOM.js",
      "churn": 289,
      "complexity_cyclomatic": 12,
      "language": "JavaScript",
      "loc": 450,
      "risk_level": "medium"
    }
  ],
  "risk_zone": [
    {
      "path": "src/ReactFiber.js",
      "churn": 342,
      "complexity_cyclomatic": 47,
      "language": "JavaScript",
      "loc": 1200,
      "risk_level": "critical"
    }
  ],
  "thresholds": {
    "churn_p50": 15.0,
    "churn_p75": 45.0,
    "churn_p90": 120.0,
    "complexity_p50": 5.0,
    "complexity_p75": 12.0,
    "complexity_p90": 25.0
  },
  "stale_files": [
    {
      "path": "src/legacy/OldModule.js",
      "last_modified": "2025-06-15T10:00:00Z",
      "months_inactive": 9
    }
  ],
  "total_commits": 4200,
  "commit_window_start": "2024-04-01T00:00:00Z",
  "commit_window_end": "2026-04-01T00:00:00Z",
  "total_files_churned": 856,
  "critical_count": 12,
  "stale_count": 3,
  "analyzed_at": "2026-04-01T12:00:00Z",
  "cached": true,
  "cached_at": "2026-04-01T12:00:05Z"
}
```

#### 202 Accepted — Churn job in progress

**Body**:

```json
{
  "status": "running",
  "job_id": "abc123-def456",
  "message": "Churn analysis is in progress",
  "poll_url": "/api/v1/jobs/abc123-def456"
}
```

#### 400 Bad Request — Invalid parameters

```json
{
  "error": "bad_request",
  "message": "Invalid repository identifier: bad owner/repo"
}
```

#### 404 Not Found — No analysis exists

```json
{
  "error": "not_found",
  "message": "No churn analysis found for owner/repo"
}
```

---

## Churn Summary (embedded in main analysis response)

When `GET /api/:owner/:repo` returns a cached core result, the response includes a `churn_summary` field:

### When churn data is available:

```json
{
  "churn_summary": {
    "status": "complete",
    "total_files": 856,
    "critical_count": 12,
    "stale_count": 3,
    "churn_matrix_url": "/api/facebook/react/churn-matrix"
  }
}
```

### When churn analysis is pending:

```json
{
  "churn_summary": {
    "status": "pending",
    "total_files": 0,
    "critical_count": 0,
    "stale_count": 0,
    "churn_matrix_url": "/api/facebook/react/churn-matrix"
  }
}
```

---

## Cache Deletion Side-Effect

`DELETE /api/:owner/:repo/cache` now clears churn cache entries in addition to core and complexity cache entries.
