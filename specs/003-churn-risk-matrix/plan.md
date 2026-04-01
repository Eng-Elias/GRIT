# Implementation Plan: Git Churn Analysis, Risk Matrix & Dead Code Estimation

**Branch**: `003-churn-risk-matrix` | **Date**: 2026-04-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-churn-risk-matrix/spec.md`

## Summary

Add churn analysis as GRIT's third analysis pillar. Walk the default branch commit
log using go-git's `CommitIterator`, building a `map[string]int` for per-file churn
counts. Merge churn data with complexity results from Redis to produce a risk matrix
with percentile-based risk levels. Detect potentially unused files via a heuristic
combining commit recency, file extension filtering, and `strings.Contains` import
scanning. Serve all results from a single endpoint `GET /api/:owner/:repo/churn-matrix`.
Churn jobs run asynchronously via NATS subject `grit.jobs.churn`, enqueued by the
complexity worker on completion, with a 30-second requeue delay if complexity data
is not yet available.

## Technical Context

**Language/Version**: Go 1.22
**Primary Dependencies**: go-git v5 (existing), chi router (existing), go-redis/v9 (existing), nats.go (existing)
**Storage**: Redis 7 (cache key `{owner}/{repo}:{sha}:churn`, TTL 24h)
**Testing**: `go test` with table-driven tests, fixture repo, mocked externals
**Target Platform**: Linux server (Docker Compose)
**Project Type**: Web service (new analysis pillar within existing Go backend)
**Performance Goals**: 5,000 commits analyzed within 60 seconds; cached results returned <50ms p95
**Constraints**: 5-minute analysis timeout; commit window capped at 2 years OR 5,000 commits
**Scale/Scope**: Repos up to 100k files, commit histories up to 5,000 commits per analysis window

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. API-First Design** | ✅ PASS | New JSON endpoint `GET /api/:owner/:repo/churn-matrix`; `churn_summary` embedded in existing analysis response |
| **II. Modular Analysis Pillars** | ✅ PASS | New pillar under `internal/analysis/churn/`; no imports from other pillars; shared types in `internal/models/` |
| **III. Async-First Execution** | ✅ PASS | Churn jobs published to NATS subject `grit.jobs.churn`; auto-triggered after complexity completion; returns 202 if pending |
| **IV. Cache-First with Redis** | ✅ PASS | Redis key `{owner}/{repo}:{sha}:churn`, 24h TTL per constitution |
| **V. Defensive AI Integration** | ✅ N/A | No AI integration in this pillar |
| **VI. Self-Hostable by Default** | ✅ PASS | No new infrastructure; uses existing go-git for commit log walking |
| **VII. Clean Handler Separation** | ✅ PASS | Handler does request parsing + cache check only; all churn/risk/stale logic in `internal/analysis/churn/` |
| **VIII. Test Discipline** | ✅ PASS | Table-driven tests for churn counting, risk classification, stale detection; fixture repo integration test |

**No new dependencies**: This feature uses only existing dependencies (go-git, go-redis, nats.go, chi). No dependency justification needed.

**GATE RESULT**: ✅ ALL PASS — proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/003-churn-risk-matrix/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── api.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── analysis/
│   ├── core/                    # Existing (feature 001)
│   ├── complexity/              # Existing (feature 002)
│   └── churn/                   # NEW — this feature
│       ├── analyzer.go          # ChurnAnalyzer: orchestrates churn + risk + stale detection
│       ├── churn.go             # Commit log walker: builds map[string]int churn counts
│       ├── risk.go              # Risk matrix: joins churn × complexity, classifies risk levels
│       ├── stale.go             # Stale file detection: recency + extension + import heuristic
│       ├── percentile.go        # Sort-based percentile calculation (p50, p75, p90)
│       ├── analyzer_test.go     # Integration test with fixture repo
│       ├── churn_test.go        # Churn counting tests
│       ├── risk_test.go         # Risk classification tests
│       ├── stale_test.go        # Stale file detection tests
│       └── percentile_test.go   # Percentile calculation tests
├── cache/
│   └── redis.go                 # MODIFIED — add GetChurn/SetChurn/DeleteChurn + active churn job methods
├── handler/
│   ├── churn.go                 # NEW — GET /api/:owner/:repo/churn-matrix handler
│   ├── churn_test.go            # NEW
│   ├── analysis.go              # MODIFIED — embed churn_summary in response
│   └── cache.go                 # MODIFIED — extend deletion to clear churn cache
├── job/
│   ├── publisher.go             # MODIFIED — add PublishChurn method
│   ├── complexity_worker.go     # MODIFIED — auto-trigger churn job after complexity completion
│   ├── churn_worker.go          # NEW — churn job consumer with 30s requeue logic
│   └── worker_test.go           # MODIFIED — add churn worker + publisher tests
├── models/
│   └── churn.go                 # NEW — FileChurn, RiskEntry, StaleFile, ChurnMatrixResult, ChurnSummary
└── metrics/
    └── prometheus.go            # MODIFIED — add churn_analysis_duration_seconds
```

**Structure Decision**: Follows Constitution Principle II — new pillar is an independent
package under `internal/analysis/churn/`. Shared types in `internal/models/churn.go`.
No cross-pillar imports. Handler separation per Principle VII.

## Complexity Tracking

> No constitution violations — this section is intentionally empty.
