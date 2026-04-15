# Tasks: Temporal Intelligence

**Input**: Design documents from `/specs/005-temporal-intelligence/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are included per project convention (test discipline principle VIII).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create model types and foundational infrastructure shared across all user stories

- [ ] T001 Create temporal model types (MonthlySnapshot, WeeklyActivity, AuthorWeeklyActivity, CommitCadence, PRMergeTime, VelocityMetrics, RefactorPeriod, TemporalResult, TemporalSummary) in internal/models/temporal.go
- [ ] T002 [P] Add TemporalSubject constant (`grit.jobs.temporal`) to internal/job/publisher.go
- [ ] T003 [P] Add TemporalAnalysisTTL (12h) and temporal cache key helpers to internal/cache/redis.go
- [ ] T004 [P] Add temporal metrics (TemporalAnalysisDuration histogram, TemporalJobsCompletedTotal counter) to internal/metrics/prometheus.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Cache methods, publisher method, and NATS plumbing that all user stories depend on

- [ ] T005 Add temporal cache methods (GetTemporal, SetTemporal, DeleteTemporal, GetActiveTemporalJob, SetActiveTemporalJob, DeleteActiveTemporalJob) to internal/cache/redis.go
- [ ] T006 Add PublishTemporal method with deduplication to internal/job/publisher.go

**Checkpoint**: Foundation ready — user story implementation can now begin

---

## Phase 3: User Story 1 — LOC Over Time (Priority: P1) MVP

**Goal**: Monthly LOC snapshots going back up to 3 years, with top-10 language breakdown per checkpoint

**Independent Test**: Request `GET /api/:owner/:repo/temporal` and verify `loc_over_time` array with date, total_loc, and by_language per entry

### Tests for User Story 1

- [ ] T007 [P] [US1] Tests for monthly boundary resolution and LOC counting in internal/analysis/temporal/loc_test.go

### Implementation for User Story 1

- [ ] T008 [US1] Implement monthly boundary commit resolution (find latest commit on or before 1st of each month) in internal/analysis/temporal/loc.go
- [ ] T009 [US1] Implement in-memory tree walk LOC counting with top-10 language breakdown in internal/analysis/temporal/loc.go

**Checkpoint**: LOC over time logic is complete and unit-tested

---

## Phase 4: User Story 2 — Velocity Metrics (Priority: P1)

**Goal**: Weekly additions/deletions, per-author activity, rolling commit cadence, and PR merge time

**Independent Test**: Verify `velocity` object contains weekly_activity, author_activity, commit_cadence, and pr_merge_time

### Tests for User Story 2

- [ ] T010 [P] [US2] Tests for weekly velocity aggregation (weekly_activity, author_activity, commit_cadence) in internal/analysis/temporal/velocity_test.go
- [ ] T011 [P] [US2] Tests for GitHub GraphQL PR merge time client in internal/github/graphql_test.go

### Implementation for User Story 2

- [ ] T012 [US2] Implement weekly velocity walk: aggregate commit Stats() into ISO-week buckets with per-author tracking in internal/analysis/temporal/velocity.go
- [ ] T013 [US2] Implement rolling 4-week commit cadence computation in internal/analysis/temporal/velocity.go
- [ ] T014 [US2] Implement GitHub GraphQL client for PR merge time (raw net/http POST, cursor pagination, percentile calc) in internal/github/graphql.go

**Checkpoint**: Velocity metrics logic is complete and unit-tested

---

## Phase 5: User Story 3 — Refactor Detection (Priority: P2)

**Goal**: Identify consecutive weeks with negative net LOC and above-median commits, group into refactor periods

**Independent Test**: Verify `refactor_periods` array with start, end, net_loc_change, weeks

### Tests for User Story 3

- [ ] T015 [P] [US3] Tests for refactor detection (single period, split periods, no refactor, below-median exclusion) in internal/analysis/temporal/refactor_test.go

### Implementation for User Story 3

- [ ] T016 [US3] Implement refactor period detection from weekly velocity data in internal/analysis/temporal/refactor.go

**Checkpoint**: Refactor detection logic is complete and unit-tested

---

## Phase 6: User Story 4 — Async Temporal Job Pipeline (Priority: P1)

**Goal**: End-to-end async pipeline: core worker triggers temporal job, temporal worker processes analysis, handler serves cached results

**Independent Test**: Trigger core analysis, verify temporal job enqueued, poll endpoint from 202 → 200

### Tests for User Story 4

- [ ] T017 [P] [US4] Tests for Analyze orchestrator in internal/analysis/temporal/analyzer_test.go
- [ ] T018 [P] [US4] Tests for TemporalHandler (200/202/404/400 responses) in internal/handler/temporal_test.go

### Implementation for User Story 4

- [ ] T019 [US4] Implement Analyze orchestrator (coordinate LOC + velocity + refactor + PR merge time) in internal/analysis/temporal/analyzer.go
- [ ] T020 [US4] Implement TemporalWorker (NATS consumer, open repo, call analyzer, cache result) in internal/job/temporal_worker.go
- [ ] T021 [US4] Modify core worker to publish temporal job alongside blame in internal/job/worker.go
- [ ] T022 [US4] Implement TemporalHandler (GET /api/:owner/:repo/temporal with period validation) in internal/handler/temporal.go
- [ ] T023 [US4] Wire temporal worker, handler, and route in cmd/grit/main.go

**Checkpoint**: Full async pipeline works end-to-end — temporal endpoint returns data after core analysis

---

## Phase 7: User Story 5 — Temporal Summary in Main Analysis (Priority: P2)

**Goal**: Embed temporal_summary in the main GET /api/:owner/:repo response

**Independent Test**: Verify main endpoint includes temporal_summary with current_loc, loc_trend_6m_percent, avg_weekly_commits

### Implementation for User Story 5

- [ ] T024 [US5] Implement embedTemporalSummary and call it in HandleAnalysis in internal/handler/analysis.go

**Checkpoint**: Main analysis response includes temporal_summary

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Cache cleanup, full test suite validation, manual verification

- [ ] T025 Update HandleDeleteCache to clear temporal cache entries in internal/handler/cache.go
- [ ] T026 Verify all existing tests pass (go test ./... -count=1)
- [ ] T027 Quickstart manual validation per specs/005-temporal-intelligence/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on T001 (model types) — BLOCKS all user stories
- **US1 LOC (Phase 3)**: Depends on Phase 2 completion
- **US2 Velocity (Phase 4)**: Depends on Phase 2 completion — can run in parallel with US1
- **US3 Refactor (Phase 5)**: Depends on US2 (uses weekly velocity data)
- **US4 Pipeline (Phase 6)**: Depends on US1, US2, US3 (orchestrator calls all three)
- **US5 Summary (Phase 7)**: Depends on US4 (needs cached temporal data)
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (LOC Over Time)**: Independent after Phase 2
- **US2 (Velocity Metrics)**: Independent after Phase 2 — can run in parallel with US1
- **US3 (Refactor Detection)**: Depends on US2 (uses WeeklyActivity data)
- **US4 (Async Pipeline)**: Depends on US1 + US2 + US3 (orchestrator invokes all)
- **US5 (Temporal Summary)**: Depends on US4 (needs pipeline operational)

### Parallel Opportunities

**Within Phase 1** (all [P] tasks):
- T002, T003, T004 can all run in parallel after T001

**Within Phase 3+4** (US1 and US2 in parallel):
- T007 + T010 + T011 (all test tasks)
- T008/T009 (US1 LOC) parallel with T012/T013/T014 (US2 velocity)

**Within Phase 6** (test tasks):
- T017 + T018 can run in parallel

---

## Parallel Example: US1 + US2 Simultaneously

```text
# After Phase 2 completes, launch US1 and US2 tests in parallel:
T007: Tests for LOC counting in loc_test.go
T010: Tests for velocity aggregation in velocity_test.go
T011: Tests for GraphQL PR merge time in graphql_test.go

