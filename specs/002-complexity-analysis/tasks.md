# Tasks: AST-Based Code Complexity Analysis

**Input**: Design documents from `/specs/002-complexity-analysis/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included per Constitution Principle VIII (Test Discipline). Table-driven tests, fixture repo integration, mocked externals.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Add tree-sitter dependencies and create package structure

- [x] T001 Add go-tree-sitter and all 9 grammar packages to go.mod via `go get github.com/smacker/go-tree-sitter@latest` and per-language grammar packages (golang, typescript, javascript, python, rust, java, c, cpp, ruby)
- [x] T002 Create package directory `internal/analysis/complexity/`
- [x] T003 [P] Create complexity model types (FunctionComplexity, FileComplexity, ComplexityResult, ComplexityDistribution, ComplexitySummary) in `internal/models/complexity.go`
- [x] T004 [P] Add complexity fixture files to `testdata/fixture-repo/`: a Go file with known complexity (nested if/for/switch), a Python file with known complexity, and a JS file with known complexity — each with pre-calculated expected cyclomatic and cognitive values documented in comments

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Language registry and tree-sitter parsing infrastructure that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Implement language registry in `internal/analysis/complexity/languages.go`: map file extensions → tree-sitter `*sitter.Language` + function node type names + decision point node type names per language (Go, TypeScript, JavaScript, Python, Rust, Java, C, C++, Ruby) per research.md R1 table
- [x] T006 Implement tree-sitter file parser in `internal/analysis/complexity/parser.go`: accept file path + source bytes + language config, parse with `sitter.NewParser()`, walk AST to extract []FunctionComplexity per file; return FileComplexity struct; skip files that fail to parse (log warning, return nil)
- [x] T007 [P] Write table-driven tests for language registry in `internal/analysis/complexity/languages_test.go`: verify all 9 languages return non-nil Language, verify unsupported extensions return nil, verify function node types are correctly mapped
- [x] T008 [P] Write table-driven parser tests in `internal/analysis/complexity/parser_test.go`: parse each fixture file (Go, Python, JS), verify function names extracted, verify function count matches expected, verify parse error returns nil without panic

**Checkpoint**: Tree-sitter parsing infrastructure ready — can extract functions from source files in all 9 languages

---

## Phase 3: User Story 1 — Per-File Complexity Analysis (Priority: P1) 🎯 MVP

**Goal**: Parse each supported-language file via AST, compute cyclomatic + cognitive complexity per function, aggregate to per-file metrics

**Independent Test**: Call complexity analyzer on fixture repo and verify per-file metrics (cyclomatic, cognitive, function count, avg, max) match pre-calculated expected values

### Tests for User Story 1

- [x] T009 [P] [US1] Write table-driven cyclomatic complexity tests in `internal/analysis/complexity/cyclomatic_test.go`: test Go snippet with if/for/switch/&&/|| → verify CC = 1 + decision_points per function; test Python snippet with if/elif/for/and/or; test JS snippet with ternary/while/for; test empty function → CC=1; test function with only logical operators
- [x] T010 [P] [US1] Write table-driven cognitive complexity tests in `internal/analysis/complexity/cognitive_test.go`: test flat if → cognitive=1; test nested if-in-if → cognitive=1+(1+1)=3; test triple-nested loop → verify nesting multiplier; test else clause increments but doesn't nest; test logical operator sequences (same op = +1, mixed ops = +2)

### Implementation for User Story 1

- [x] T011 [US1] Implement cyclomatic complexity calculator in `internal/analysis/complexity/cyclomatic.go`: walk AST nodes within a function, count decision points (if, else if, for, while, switch case, &&, ||, ternary) per research.md R2 table, return 1 + count; language-specific node type names come from languages.go registry
- [x] T012 [US1] Implement cognitive complexity calculator in `internal/analysis/complexity/cognitive.go`: recursive AST walker tracking nesting depth; increment +1 for each flow break; add +nesting_level for nesting-eligible nodes (if, for, while, switch, catch); increase nesting when entering if/else/for/while/switch/catch/lambda per research.md R3 rules
- [x] T013 [US1] Integrate cyclomatic + cognitive into parser.go: after extracting functions, call cyclomatic and cognitive calculators for each function's AST subtree; populate FunctionComplexity.Cyclomatic and .Cognitive; compute FileComplexity derived fields (avg_function_complexity, max_function_complexity, complexity_density)
- [x] T014 [US1] Write integration test in `internal/analysis/complexity/analyzer_test.go`: use fixture-repo files, call parser on each, verify per-file cyclomatic/cognitive/function_count match expected values; verify unsupported language files are skipped; verify file with syntax error returns nil gracefully

**Checkpoint**: Per-file complexity analysis works for all 9 languages with correct cyclomatic and cognitive scores

---

## Phase 4: User Story 2 — Hot Files and Repository-Level Aggregation (Priority: P1)

**Goal**: Aggregate []FileComplexity into hot_files ranking, repository-level stats (mean, median, p90), and distribution histogram

**Independent Test**: Create []FileComplexity with known values, run aggregator, verify hot_files order, aggregate math, and histogram bucket counts

### Tests for User Story 2

- [x] T015 [P] [US2] Write table-driven aggregator tests in `internal/analysis/complexity/aggregator_test.go`: test hot_files sorted by density descending; test max 20 hot files cap; test mean/median/p90 calculation with known inputs; test distribution histogram bucket boundaries (Low 1-5, Medium 6-10, High 11-20, Critical 21+); test empty input → zero aggregates; test files with 0 functions excluded from aggregates and distribution

### Implementation for User Story 2

- [x] T016 [US2] Implement aggregator in `internal/analysis/complexity/aggregator.go`: accept []FileComplexity, compute mean/median/p90 cyclomatic (excluding 0-function files), sort by complexity_density descending → top 20 hot_files (omit functions array in hot_files), count distribution histogram buckets by avg_function_complexity, return ComplexityResult with all aggregate fields populated
- [x] T017 [US2] Add ComplexityResult builder in `internal/analysis/complexity/aggregator.go`: accept repository info + []FileComplexity, call aggregate functions, set analyzed_at timestamp, return fully populated ComplexityResult

**Checkpoint**: Aggregation produces correct hot_files, stats, and histogram from per-file data

---

## Phase 5: User Story 3 — Async Complexity Job Triggered After Core Analysis (Priority: P2)

**Goal**: Auto-enqueue complexity job via NATS after core analysis completes; cache results in Redis; serve via dedicated endpoint and embed summary in main response

**Independent Test**: Trigger core analysis, verify complexity job auto-fires, verify `GET /api/:owner/:repo/complexity` returns 200, verify `GET /api/:owner/:repo` includes `complexity_summary`

### Tests for User Story 3

- [x] T018 [P] [US3] Write cache complexity tests in `internal/handler/complexity_test.go`: test GET /api/:owner/:repo/complexity returns 200 with cached data and X-Cache HIT; test returns 404 when core analysis missing; test returns 202 when complexity job in progress; test 400 for invalid owner/repo
- [x] T019 [P] [US3] Write complexity worker tests in `internal/job/worker_test.go` (append to existing): test complexity job processes message, caches result, marks job completed; test complexity job failure marks job failed; test auto-trigger publishes complexity job after core completion

### Implementation for User Story 3

- [x] T020 [US3] Add complexity cache methods to `internal/cache/redis.go`: `complexityKey(owner, repo, sha)` returning `{owner}/{repo}:{sha}:complexity`; `GetComplexity(ctx, owner, repo, sha)` with SCAN fallback for empty SHA (same pattern as findAnalysis); `SetComplexity(ctx, owner, repo, sha, data)` with 24h TTL; `DeleteComplexity(ctx, owner, repo)` via SCAN pattern delete
- [x] T021 [US3] Add NATS complexity subject constant and PublishComplexity method in `internal/job/publisher.go`: subject `grit.jobs.complexity`; ensure stream config includes `grit.jobs.complexity` subject; PublishComplexity creates job payload and publishes to complexity subject
- [x] T022 [US3] Implement ComplexityAnalyzer orchestrator in `internal/analysis/complexity/analyzer.go`: accept clone dir path + []FileStats (from core analysis), filter to supported languages, parse each file sequentially (pool added in US4), aggregate results, return ComplexityResult
- [x] T023 [US3] Add complexity worker to `internal/job/worker.go`: new ComplexityWorker struct with Start method subscribing to `grit.jobs.complexity` with durable consumer `grit-complexity-worker`; processComplexityMessage reads clone dir, runs ComplexityAnalyzer, caches result, updates job status
- [x] T024 [US3] Add auto-trigger in core worker `internal/job/worker.go`: after successful core job completion (after cache set + job completion update), call publisher.PublishComplexity with same owner/repo/sha/token to enqueue complexity analysis
- [x] T025 [US3] Implement complexity handler in `internal/handler/complexity.go`: GET /api/:owner/:repo/complexity; validate owner/repo; check complexity cache → 200 with X-Cache HIT; check active complexity job → 202 with job status; check if core analysis exists → 404 if not; otherwise return 404 (no complexity data yet)
- [x] T026 [US3] Modify analysis handler in `internal/handler/analysis.go`: after serving cached core result, look up complexity cache; if found, build ComplexitySummary with status "complete" + aggregate fields; if not found, build ComplexitySummary with status "pending"; attach as `complexity_summary` field to response JSON
- [x] T027 [US3] Modify cache deletion in `internal/handler/cache.go` or `internal/cache/redis.go`: extend DeleteAnalysis (or add call alongside it) to also call DeleteComplexity so cache bust clears both core and complexity
- [x] T028 [US3] Register complexity handler and worker in `cmd/grit/main.go`: create ComplexityAnalyzer, create ComplexityWorker with Start(ctx); add route `r.Get("/complexity", complexityHandler.HandleComplexity)` inside existing `/api/{owner}/{repo}` route group
- [x] T029 [US3] Add `grit_complexity_analysis_duration_seconds` histogram metric in `internal/metrics/metrics.go`; observe duration in complexity worker after analysis completes
- [x] T030 [US3] Update NATS stream config in `internal/job/publisher.go` EnsureStream: add `grit.jobs.complexity` to stream subjects list

**Checkpoint**: Full async pipeline works — core completion triggers complexity job, results cached and served via API

---

## Phase 6: User Story 4 — Parallel File Parsing (Priority: P2)

**Goal**: Parse files in parallel using runtime.NumCPU() bounded worker pool for performance

**Independent Test**: Analyze a repository with many files and verify wall-clock time is reduced vs sequential; verify results are identical to sequential run

### Tests for User Story 4

- [x] T031 [P] [US4] Write worker pool tests in `internal/analysis/complexity/pool_test.go`: test pool processes all files and collects results; test pool continues on parse error (result excluded, error logged); test pool respects context cancellation; test pool with 0 files returns empty slice

### Implementation for User Story 4

- [x] T032 [US4] Implement worker pool in `internal/analysis/complexity/pool.go`: fan-out pattern with runtime.NumCPU() goroutines reading from files channel, each creating own sitter.Parser; results sent to results channel; collector goroutine assembles []FileComplexity; per-file errors logged and skipped; context cancellation stops all workers
- [x] T033 [US4] Integrate pool into ComplexityAnalyzer in `internal/analysis/complexity/analyzer.go`: replace sequential file parsing loop with pool.Run(ctx, files) call; pass clone dir and language registry to pool

**Checkpoint**: Parallel parsing operational — performance improved on multi-core systems

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T034 [P] Ensure Dockerfile build stage includes `build-base` (or equivalent C compiler) for cgo/tree-sitter compilation in `Dockerfile`
- [x] T035 [P] Add edge case handling in analyzer: zero supported files → valid empty ComplexityResult; context timeout → return partial results for files analyzed so far
- [ ] T036 Run `CGO_ENABLED=1 go test ./...` and verify all tests pass (existing core tests + new complexity tests)
- [ ] T037 Run quickstart.md verification: start server, execute all curl commands, verify responses match contracts/api.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 completion — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 — core complexity calculations
- **US2 (Phase 4)**: Depends on Phase 3 (needs []FileComplexity output to aggregate)
- **US3 (Phase 5)**: Depends on Phase 4 (needs full ComplexityResult for caching/serving)
- **US4 (Phase 6)**: Depends on Phase 3 (parallelizes the sequential parsing from US1)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational only — core per-file analysis
- **US2 (P1)**: Depends on US1 — aggregates US1's per-file output
- **US3 (P2)**: Depends on US2 — needs full ComplexityResult to cache and serve
- **US4 (P2)**: Depends on US1 — parallelizes the parsing; can be done in parallel with US3

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Language registry before parser
- Cyclomatic/cognitive calculators before parser integration
- Parser before aggregator
- Aggregator before handler/worker
- Core implementation before integration

### Parallel Opportunities

Within Phase 1:
- T003 and T004 can run in parallel (different files)

Within Phase 2:
- T007 and T008 can run in parallel (different test files)

Within Phase 3 (US1):
- T009 and T010 can run in parallel (different test files)

Within Phase 5 (US3):
- T018 and T019 can run in parallel (different test files)
- T020 and T021 can run in parallel (different source files)

US3 and US4 can proceed in parallel after US2 is complete.

---

## Parallel Example: User Story 1

```
# Launch test files in parallel:
T009: "Cyclomatic tests in internal/analysis/complexity/cyclomatic_test.go"
T010: "Cognitive tests in internal/analysis/complexity/cognitive_test.go"

