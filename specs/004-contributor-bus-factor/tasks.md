# Tasks: Contributor Attribution & Bus Factor Analysis

**Input**: Design documents from `/specs/004-contributor-bus-factor/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Tests are included — the spec explicitly calls for table-driven unit tests (Constitution Principle VIII).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Go project at repository root: `internal/`, `cmd/`
- Tests co-located with source files (`_test.go` suffix)

---

## Phase 1: Setup

**Purpose**: Create the blame analysis package skeleton and shared model types

- [ ] T001 Create contributor model types (Author, FileAuthor, FileContributors, ContributorResult, ContributorSummary, AuthorBrief) in `internal/models/contributor.go`
- [ ] T002 [P] Create source file extension registry in `internal/analysis/blame/extensions.go` — define `SupportedSourceExtensions` map (`.go`, `.py`, `.js`, `.ts`, `.java`, `.rb`, `.rs`, `.c`, `.cpp`, `.jsx`, `.tsx`) and `IsSourceFile(path) bool` helper
- [ ] T003 [P] Add `BlameSubject` constant (`grit.jobs.blame`) to `internal/job/publisher.go`
- [ ] T004 [P] Add `ContributorAnalysisTTL = 48 * time.Hour` constant to `internal/cache/redis.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Redis cache methods and publisher method that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T005 Add contributor cache key helpers (`contributorKey`, `activeBlameKey`) and cache methods (`GetContributors`, `findContributors`, `SetContributors`, `DeleteContributors`, `GetActiveBlameJob`, `SetActiveBlameJob`, `DeleteActiveBlameJob`) to `internal/cache/redis.go` — follow existing churn cache pattern
- [ ] T006 [P] Add `PublishBlame(ctx, owner, repo, sha, token) (string, error)` method to `internal/job/publisher.go` — follow existing `PublishComplexity`/`PublishChurn` pattern with deduplication via `GetActiveBlameJob`
- [ ] T007 [P] Add blame metrics (`BlameAnalysisDuration` histogram, `BlameJobsCompletedTotal` counter) to `internal/metrics/prometheus.go`

**Checkpoint**: Foundation ready — user story implementation can now begin

---

## Phase 3: User Story 1 — Per-Author Contributor Statistics (Priority: P1) 🎯 MVP

**Goal**: Run git blame on every source file, attribute lines to authors by email (case-insensitive dedup), compute per-author stats (lines owned, ownership %, files touched, first/last commit dates, primary languages, active flag)

**Independent Test**: `go test ./internal/analysis/blame/... -v -run TestBlame` — verify blame output has correct author count, line counts, email dedup, language detection, active flag

### Tests for User Story 1

- [ ] T008 [P] [US1] Write table-driven tests for `BlameFile` in `internal/analysis/blame/blame_test.go` — test single-author file, multi-author file, empty file skip, binary file skip
- [ ] T009 [P] [US1] Write table-driven tests for `Aggregate` in `internal/analysis/blame/aggregator_test.go` — test email case dedup, most-recent-name resolution, ownership % summing to 100%, primary languages top-3, `is_active` flag (6-month threshold), files_touched count

### Implementation for User Story 1

- [ ] T010 [US1] Implement `BlameFile(ctx, repo *git.Repository, commitHash plumbing.Hash, path string) (*FileBlameResult, error)` in `internal/analysis/blame/blame.go` — wraps `git.Blame()`, strips angle brackets from email, returns per-line author/email/date/language
- [ ] T011 [US1] Implement `Aggregate(fileResults map[string]*FileBlameResult) *AggregateResult` in `internal/analysis/blame/aggregator.go` — case-insensitive email dedup into `map[string]*authorAccumulator`, resolve display name from most-recent commit date, compute `total_lines_owned`, `ownership_percent`, `files_touched`, `first_commit_date`, `last_commit_date`, `primary_languages` (top 3 by LOC), `is_active` (last commit within 6 months), sort authors by `total_lines_owned` desc

**Checkpoint**: Per-author stats are computed correctly from blame data. Verifiable with unit tests.

---

## Phase 4: User Story 2 — Bus Factor Calculation (Priority: P1)

**Goal**: Compute bus factor (minimum authors for 80% ownership) and identify key people

**Independent Test**: `go test ./internal/analysis/blame/... -v -run TestBusFactor` — verify bus factor = 1 when single author owns 85%, bus factor = 3 for 30/28/25 split, bus factor = 0 for empty repo

### Tests for User Story 2

- [ ] T012 [P] [US2] Write table-driven tests for `ComputeBusFactor` in `internal/analysis/blame/aggregator_test.go` — test single dominant author (85%), three-way split (30/28/25 = bus factor 3), equal distribution (10 × 10% = bus factor 8), empty repo (bus factor 0), single author (bus factor 1)