# Then implement US1 and US2 in parallel:
T008+T009: LOC boundary resolution + tree walk counting
T012+T013+T014: Velocity walk + cadence + GraphQL client
```

---

## Implementation Strategy

### MVP First (US1 + US4 Only)

1. Complete Phase 1: Setup (model types + constants)
2. Complete Phase 2: Foundational (cache + publisher)
3. Complete Phase 3: US1 — LOC Over Time
4. Stub velocity and refactor in analyzer, complete Phase 6: US4 — Pipeline
5. **STOP and VALIDATE**: Temporal endpoint returns LOC data
6. Incrementally add US2 (velocity), US3 (refactor), US5 (summary)

### Incremental Delivery

1. Setup + Foundational → Infrastructure ready
2. US1 (LOC) → Core temporal metric available
3. US2 (Velocity) → Weekly activity, author stats, PR merge time
4. US3 (Refactor) → Refactor period detection
5. US4 (Pipeline) → Full async end-to-end
6. US5 (Summary) → Embedded in main response
7. Polish → Cache cleanup, validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- The GitHub GraphQL client (T014) has no dependency on any temporal-specific code — only on the existing github.Client pattern
- Period parameter handling is in the handler (T022) and passed to the analyzer (T019)
- Refactor detection (T016) takes WeeklyActivity as input, making it a pure function testable in isolation