# Then implement sequentially:
T011: "Cyclomatic calculator in internal/analysis/complexity/cyclomatic.go"
T012: "Cognitive calculator in internal/analysis/complexity/cognitive.go"
T013: "Integrate into parser.go"
T014: "Integration test with fixture repo"
```

---

## Implementation Strategy

### MVP First (User Story 1 + 2)

1. Complete Phase 1: Setup (T001–T004)
2. Complete Phase 2: Foundational (T005–T008)
3. Complete Phase 3: US1 — Per-File Complexity (T009–T014)
4. Complete Phase 4: US2 — Aggregation (T015–T017)
5. **STOP and VALIDATE**: Run aggregator on fixture repo, verify all metrics correct

### Incremental Delivery

1. Setup + Foundational → parsing infrastructure ready
2. US1 → per-file complexity works → validates tree-sitter integration
3. US2 → aggregation works → hot files, stats, histogram verified
4. US3 → full async pipeline → endpoint + auto-trigger + caching
5. US4 → parallel parsing → performance optimization
6. Polish → edge cases, Docker, quickstart verification

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All tests use table-driven style per constitution Principle VIII
- External deps (Redis, NATS) mocked in unit tests; integration tests use real services via Docker Compose
- CGO_ENABLED=1 required for all go build/test commands (tree-sitter cgo dependency)
