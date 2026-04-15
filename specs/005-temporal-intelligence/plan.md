# Implementation Plan: Temporal Intelligence

**Branch**: `005-temporal-intelligence` | **Date**: 2026-04-14 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-temporal-intelligence/spec.md`

## Summary

Add temporal analysis to GRIT — tracking repository evolution over time. The system performs three analyses: (1) monthly LOC snapshots going back up to 3 years using `git log --first-parent` on the default branch to resolve monthly boundary commits, then walking each commit tree via go-git in-memory storage to count lines per language; (2) weekly velocity metrics including additions/deletions, per-author activity, rolling 4-week commit cadence, and PR merge time via GitHub GraphQL with cursor-based pagination; (3) refactor detection by identifying consecutive weeks with negative net LOC and above-median commit counts. Results are cached in Redis with 12h TTL. The temporal job is published to NATS subject `grit.jobs.temporal` by the core analysis worker alongside blame, and processes checkpoints sequentially.

## Technical Context

**Language/Version**: Go 1.22
**Primary Dependencies**: go-git v5 (commit log walking, in-memory tree traversal, DiffTree), chi router, NATS JetStream, Redis 7, GitHub GraphQL API (`github.com/shurcooL/graphql` or `net/http` with raw queries)
**Storage**: Redis 7 (cache only, key pattern `{owner}/{repo}:{sha}:temporal`, TTL 12h)
**Testing**: `go test` with table-driven tests, `testify/assert` + `testify/require`
**Target Platform**: Linux server (Docker Compose), self-hostable
**Project Type**: Web service (Go backend + React frontend)
**Performance Goals**: 3-year LOC analysis on a 1000-file repository within 5 minutes
**Constraints**: `CGO_ENABLED=0` (pure Go only), sequential monthly checkpoint processing, GitHub GraphQL rate limits
**Scale/Scope**: Repositories up to 3 years of history; 36 monthly checkpoints max for LOC, 52 weekly checkpoints for velocity

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. API-First Design | PASS | New JSON endpoint: `/temporal`. Summary embedded in main endpoint. No SSR. |
| II. Modular Analysis Pillars | PASS | New `internal/analysis/temporal/` package. No cross-pillar imports. Shared types in `internal/models/`. |
| III. Async-First Execution | PASS | NATS subject `grit.jobs.temporal`. Returns 202 with job ID. Worker horizontally scalable. |
| IV. Cache-First with Redis | PASS | Key: `{owner}/{repo}:{sha}:temporal`. TTL 12h (matches constitution's "Temporal: 12h"). |
| V. Defensive AI Integration | N/A | No AI integration in this feature. |
| VI. Self-Hostable by Default | PASS | GitHub GraphQL is optional (PR data gracefully omitted). Config via environment variables. |
| VII. Clean Handler Separation | PASS | Handlers do request parsing + cache check + response. Domain logic in `internal/analysis/temporal/`. |
| VIII. Test Discipline | PASS | Table-driven unit tests for LOC counting, velocity, refactor detection. Integration test against fixture repo. |

**Gate result**: All applicable gates pass. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/005-temporal-intelligence/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (API contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── analysis/
│   └── temporal/                  # NEW — temporal analysis pillar
│       ├── loc.go                 # Monthly LOC snapshots (tree walking, line counting)
│       ├── loc_test.go
│       ├── velocity.go            # Weekly diffs, per-author activity, commit cadence
│       ├── velocity_test.go
│       ├── refactor.go            # Refactor period detection
│       ├── refactor_test.go
│       ├── analyzer.go            # Orchestrator: coordinate LOC + velocity + refactor
│       └── analyzer_test.go
├── github/
│   ├── graphql.go                 # NEW — GitHub GraphQL client for PR merge times
│   └── graphql_test.go            # NEW
├── cache/
│   └── redis.go                   # MODIFIED — add temporal cache methods
├── handler/
│   ├── temporal.go                # NEW — /temporal handler
│   ├── temporal_test.go           # NEW
│   └── analysis.go                # MODIFIED — embed temporal_summary
├── job/
│   ├── publisher.go               # MODIFIED — add PublishTemporal()
│   ├── temporal_worker.go         # NEW — NATS consumer for temporal jobs
│   └── worker.go                  # MODIFIED — auto-trigger temporal after core
├── models/
│   └── temporal.go                # NEW — MonthlySnapshot, WeeklyActivity, TemporalResult, etc.
└── metrics/
    └── prometheus.go              # MODIFIED — add temporal duration + counter
```

**Structure Decision**: Follows existing pillar pattern — new `internal/analysis/temporal/` package with its own analyzer, mirroring `blame/`, `churn/`, and `complexity/`. New models in `internal/models/temporal.go`. New handler in `internal/handler/temporal.go`. New worker in `internal/job/temporal_worker.go`. Core worker modified to enqueue temporal job (parallel to complexity, churn, blame). GitHub GraphQL client added as `internal/github/graphql.go` alongside existing REST client.

## Complexity Tracking

No constitution violations. No entries needed.
