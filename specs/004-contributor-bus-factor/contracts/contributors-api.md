# API Contract: Contributor Endpoints

**Date**: 2026-04-04  
**Feature**: [spec.md](../spec.md) | [plan.md](../plan.md)

## GET /api/:owner/:repo/contributors

Full contributor data with bus factor analysis.

### Request

```
GET /api/{owner}/{repo}/contributors
```

**Path Parameters**:
- `owner`: Repository owner (alphanumeric + `.` `-` `_`)
- `repo`: Repository name (alphanumeric + `.` `-` `_`)

**Query Parameters**:
- `sha` (optional): Specific commit SHA. If omitted, latest cached result is used.

### Response: 200 OK (Cached)

**Headers**:
- `Content-Type: application/json`
- `X-Cache: HIT`

```json
{
  "repository": {
    "owner": "octocat",
    "name": "hello-world",
    "full_name": "octocat/hello-world",
    "latest_sha": "abc123..."
  },
  "authors": [
    {
      "email": "alice@example.com",
      "name": "Alice Smith",
      "total_lines_owned": 4200,
      "ownership_percent": 42.0,
      "files_touched": 15,
      "first_commit_date": "2024-01-15T10:30:00Z",
      "last_commit_date": "2026-03-28T14:00:00Z",
      "primary_languages": ["Go", "Python", "JavaScript"],
      "is_active": true
    }
  ],
  "bus_factor": 2,
  "key_people": [
    {
      "email": "alice@example.com",
      "name": "Alice Smith",
      "total_lines_owned": 4200,
      "ownership_percent": 42.0,
      "files_touched": 15,
      "first_commit_date": "2024-01-15T10:30:00Z",
      "last_commit_date": "2026-03-28T14:00:00Z",
      "primary_languages": ["Go", "Python", "JavaScript"],
      "is_active": true
    },
    {
      "email": "bob@example.com",
      "name": "Bob Jones",
      "total_lines_owned": 3900,
      "ownership_percent": 39.0,
      "files_touched": 12,
      "first_commit_date": "2024-03-01T09:00:00Z",
      "last_commit_date": "2026-04-01T11:30:00Z",
      "primary_languages": ["Go", "TypeScript"],
      "is_active": true
    }
  ],
  "total_lines_analyzed": 10000,
  "total_files_analyzed": 42,
  "partial": false,
  "analyzed_at": "2026-04-04T12:00:00Z",
  "cached": true,
  "cached_at": "2026-04-04T12:00:00Z"
}
```

### Response: 202 Accepted (Job in progress)

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "poll_url": "/api/octocat/hello-world/status"
}
```

### Response: 404 Not Found (No analysis)

```json
{
  "error": "not_found",
  "message": "Contributor analysis not yet available for octocat/hello-world."
}
```

### Response: 400 Bad Request

```json
{
  "error": "bad_request",
  "message": "Invalid repository identifier: ..."
}
```

---

## GET /api/:owner/:repo/contributors/files

Per-file contributor breakdown with top 3 authors per file.

### Request

```
GET /api/{owner}/{repo}/contributors/files
```

**Path Parameters**: Same as above.  
**Query Parameters**: Same as above.

### Response: 200 OK (Cached)

**Headers**:
- `Content-Type: application/json`
- `X-Cache: HIT`

```json
{
  "repository": {
    "owner": "octocat",
    "name": "hello-world",
    "full_name": "octocat/hello-world",
    "latest_sha": "abc123..."
  },
  "files": [
    {
      "path": "cmd/main.go",
      "total_lines": 150,
      "top_authors": [
        {
          "name": "Alice Smith",
          "email": "alice@example.com",
          "lines_owned": 100,
          "ownership_percent": 66.7
        },
        {
          "name": "Bob Jones",
          "email": "bob@example.com",
          "lines_owned": 40,
          "ownership_percent": 26.7
        },
        {
          "name": "Carol White",
          "email": "carol@example.com",
          "lines_owned": 10,
          "ownership_percent": 6.7
        }
      ]
    }
  ],
  "total_files": 42,
  "partial": false,
  "analyzed_at": "2026-04-04T12:00:00Z",
  "cached": true,
  "cached_at": "2026-04-04T12:00:00Z"
}
```

### Response: 202/404/400

Same structure as `/contributors` endpoint above.

---

## Embedded: contributor_summary in GET /api/:owner/:repo

Added to the existing main analysis response when contributor data is available.

### When contributor data IS available

```json
{
  "...existing fields...": "...",
  "contributor_summary": {
    "status": "complete",
    "bus_factor": 2,
    "top_contributors": [
      {
        "name": "Alice Smith",
        "email": "alice@example.com",
        "lines_owned": 4200
      },
      {
        "name": "Bob Jones",
        "email": "bob@example.com",
        "lines_owned": 3900
      }
    ],
    "total_authors": 5,
    "active_authors": 3,
    "contributors_url": "/api/octocat/hello-world/contributors"
  }
}
```

### When contributor data is NOT yet available

```json
{
  "...existing fields...": "...",
  "contributor_summary": {
    "status": "pending",
    "bus_factor": 0,
    "top_contributors": [],
    "total_authors": 0,
    "active_authors": 0,
    "contributors_url": "/api/octocat/hello-world/contributors"
  }
}
```
