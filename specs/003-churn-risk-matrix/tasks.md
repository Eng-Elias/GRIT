# Tasks: Git Churn Analysis, Risk Matrix & Dead Code Estimation

**Input**: Design documents from `/specs/003-churn-risk-matrix/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.md

**Tests**: Included — the feature spec and plan.md require table-driven tests per constitution Principle VIII.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Create package structure and shared model types

- [ ] T001 Create package directory `internal/analysis/churn/`
- [ ] T002 [P] Create churn model types (FileChurn, RiskEntry, StaleFile, Thresholds, ChurnMatrixResult, ChurnSummary) in `internal/models/churn.go` per data-model.md
- [ ] T003 [P] Add `grit_churn_analysis_duration_seconds` Prometheus histogram metric in `internal/metrics/prometheus.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Churn cache methods and NATS publisher — required by ALL user stories

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 Add churn cache methods to `internal/cache/redis.go`: `churnKey(owner, repo, sha)` returning `{owner}/{repo}:{sha}:churn`; `GetChurn(ctx, owner, repo, sha)` with SCAN fallback for empty SHA; `SetChurn(ctx, owner, repo, sha, data)` with 24h TTL; `DeleteChurn(ctx, owner, repo)` via SCAN pattern delete; `GetActiveChurnJob(ctx, owner, repo, sha)`; `SetActiveChurnJob(ctx, owner, repo, sha, jobID)`; `DeleteActiveChurnJob(ctx, owner, repo, sha)`
- [ ] T005 Add NATS churn subject constant `ChurnSubject = "grit.jobs.churn"` and `PublishChurn` method in `internal/job/publisher.go` — same pattern as `PublishComplexity`: deduplication via `GetActiveChurnJob`, create job payload, publish to `grit.jobs.churn`, store job state and active key

**Checkpoint**: Foundation ready — churn cache and job publishing available for all user stories

---

## Phase 3: User Story 1 — Per-File Churn Scores (Priority: P1) 🎯 MVP

**Goal**: Walk default branch commit log and produce per-file churn counts with commit window caps (2 years / 5,000 commits)

**Independent Test**: Verify `churn` array in response has correct per-file modification counts; verify commit window caps are respected

### Tests for User Story 1

- [ ] T006 [P] [US1] Write table-driven churn counting tests in `internal/analysis/churn/churn_test.go`: test counting commits per file from a go-git repository; test 5,000-commit cap stops iteration; test 2-year date cap stops iteration; test empty repository returns empty map; test merge commits count file modifications; test renamed files tracked independently

### Implementation for User Story 1

- [ ] T007 [US1] Implement commit log walker in `internal/analysis/churn/churn.go`: accept `*git.Repository`, walk `repo.Log()` from HEAD using `CommitIterator`, build `map[string]int` for churn counts and `map[string]time.Time` for last-modified dates; enforce 2-year and 5,000-commit caps (whichever is smaller); return `[]models.FileChurn` sorted by churn descending; also return `totalCommits`, `windowStart`, `windowEnd` metadata
- [ ] T008 [P] [US1] Implement sort-based percentile calculator in `internal/analysis/churn/percentile.go`: accept `[]float64`, return p50/p75/p90 using sorted-index lookup; handle edge cases (empty slice, single element, fewer than 4 elements)
- [ ] T009 [P] [US1] Write table-driven percentile tests in `internal/analysis/churn/percentile_test.go`: test known percentile values; test empty input; test single element; test degenerate cases (all same values)

**Checkpoint**: Churn counting and percentile calculation work in isolation

---

## Phase 4: User Story 2 — Risk Matrix (Priority: P1)

**Goal**: Join churn data with complexity data from Redis, classify files into risk levels using percentile thresholds, produce risk zone of critical files

**Independent Test**: Seed known churn + complexity values, verify risk classifications match expected levels, verify risk_zone sorted by churn × complexity descending, verify thresholds contain all 6 percentile values

### Tests for User Story 2

- [ ] T010 [P] [US2] Write table-driven risk classification tests in `internal/analysis/churn/risk_test.go`: test critical (both churn and complexity > p75); test high (either > p90); test medium (either > p50); test low (both below p50); test risk_zone contains only critical entries sorted by churn×complexity desc; test empty inputs; test single file; test thresholds struct is correct

