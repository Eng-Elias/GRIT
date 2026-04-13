# Implementation Plan: Contributor Attribution & Bus Factor Analysis

**Branch**: `004-contributor-bus-factor` | **Date**: 2026-04-04 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-contributor-bus-factor/spec.md`

## Summary

Add contributor attribution and bus factor analysis to GRIT. The system runs go-git blame on every source file at HEAD using a goroutine pool (one file per worker, `min(4, NumCPU())`), attributes lines to authors by email (case-insensitive dedup), computes per-author statistics and a bus factor (minimum authors for 80% ownership), and provides a per-file top-3 breakdown. Results are collected with a mutex-protected map. A 10-minute `context.WithTimeout` governs the entire job; partial results are stored on timeout with `partial: true`. Redis TTL is 48h. The blame job is published to NATS subject `grit.jobs.blame` by the **core analysis worker** after core analysis completes, independent of complexity and churn pipelines.

## Technical Context

**Language/Version**: Go 1.22  
**Primary Dependencies**: go-git v5 (`go-git/go-git/v5/git.Blame`), chi router, NATS JetStream, Redis 7  
**Storage**: Redis 7 (cache only, key pattern `{owner}/{repo}:{sha}:contributors`, TTL 48h)  
**Testing**: `go test` with table-driven tests, `testify/assert` + `testify/require`  
**Target Platform**: Linux server (Docker Compose), self-hostable  
**Project Type**: Web service (Go backend + React frontend)  
**Performance Goals**: Blame analysis for 500 source files within 10-minute hard timeout  
**Constraints**: `CGO_ENABLED=0` (pure Go only), 10-minute job timeout, partial results on timeout  
**Scale/Scope**: Repositories up to ~500 source files; goroutine pool capped at `min(4, NumCPU())`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. API-First Design | ✅ PASS | New JSON endpoints: `/contributors`, `/contributors/files`. Summary embedded in main endpoint. No SSR. |
| II. Modular Analysis Pillars | ✅ PASS | New `internal/analysis/blame/` package. No cross-pillar imports. Shared types in `internal/models/`. |
| III. Async-First Execution | ✅ PASS | NATS subject `grit.jobs.blame`. Returns 202 with job ID. Workers horizontally scalable. |
| IV. Cache-First with Redis | ✅ PASS | Key: `{owner}/{repo}:{sha}:contributors`. TTL 48h (matches constitution's "Contributor (blame): 48h"). |
| V. Defensive AI Integration | N/A | No AI integration in this feature. |
| VI. Self-Hostable by Default | ✅ PASS | No new external dependencies. Config via environment variables. |
| VII. Clean Handler Separation | ✅ PASS | Handlers do request parsing + cache check + response. Domain logic in `internal/analysis/blame/`. |
| VIII. Test Discipline | ✅ PASS | Table-driven unit tests for blame, bus factor, aggregation. Integration test against fixture repo. |

**Gate result**: All applicable gates pass. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/004-contributor-bus-factor/
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
│   └── blame/                  # NEW — contributor analysis pillar
│       ├── blame.go            # Per-file blame logic (go-git Blame wrapper)
│       ├── blame_test.go
│       ├── aggregator.go       # Author dedup, stats, bus factor calc
│       ├── aggregator_test.go
│       ├── analyzer.go         # Orchestrator: pool, timeout, partial results
│       └── analyzer_test.go
├── cache/
│   └── redis.go                # MODIFIED — add contributor cache methods
├── handler/
│   ├── contributors.go         # NEW — /contributors + /contributors/files handlers
│   ├── contributors_test.go    # NEW
│   └── analysis.go             # MODIFIED — embed contributor_summary
├── job/
│   ├── publisher.go            # MODIFIED — add PublishBlame()
│   ├── blame_worker.go         # NEW — NATS consumer for blame jobs
│   └── worker.go               # MODIFIED — auto-trigger blame after core
├── models/
│   └── contributor.go          # NEW — Author, ContributorResult, ContributorSummary, FileContributors
└── metrics/
    └── metrics.go              # MODIFIED — add blame duration + counter
```

**Structure Decision**: Follows existing pillar pattern — new `internal/analysis/blame/` package with its own analyzer, mirroring `churn/` and `complexity/`. New models in `internal/models/contributor.go`. New handler in `internal/handler/contributors.go`. New worker in `internal/job/blame_worker.go`. Core worker modified to enqueue blame job (parallel to complexity).

## Complexity Tracking

No constitution violations. No entries needed.
