# Implementation Plan: AST-Based Code Complexity Analysis

**Branch**: `002-complexity-analysis` | **Date**: 2026-03-30 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-complexity-analysis/spec.md`

## Summary

Add AST-based structural complexity analysis as GRIT's second analysis pillar.
Uses `github.com/smacker/go-tree-sitter` with per-language grammar packages to
parse source files and compute cyclomatic complexity, cognitive complexity,
function counts, and per-function metrics. Files are parsed in parallel via a
`runtime.NumCPU()` worker pool. Results are collected into `[]FileComplexity`,
aggregated into repository-level statistics (mean, median, p90, hot files,
distribution histogram), cached in Redis under a separate `:complexity` key,
and served via `GET /api/:owner/:repo/complexity`. Complexity jobs are enqueued
via NATS after core analysis jobs signal completion.

## Technical Context

**Language/Version**: Go 1.22  
**Primary Dependencies**: go-tree-sitter (`github.com/smacker/go-tree-sitter`), per-language grammar packages (go, typescript, javascript, python, rust, java, c, cpp, ruby), chi router, go-redis/v9, nats.go  
**Storage**: Redis 7 (cache key `{owner}/{repo}:{sha}:complexity`, TTL 24h)  
**Testing**: `go test` with table-driven tests, fixture repo, mocked externals  
**Target Platform**: Linux server (Docker Compose)  
**Project Type**: Web service (new analysis pillar within existing Go backend)  
**Performance Goals**: 5,000 supported-language files analyzed within 3 minutes; cached results returned <50ms p95  
**Constraints**: 5-minute analysis timeout (shared with core pillar); worker pool bounded by `runtime.NumCPU()`  
**Scale/Scope**: Same as core pillar — repos up to 100k files, 9 supported languages

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. API-First Design** | ✅ PASS | New JSON endpoint `GET /api/:owner/:repo/complexity`; complexity_summary embedded in existing analysis response |
| **II. Modular Analysis Pillars** | ✅ PASS | New pillar under `internal/analysis/complexity/`; no imports from `internal/analysis/core/`; shared types in `internal/models/` |
| **III. Async-First Execution** | ✅ PASS | Complexity jobs published to NATS subject `grit.jobs.complexity`; auto-triggered after core completion; returns 202 if pending |
| **IV. Cache-First with Redis** | ✅ PASS | Redis key `{owner}/{repo}:{sha}:complexity`, 24h TTL per constitution |
| **V. Defensive AI Integration** | ✅ N/A | No AI integration in this pillar |
| **VI. Self-Hostable by Default** | ✅ PASS | No new infrastructure; tree-sitter is a Go dependency compiled into the binary |
| **VII. Clean Handler Separation** | ✅ PASS | Handler does request parsing + cache check only; all complexity logic in `internal/analysis/complexity/` |
| **VIII. Test Discipline** | ✅ PASS | Table-driven tests per language; fixture repo integration test; mocked externals in unit tests |

**New dependency justification**: `github.com/smacker/go-tree-sitter` + 9 grammar packages. Required for AST parsing — no Go-native alternative exists that supports 9 languages. These are compile-time dependencies with no runtime services; the tree-sitter C library is statically linked via cgo.

**GATE RESULT**: ✅ ALL PASS — proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/002-complexity-analysis/
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
│   └── complexity/              # NEW — this feature
│       ├── analyzer.go          # ComplexityAnalyzer: orchestrates parsing + aggregation
│       ├── parser.go            # Tree-sitter parsing: per-file AST → FunctionComplexity
│       ├── languages.go         # Language registry: maps extensions → tree-sitter grammars + node queries
│       ├── cyclomatic.go        # Cyclomatic complexity calculation from AST nodes
│       ├── cognitive.go         # Cognitive complexity calculation with nesting weights
│       ├── aggregator.go        # Aggregates []FileComplexity → hot files, stats, histogram
│       ├── pool.go              # Worker pool: fan-out file parsing across NumCPU goroutines
│       ├── analyzer_test.go     # Integration test with fixture repo
│       ├── parser_test.go       # Per-language parsing tests
│       ├── cyclomatic_test.go   # Cyclomatic calculation tests
│       ├── cognitive_test.go    # Cognitive calculation tests
│       └── aggregator_test.go   # Aggregation + hot files tests
├── cache/
│   └── redis.go                 # MODIFIED — add GetComplexity/SetComplexity/DeleteComplexity
├── handler/
│   ├── complexity.go            # NEW — GET /api/:owner/:repo/complexity handler
│   ├── complexity_test.go       # NEW
│   └── analysis.go              # MODIFIED — embed complexity_summary in response
├── job/
│   ├── publisher.go             # MODIFIED — add PublishComplexity method
│   ├── worker.go                # MODIFIED — add complexity worker + auto-trigger after core
│   └── worker_test.go           # MODIFIED — add complexity job tests
├── models/
│   └── complexity.go            # NEW — FileComplexity, FunctionComplexity, ComplexityResult, etc.
└── metrics/
    └── metrics.go               # MODIFIED — add complexity_analysis_duration_seconds
```

**Structure Decision**: Follows Constitution Principle II — new pillar is an independent
package under `internal/analysis/complexity/`. Shared types in `internal/models/complexity.go`.
No cross-pillar imports. Handler separation per Principle VII.

## Complexity Tracking

> No constitution violations — this section is intentionally empty.