### Implementation for User Story 2

- [ ] T011 [US2] Implement risk matrix builder in `internal/analysis/churn/risk.go`: accept `[]models.FileChurn` and complexity `[]models.FileComplexity`, join on file path, compute churn and complexity percentile thresholds, classify each joined file into risk level per FR-006, build `risk_zone` (critical files sorted by churn×complexity desc), return `[]models.RiskEntry`, `[]models.RiskEntry` (risk_zone), and `models.Thresholds`

**Checkpoint**: Risk classification works correctly given churn + complexity inputs

---

## Phase 5: User Story 3 — Dead Code Estimation (Priority: P2)

**Goal**: Detect potentially unused source files using 3-condition heuristic: no recent commits + source extension + no import references

**Independent Test**: Fixture repo with a stale unreferenced source file → appears in `stale_files`; config/doc files excluded; referenced files excluded

### Tests for User Story 3

- [ ] T012 [P] [US3] Write table-driven stale file detection tests in `internal/analysis/churn/stale_test.go`: test source file with no recent commits and no references → flagged stale; test source file with no recent commits but referenced → NOT stale; test non-source file (.md, .yaml) → NOT stale; test file with recent commits → NOT stale; test months_inactive calculation; test empty candidate list

### Implementation for User Story 3

- [ ] T013 [US3] Implement stale file detector in `internal/analysis/churn/stale.go`: define source extension allowlist and non-source excludelist; accept `[]models.FileChurn` (with last_modified dates), clone dir path, and list of all files at HEAD; filter candidates: zero commits in past 6 months + source extension; for each candidate, scan all source files for `strings.Contains(content, baseName)` — if not found, mark as stale; return `[]models.StaleFile` with path, last_modified, months_inactive

**Checkpoint**: Stale file detection works in isolation with correct heuristic filtering

---

## Phase 6: User Story 4 — Async Churn Job Pipeline (Priority: P1)

**Goal**: Wire churn analysis into async job pipeline — auto-trigger from complexity worker, requeue if complexity not ready, cache results, serve via HTTP endpoint

**Independent Test**: Trigger core analysis → complexity completes → churn job auto-enqueued → endpoint transitions 202→200; verify cache hit on second request

### Tests for User Story 4

- [ ] T014 [P] [US4] Write churn handler tests in `internal/handler/churn_test.go`: test GET returns 200 with cached data and X-Cache HIT; test returns 404 when no analysis exists; test returns 202 when churn job in progress; test returns 400 for invalid owner/repo params
- [ ] T015 [P] [US4] Write churn worker tests (append to `internal/job/worker_test.go`): test churn worker updates job status to running; test churn worker completion stores result and marks job completed; test churn worker failure marks job failed; test churn worker requeues when complexity data not available; test PublishChurn deduplication

### Implementation for User Story 4

- [ ] T016 [US4] Implement ChurnAnalyzer orchestrator in `internal/analysis/churn/analyzer.go`: accept clone dir, `*git.Repository`, `[]models.FileComplexity` (from complexity cache), and `models.Repository`; call churn walker (T007), risk matrix builder (T011), stale detector (T013); assemble `ChurnMatrixResult`; handle context cancellation (return partial results)
- [ ] T017 [P] [US4] Write analyzer integration test in `internal/analysis/churn/analyzer_test.go`: test full analysis on fixture repo produces valid ChurnMatrixResult with churn, risk_matrix, stale_files populated; test empty repo; test repo with no supported files
- [ ] T018 [US4] Implement ChurnWorker in `internal/job/churn_worker.go`: subscribe to `grit.jobs.churn` with durable consumer `grit-churn-worker`; processMessage: unmarshal payload, update job to running, check complexity cache exists (if not, `msg.NakWithDelay(30s)` and return), open cloned repo with go-git, run ChurnAnalyzer, cache result via `SetChurn`, update job to completed; on error mark job failed and ack
- [ ] T019 [US4] Modify complexity worker in `internal/job/complexity_worker.go`: after successful complexity completion (after cache set), call `publisher.PublishChurn(ctx, owner, repo, sha, token)` to auto-trigger churn job — same pattern as core worker triggers complexity
- [ ] T020 [US4] Implement churn handler in `internal/handler/churn.go`: GET /api/:owner/:repo/churn-matrix; validate owner/repo with `ownerRepoRegex`; check churn cache → 200 with X-Cache HIT; check active churn job → 202 with job status; otherwise 404
- [ ] T021 [US4] Modify analysis handler in `internal/handler/analysis.go`: add `embedChurnSummary` method (same pattern as `embedComplexitySummary`); look up churn cache, build ChurnSummary with status/total_files/critical_count/stale_count/url, merge into JSON response
- [ ] T022 [US4] Extend cache deletion in `internal/handler/cache.go`: add `DeleteChurn` call alongside existing `DeleteComplexity` call so cache bust clears all three pillars
- [ ] T023 [US4] Register churn worker and handler in `cmd/grit/main.go`: create ChurnAnalyzer, create ChurnWorker with `Start(ctx)`, pass publisher to complexity worker for churn auto-trigger, add route `r.Get("/churn-matrix", churnHandler.HandleChurnMatrix)` inside existing `/api/{owner}/{repo}` group

