# Implementation Plan: Core Analysis Engine

**Branch**: `001-core-analysis` | **Date**: 2026-03-23 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-core-analysis/spec.md`

## Summary

Build the foundational analysis pillar for GRIT: a Go backend that accepts
a GitHub `owner/repo` identifier, shallow-clones the repository via go-git,
walks the file tree to count lines (total/code/comment/blank) per file
grouped by language, fetches GitHub metadata and 52-week commit activity via
the REST API, caches results in Redis (24h TTL), and processes all work
asynchronously via NATS JetStream. Exposes four JSON endpoints: full
analysis, job status polling, shields.io badge, and cache invalidation.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: go-chi/chi v5 (router), go-git/go-git v5 (clone),
redis/go-redis v9 (cache), nats-io/nats.go with JetStream (job queue),
joho/godotenv (config), prometheus/client_golang (metrics)
**Storage**: Redis 7 (cache only — no relational database)
**Testing**: go test with testify/assert, fixture Git repo in testdata/
**Target Platform**: Linux (Docker container), single-stage build, port 8080
**Project Type**: Web service (JSON API)
**Performance Goals**: <2 min analysis for repos ≤10k files; <50ms p95
for cached responses
**Constraints**: 5-min hard timeout per analysis job; shallow clone (depth=1);
in-memory clone for repos <50 MB, disk clone to /tmp for larger repos
**Scale/Scope**: Single VPS deployment via Docker Compose; concurrent
analysis of multiple repos; self-hosted

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. API-First Design | ✅ PASS | All 4 endpoints return JSON; no SSR; Go serves static frontend files |
| II. Modular Pillars | ✅ PASS | Core pillar isolated in `internal/analysis/core/`; no cross-pillar imports |
| III. Async-First | ✅ PASS | NATS JetStream on `grit.jobs.analysis`; 202 + polling pattern |
| IV. Cache-First Redis | ✅ PASS | Redis keyed by `{owner}/{repo}:{sha}:core`; 24h TTL |
| V. Defensive AI | ✅ N/A | Core pillar has no AI dependency |
| VI. Self-Hostable | ✅ PASS | Docker Compose; env-var config via godotenv; no cloud deps |
| VII. Clean Handlers | ✅ PASS | Handlers delegate to service layer in `internal/`; consistent error envelope |
| VIII. Test Discipline | ✅ PASS | testify/assert; table-driven tests; fixture repo in testdata/ |

**Gate result: PASS** — no violations. Proceeding to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/001-core-analysis/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (API endpoint contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── grit/
    └── main.go              # Entrypoint: chi router, NATS connect, Redis connect

internal/
├── analysis/
│   └── core/
│       ├── analyzer.go      # Orchestrates clone → walk → count → aggregate
│       ├── counter.go       # Line counting logic (code/comment/blank)
│       ├── languages.go     # Extension → language mapping (40+ langs)
│       └── walker.go        # File tree walker with .gitignore support
├── cache/
│   └── redis.go             # Redis get/set/delete with TTL management
├── clone/
│   └── clone.go             # go-git shallow clone (memory vs disk strategy)
├── config/
│   └── config.go            # Env-var config via godotenv
├── github/
│   ├── client.go            # Typed HTTP client wrapping net/http
│   ├── metadata.go          # Repo metadata fetcher (REST)
│   └── stats.go             # Commit activity fetcher (Stats API + 202 retry)
├── handler/
│   ├── analysis.go          # GET /api/:owner/:repo handler
│   ├── badge.go             # GET /api/:owner/:repo/badge handler
│   ├── cache.go             # DELETE /api/:owner/:repo/cache handler
│   └── status.go            # GET /api/:owner/:repo/status handler
├── job/
│   ├── publisher.go         # NATS JetStream publisher (with dedup)
│   └── worker.go            # NATS JetStream subscriber / job executor
├── metrics/
│   └── prometheus.go        # Prometheus metrics registration + /metrics
└── models/
    ├── analysis.go          # AnalysisResult, FileStats, LanguageBreakdown
    ├── job.go               # AnalysisJob, JobStatus enum
    └── repository.go        # Repository, GitHubMetadata, CommitActivity

testdata/
└── fixture-repo/            # Small Git repo for integration tests

Dockerfile
docker-compose.yml
.env.example
go.mod
go.sum
```

**Structure Decision**: Go standard layout with `cmd/` for the entrypoint
and `internal/` for all domain packages. Constitution Principle II requires
the core pillar in its own package (`internal/analysis/core/`). Shared types
live in `internal/models/`. No frontend code in this feature — backend only.

## Complexity Tracking

> No violations detected — table intentionally left empty.
