# Quickstart: AST-Based Code Complexity Analysis

**Branch**: `002-complexity-analysis`
**Date**: 2026-03-30

## Prerequisites

- Go 1.22+ with cgo enabled (for tree-sitter C bindings)
- Docker and Docker Compose
- A C compiler (gcc) — included in most Go Docker images
- Feature 001 (core analysis) must be complete and merged

## 1. Install Dependencies

```bash
go get github.com/smacker/go-tree-sitter@latest
go get github.com/smacker/go-tree-sitter/golang@latest
go get github.com/smacker/go-tree-sitter/typescript@latest
go get github.com/smacker/go-tree-sitter/javascript@latest
go get github.com/smacker/go-tree-sitter/python@latest
go get github.com/smacker/go-tree-sitter/rust@latest
go get github.com/smacker/go-tree-sitter/java@latest
go get github.com/smacker/go-tree-sitter/c@latest
go get github.com/smacker/go-tree-sitter/cpp@latest
go get github.com/smacker/go-tree-sitter/ruby@latest
```

## 2. Start Infrastructure

```bash
docker compose up -d redis nats
```

## 3. Run the Backend

```bash
CGO_ENABLED=1 go run ./cmd/grit/
```

The server starts on `http://localhost:8080`.

## 4. Trigger Analysis

### Request core analysis (which auto-triggers complexity):

```bash
curl -s http://localhost:8080/api/sindresorhus/is
```

Wait for core analysis to complete (poll status), then complexity analysis runs automatically.

### Get complexity data:

```bash
curl -s http://localhost:8080/api/sindresorhus/is/complexity
```

Expected response (once complete):

```json
{
  "repository": { "owner": "sindresorhus", "name": "is", "..." : "..." },
  "files": [ { "path": "source/index.ts", "cyclomatic": 245, "..." : "..." } ],
  "hot_files": [ { "path": "source/index.ts", "complexity_density": 0.168 } ],
  "total_files_analyzed": 4,
  "total_function_count": 72,
  "mean_complexity": 65.5,
  "median_complexity": 12.0,
  "p90_complexity": 245.0,
  "distribution": { "low": 2, "medium": 1, "high": 0, "critical": 1 },
  "analyzed_at": "2026-03-30T12:00:00Z",
  "cached": false
}
```

### Check complexity summary in main endpoint:

```bash
curl -s http://localhost:8080/api/sindresorhus/is | jq '.complexity_summary'
```

Expected:

```json
{
  "status": "complete",
  "mean_complexity": 65.5,
  "total_function_count": 72,
  "hot_file_count": 4,
  "complexity_url": "/api/sindresorhus/is/complexity"
}
```

### Bust the cache (now also clears complexity):

```bash
curl -X DELETE http://localhost:8080/api/sindresorhus/is/cache
```

## 5. Run Tests

```bash
CGO_ENABLED=1 go test ./internal/analysis/complexity/... -v
CGO_ENABLED=1 go test ./... -v
```

Integration tests require Redis and NATS:

```bash
docker compose up -d redis nats
CGO_ENABLED=1 go test ./internal/... -v
```

## Verification Checklist

- [ ] `go build ./...` succeeds with cgo enabled (tree-sitter compiles)
- [ ] Core analysis of `sindresorhus/is` completes successfully
- [ ] Complexity analysis auto-triggers after core completion (check server logs)
- [ ] `GET /api/sindresorhus/is/complexity` returns 200 with per-file metrics
- [ ] Response includes `hot_files` ranked by `complexity_density`
- [ ] Response includes `mean_complexity`, `median_complexity`, `p90_complexity`
- [ ] Response includes `distribution` with low/medium/high/critical counts
- [ ] `GET /api/sindresorhus/is` includes `complexity_summary` field
- [ ] `DELETE /api/sindresorhus/is/cache` clears both core and complexity cache
- [ ] Second `GET /api/sindresorhus/is/complexity` returns `cached: true` with `X-Cache: HIT`
- [ ] Unsupported language files are excluded from complexity results
- [ ] `GET /metrics` includes `grit_complexity_analysis_duration_seconds`
