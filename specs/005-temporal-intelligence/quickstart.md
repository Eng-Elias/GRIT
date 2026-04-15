# Quickstart: Temporal Intelligence

## Prerequisites

- Go 1.22+
- Docker Compose (Redis 7 + NATS with JetStream)
- GitHub token (optional — required for PR merge time data)

## Development Workflow

### 1. Start the Stack

```bash
docker compose up -d redis nats
export REDIS_URL=localhost:6379
export NATS_URL=nats://localhost:4222
export GITHUB_TOKEN=ghp_your_token_here  # optional
go run ./cmd/grit
```

### 2. Run Tests

```bash
# Temporal analysis unit tests
go test ./internal/analysis/temporal/... -v

# GitHub GraphQL client tests
go test ./internal/github/... -v

# Handler tests
go test ./internal/handler/... -v -run TestTemporal

# All tests
go test ./... -count=1
```

### 3. Manual Endpoint Testing

```bash
# Trigger core analysis (auto-enqueues temporal job)
curl http://localhost:8080/api/golang/go

# Poll status
curl http://localhost:8080/api/golang/go/status

# Get temporal data (after job completes)
curl http://localhost:8080/api/golang/go/temporal

# Get temporal data with 1-year period
curl "http://localhost:8080/api/golang/go/temporal?period=1y"

# Verify temporal_summary in main response
curl http://localhost:8080/api/golang/go | jq '.temporal_summary'

# Delete cache (includes temporal)
curl -X DELETE http://localhost:8080/api/golang/go/cache
```

## Pipeline Flow

```
POST /api/:owner/:repo
  → Core analysis job (grit.jobs.analysis)
    → Core worker completes
      → Publishes grit.jobs.complexity (parallel)
      → Publishes grit.jobs.churn (parallel, via complexity worker)
      → Publishes grit.jobs.blame (parallel)
      → Publishes grit.jobs.temporal (parallel)  ← NEW
        → Temporal worker:
          1. Walk commit log for monthly boundaries (LOC snapshots)
          2. Walk commit log for weekly stats (velocity + author activity)
          3. Compute rolling 4-week commit cadence
          4. Fetch PR merge times via GitHub GraphQL (optional)
          5. Detect refactor periods from weekly data
          6. Cache result in Redis (12h TTL)
```

## Key Implementation Notes

### Monthly LOC Sampling
- Use `git.Log` to walk commits chronologically
- For each month boundary, find the latest commit on or before the 1st
- Walk `commit.Tree().Files()` to count lines per file
- Detect language via `blame.LanguageForFile(path)` — reuse existing extension registry
- Top 10 languages by LOC + "Other" bucket

### Weekly Velocity
- Walk commits for past 52 weeks
- Use `commit.Stats()` for per-commit additions/deletions (same API as churn)
- Bucket by ISO week (`time.ISOWeek()`)
- Track per-author activity during the same walk

### PR Merge Time
- GitHub GraphQL API: `POST https://api.github.com/graphql`
- Cursor-based pagination, page size 25, up to 4 pages (100 PRs)
- Extract `createdAt` → `mergedAt` duration
- Compute median, p75, p95 percentiles
- Gracefully degrade to `null` on any error

### Redis Key Examples

```
facebook/react:abc123:temporal:3y     → JSON TemporalResult (TTL 12h)
facebook/react:abc123:temporal:1y     → JSON TemporalResult (TTL 12h)
active:facebook/react:abc123:temporal → Job ID (TTL 30m)
```

### Period Parameter
- `?period=3y` (default): 36 monthly snapshots
- `?period=2y`: 24 monthly snapshots
- `?period=1y`: 12 monthly snapshots
- Velocity data (52 weeks) is independent of the period parameter
- Each period produces a separate cache entry
