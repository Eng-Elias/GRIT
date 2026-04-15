# API Contract: Temporal Intelligence

## GET /api/:owner/:repo/temporal

Retrieve temporal analysis data for a repository including LOC over time, velocity metrics, and refactor detection.

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| period | string | No | "3y" | Time range for LOC data: "1y", "2y", or "3y" |
| sha | string | No | "" | Specific commit SHA (uses latest cached if empty) |

### Response: 200 OK (Cache HIT)

**Headers**: `Content-Type: application/json`, `X-Cache: HIT`

```json
{
  "repository": {
    "owner": "facebook",
    "name": "react",
    "full_name": "facebook/react",
    "latest_sha": "abc123..."
  },
  "period": "3y",
  "loc_over_time": [
    {
      "date": "2023-05-01T00:00:00Z",
      "total_loc": 245000,
      "by_language": {
        "JavaScript": 180000,
        "TypeScript": 30000,
        "Python": 12000,
        "Go": 8000,
        "CSS": 5000,
        "HTML": 4000,
        "Shell": 2500,
        "Markdown": 1500,
        "YAML": 1000,
        "JSON": 500,
        "Other": 500
      }
    }
  ],
  "velocity": {
    "weekly_activity": [
      {
        "week": "2025-W15",
        "week_start": "2025-04-07T00:00:00Z",
        "additions": 1234,
        "deletions": 567,
        "commits": 42
      }
    ],
    "author_activity": [
      {
        "email": "dev@example.com",
        "name": "Jane Dev",
        "total_additions": 15000,
        "total_deletions": 8000,
        "weeks": [
          {
            "week": "2025-W15",
            "week_start": "2025-04-07T00:00:00Z",
            "additions": 200,
            "deletions": 100,
            "commits": 5
          }
        ]
      }
    ],
    "commit_cadence": [
      {
        "window_start": "2026-03-17T00:00:00Z",
        "window_end": "2026-04-13T00:00:00Z",
        "commits_per_day": 3.5
      }
    ],
    "pr_merge_time": {
      "median_hours": 18.5,
      "p75_hours": 48.0,
      "p95_hours": 120.0,
      "sample_size": 100
    }
  },
  "refactor_periods": [
    {
      "start": "2025-09-01T00:00:00Z",
      "end": "2025-09-21T00:00:00Z",
      "net_loc_change": -4500,
      "weeks": 3
    }
  ],
  "total_months": 36,
  "total_weeks": 52,
  "analyzed_at": "2026-04-14T15:30:00Z",
  "cached": true,
  "cached_at": "2026-04-14T15:30:00Z"
}
```

### Response: 202 Accepted (Job In Progress)

**Headers**: `Content-Type: application/json`

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "poll_url": "/api/facebook/react/status"
}
```

### Response: 404 Not Found

```json
{
  "error": "not_found",
  "message": "Temporal analysis not yet available for facebook/react."
}
```

### Response: 400 Bad Request

```json
{
  "error": "bad_request",
  "message": "Invalid period parameter: must be 1y, 2y, or 3y."
}
```

---

## temporal_summary (embedded in GET /api/:owner/:repo)

When temporal data is available, the main analysis response includes a `temporal_summary` object.

### When complete

```json
{
  "temporal_summary": {
    "status": "complete",
    "current_loc": 245000,
    "loc_trend_6m_percent": 12.5,
    "avg_weekly_commits": 42.3,
    "temporal_url": "/api/facebook/react/temporal"
  }
}
```

### When pending

```json
{
  "temporal_summary": {
    "status": "pending",
    "current_loc": 0,
    "loc_trend_6m_percent": 0,
    "avg_weekly_commits": 0,
    "temporal_url": "/api/facebook/react/temporal"
  }
}
```

---

## Notes

- `pr_merge_time` is `null` when GitHub GraphQL is unavailable, rate-limited, or no merged PRs exist.
- `by_language` contains up to 10 named languages plus "Other" for the remainder. If there are 10 or fewer languages, "Other" is omitted.
- `weekly_activity` always contains up to 52 entries regardless of the `period` parameter. The period only affects `loc_over_time`.
- `refactor_periods` may be empty if no weeks meet the refactor criteria.
- `author_activity[].weeks` only includes weeks where the author had at least one commit.
