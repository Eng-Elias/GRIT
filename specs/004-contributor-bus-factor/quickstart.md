# Quickstart: Contributor Attribution & Bus Factor Analysis

**Date**: 2026-04-04  
**Feature**: [spec.md](spec.md) | [plan.md](plan.md)

## Prerequisites

- Go 1.22+
- Docker Compose (Redis 7 + NATS with JetStream)
- Existing GRIT stack running (`docker compose up`)

## Development Workflow

### 1. Run the stack

```bash
docker compose up -d redis nats
```

### 2. Run all tests (existing + new)

```bash
# Unit tests for the blame analysis pillar
go test ./internal/analysis/blame/... -v

# Unit tests for new models
go test ./internal/models/... -v

# Handler tests
go test ./internal/handler/... -v

# Cache tests
go test ./internal/cache/... -v

# Job tests
go test ./internal/job/... -v

# All tests
go test ./... -v
```

### 3. Manual endpoint testing

```bash
# Trigger core analysis (blame auto-enqueues after core completes)
curl -X GET http://localhost:8080/api/octocat/hello-world

# Poll until complete
curl http://localhost:8080/api/octocat/hello-world/status

# Get contributor data
curl http://localhost:8080/api/octocat/hello-world/contributors

# Get per-file breakdown
curl http://localhost:8080/api/octocat/hello-world/contributors/files

# Verify contributor_summary in main response
curl http://localhost:8080/api/octocat/hello-world | jq .contributor_summary
```

## Key Implementation Notes

### Pipeline Flow

```
Core Analysis → [parallel]
  ├── Complexity Job (grit.jobs.complexity)
  ├── Churn Job (grit.jobs.churn)  — triggered by complexity worker
  └── Blame Job (grit.jobs.blame)  — independent, triggered by core worker
```

### Goroutine Pool Pattern

```go
pool := min(4, runtime.NumCPU())
fileCh := make(chan string, len(files))
var wg sync.WaitGroup
var mu sync.Mutex
results := make(map[string]*FileBlameResult)

ctx, cancel := context.WithTimeout(parentCtx, 10*time.Minute)
defer cancel()

for i := 0; i < pool; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for path := range fileCh {
            if ctx.Err() != nil {
                return
            }
            result := blameFile(ctx, repo, commitHash, path)
            mu.Lock()
            results[path] = result
            mu.Unlock()
        }
    }()
}

for _, f := range files {
    fileCh <- f
}
close(fileCh)
wg.Wait()

partial := ctx.Err() != nil
```

### Redis Key Examples

```
octocat/hello-world:abc123:contributors     → ContributorResult JSON (TTL 48h)
active:octocat/hello-world:abc123:blame     → job-uuid (TTL 10min)
job:550e8400-e29b-...                       → AnalysisJob JSON (TTL 1h)
```