**Checkpoint**: Full async pipeline works — complexity completion triggers churn job, results cached and served via API

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases, verification, and final quality pass

- [ ] T024 [P] Add edge case handling in analyzer: zero files → valid empty ChurnMatrixResult; context timeout → return partial results with logged warning
- [ ] T025 Run `go test ./...` and verify all tests pass (existing core + complexity + new churn tests)
- [ ] T026 Run quickstart.md verification: start server, execute all curl commands from quickstart.md, verify responses match contracts/api.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (models must exist) — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 — churn counting is standalone
- **US2 (Phase 4)**: Depends on Phase 2 — risk matrix depends on US1 churn output at runtime but code is independently implementable
- **US3 (Phase 5)**: Depends on Phase 2 — stale detection depends on US1 churn data at runtime but code is independently implementable
- **US4 (Phase 6)**: Depends on Phases 3, 4, 5 — the orchestrator wires US1+US2+US3 together and adds the async pipeline
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Can start after Phase 2 — no story dependencies
- **US2 (P1)**: Can start after Phase 2 — code is independent; uses US1 output at runtime
- **US3 (P2)**: Can start after Phase 2 — code is independent; uses US1 output at runtime
- **US4 (P1)**: Depends on US1 + US2 + US3 implementation — wires everything together

### Within Each User Story

- Tests written FIRST, verified to fail before implementation
- Core logic before integration
- Story complete before moving to next priority

### Parallel Opportunities

- T002 + T003: Model types and metrics can be created in parallel
- T006 + T008 + T009: Churn tests, percentile impl, and percentile tests (different files)
- T010 + T012: Risk tests and stale tests (different files, no dependencies)
- T014 + T015: Handler tests and worker tests (different files)
- T016 + T017: Analyzer impl and analyzer test can be co-developed

---

## Parallel Example: User Story 1

```bash
# Launch tests and implementation in parallel:
Task T006: "Write churn counting tests in internal/analysis/churn/churn_test.go"
Task T008: "Implement percentile calculator in internal/analysis/churn/percentile.go"
Task T009: "Write percentile tests in internal/analysis/churn/percentile_test.go"

# Then sequentially:
Task T007: "Implement commit log walker in internal/analysis/churn/churn.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T003)
2. Complete Phase 2: Foundational (T004–T005)
3. Complete Phase 3: User Story 1 (T006–T009)
4. **STOP and VALIDATE**: Churn counting works with correct caps and percentiles
5. Can demo churn scores even without risk matrix

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 (churn scores) → Test independently → MVP churn data
3. Add US2 (risk matrix) → Test independently → Risk classification added
4. Add US3 (dead code) → Test independently → Stale file detection added
5. Add US4 (async pipeline) → Full integration → Endpoint live
6. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- No new dependencies needed — uses existing go-git, go-redis, nats.go, chi
- Complexity worker modification (T019) follows same auto-trigger pattern used in core → complexity
- Churn worker requeue (T018) uses `msg.NakWithDelay(30 * time.Second)` for complexity-not-ready case
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