### Implementation for User Story 2

- [ ] T013 [US2] Implement `ComputeBusFactor(authors []Author) (int, []Author)` in `internal/analysis/blame/aggregator.go` — walk sorted authors accumulating ownership until >80%, return count and key_people slice. Return (0, nil) for empty input.

**Checkpoint**: Bus factor computed correctly. Verifiable with unit tests.

---

## Phase 5: User Story 3 — Per-File Contributor Breakdown (Priority: P2)

**Goal**: For each source file, return top 3 authors by line ownership with per-file ownership percentages

**Independent Test**: `go test ./internal/analysis/blame/... -v -run TestFileContributors` — verify top-3 slicing, single-author = 100%, 5 authors truncated to 3

### Tests for User Story 3

- [ ] T014 [P] [US3] Write table-driven tests for `ComputeFileContributors` in `internal/analysis/blame/aggregator_test.go` — test 5-author file (top 3 only), single author (100%), 2-author file (both returned), ownership_percent relative to file total

### Implementation for User Story 3

- [ ] T015 [US3] Implement `ComputeFileContributors(fileResults map[string]*FileBlameResult) []FileContributors` in `internal/analysis/blame/aggregator.go` — for each file, count lines per author, sort desc, take top 3, compute per-file ownership_percent

**Checkpoint**: Per-file breakdown works correctly. Verifiable with unit tests.

---

## Phase 6: User Story 4 — Async Blame Job Pipeline (Priority: P1)

**Goal**: Goroutine pool orchestrator, blame worker (NATS consumer), core worker trigger, contributor handler endpoints — the full async pipeline

**Independent Test**: Trigger core analysis → verify blame job auto-enqueued → poll `/contributors` transitions from 202 → 200 → verify cache HIT on re-request

### Tests for User Story 4

- [ ] T016 [P] [US4] Write table-driven tests for `Analyze` in `internal/analysis/blame/analyzer_test.go` — test pool orchestration with mock repo, partial results on context cancellation, all-files-complete case
- [ ] T017 [P] [US4] Write handler tests for `HandleContributors` and `HandleContributorFiles` in `internal/handler/contributors_test.go` — test cache hit (200), active job (202), no analysis (404), bad request (400)

### Implementation for User Story 4

- [ ] T018 [US4] Implement `Analyze(ctx context.Context, repo *git.Repository, commitHash plumbing.Hash, files []string) (*ContributorResult, error)` in `internal/analysis/blame/analyzer.go` — create `context.WithTimeout(ctx, 10*time.Minute)`, spin up `min(4, NumCPU())` goroutines reading from buffered file channel, each calls `BlameFile`, stores result in `sync.Mutex`-protected map, after `wg.Wait()` call `Aggregate` + `ComputeBusFactor` + `ComputeFileContributors`, set `Partial: ctx.Err() != nil`
- [ ] T019 [US4] Implement `BlameWorker` in `internal/job/blame_worker.go` — NATS pull subscriber on `grit.jobs.blame`, opens cloned repo, filters source files via `IsSourceFile`, calls `Analyze`, marshals `ContributorResult`, calls `cache.SetContributors`, updates job status/completion, deletes active blame job key, records `BlameAnalysisDuration` metric
- [ ] T020 [US4] Modify core `Worker.processMessage` in `internal/job/worker.go` — after core analysis completes and before `PublishComplexity`, add `publisher.PublishBlame(ctx, owner, repo, sha, token)` call so blame is enqueued alongside complexity
- [ ] T021 [US4] Implement `ContributorHandler` with `HandleContributors` and `HandleContributorFiles` in `internal/handler/contributors.go` — follow churn handler pattern: check `cache.GetContributors` → 200 HIT; check `cache.GetActiveBlameJob` → 202; check core exists → 404 with message; `/contributors/files` extracts `file_contributors` from same cached result
- [ ] T022 [US4] Register contributor routes and wire blame worker in `cmd/grit/main.go` — add `r.Get("/contributors", contributorHandler.HandleContributors)` and `r.Get("/contributors/files", contributorHandler.HandleContributorFiles)` inside the `/api/{owner}/{repo}` route group; instantiate `NewContributorHandler`; start `BlameWorker` goroutine alongside complexity and churn workers

**Checkpoint**: Full async pipeline operational — core triggers blame, worker processes, endpoints serve cached results.

---

## Phase 7: User Story 5 — Contributor Summary in Main Analysis (Priority: P2)

**Goal**: Embed `contributor_summary` in the main `GET /api/:owner/:repo` response

**Independent Test**: `curl /api/octocat/hello-world | jq .contributor_summary` — verify `bus_factor`, `top_contributors`, `total_authors`, `active_authors`, `contributors_url` fields present

### Implementation for User Story 5

- [ ] T023 [US5] Implement `embedContributorSummary` method on `AnalysisHandler` in `internal/handler/analysis.go` — follow existing `embedChurnSummary` pattern: check `cache.GetContributors`, unmarshal, populate `ContributorSummary` with `status`, `bus_factor`, `top_contributors` (top 3 AuthorBrief), `total_authors`, `active_authors`, `contributors_url`; merge into raw JSON
- [ ] T024 [US5] Call `embedContributorSummary` in `HandleAnalysis` in `internal/handler/analysis.go` — add `cachedData = h.embedContributorSummary(r.Context(), owner, repo, cachedData)` after `embedChurnSummary` call

**Checkpoint**: Main analysis response includes contributor summary. Verifiable via curl.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Cache deletion, regression safety, documentation

- [ ] T025 Update `HandleDeleteCache` in `internal/handler/cache.go` — add `cache.DeleteContributors(ctx, owner, repo)` call to clear contributor cache on DELETE
- [ ] T026 [P] Verify all existing tests pass with `go test ./... -v` — ensure no regressions in core, complexity, churn pipelines
- [ ] T027 [P] Run quickstart.md manual validation — start stack, trigger analysis, verify `/contributors` and `/contributors/files` endpoints return expected data

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (T001 for model types, T004 for TTL constant)
- **US1 (Phase 3)**: Depends on Phase 2 complete (T005 not needed yet, but T002 extensions needed)
- **US2 (Phase 4)**: Depends on US1 (T011 `Aggregate` produces the `[]Author` input)
- **US3 (Phase 5)**: Depends on US1 (T010 `BlameFile` produces `FileBlameResult` map)
- **US4 (Phase 6)**: Depends on US1 + US2 + US3 (T018 orchestrator calls all three) AND Phase 2 (T005 cache, T006 publisher)
- **US5 (Phase 7)**: Depends on US4 (needs cached contributor data to embed)
- **Polish (Phase 8)**: Depends on all user stories complete

### User Story Dependencies

```
Phase 1 (Setup)
  │
  ▼
Phase 2 (Foundational)
  │
  ├──► Phase 3 (US1: Per-Author Stats)
  │      │
  │      ├──► Phase 4 (US2: Bus Factor) ──┐
  │      │                                 │
  │      └──► Phase 5 (US3: Per-File) ─────┤
  │                                        │
  │                                        ▼
  └──────────────────────────────────► Phase 6 (US4: Pipeline)
                                           │
                                           ▼
                                       Phase 7 (US5: Summary)
                                           │
                                           ▼
                                       Phase 8 (Polish)
```

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Models/helpers before orchestrators
- Core logic before handlers
- Handlers before route wiring

### Parallel Opportunities

**Within Phase 1**: T002, T003, T004 are all independent files — run in parallel.

**Within Phase 2**: T005, T006, T007 are independent files — run in parallel (after T001, T004).

**Within Phase 3**: T008, T009 (tests) in parallel, then T010, T011 (implementation) sequentially.

**Within Phase 4**: T012 (test) then T013 (implementation).

**Within Phase 5**: T014 (test) then T015 (implementation).

**Within Phase 6**: T016, T017 (tests) in parallel, then T018 → T019 → T020 → T021 → T022 sequentially.

**Within Phase 7**: T023 → T024 sequentially.

**Within Phase 8**: T025, T026, T027 all in parallel.

---

## Parallel Example: User Story 1

```text
# Tests in parallel:
Task T008: "Write table-driven tests for BlameFile in internal/analysis/blame/blame_test.go"
Task T009: "Write table-driven tests for Aggregate in internal/analysis/blame/aggregator_test.go"

# Then implementation sequentially:
Task T010: "Implement BlameFile in internal/analysis/blame/blame.go"
Task T011: "Implement Aggregate in internal/analysis/blame/aggregator.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (T001–T004)
2. Complete Phase 2: Foundational (T005–T007)
3. Complete Phase 3: US1 — Per-Author Stats (T008–T011)
4. Complete Phase 4: US2 — Bus Factor (T012–T013)
5. **STOP and VALIDATE**: Run `go test ./internal/analysis/blame/... -v` — all unit tests pass, bus factor correctly computed

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 (per-author stats) → Test independently → Core blame logic works
3. Add US2 (bus factor) → Test independently → Bus factor computed
4. Add US3 (per-file breakdown) → Test independently → File-level data works
5. Add US4 (async pipeline) → Test independently → Full end-to-end pipeline
6. Add US5 (summary embedding) → Test independently → Main endpoint enriched
7. Polish → Regression check → Ship

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- US4 is the integration story — it wires everything together and is the largest phase
