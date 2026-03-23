# Tasks: Core Analysis Engine

**Input**: Design documents from `/specs/001-core-analysis/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are included per the constitution's Test Discipline principle (Principle VIII) — every service package MUST have unit tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go project**: `cmd/grit/` for entrypoint, `internal/` for domain packages
- **Tests**: Co-located `_test.go` files in each package
- **Fixture data**: `testdata/fixture-repo/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, Go module, Docker, and config

- [ ] T001 Initialize Go module with `go mod init github.com/your-org/grit` and add all dependencies (chi v5, go-git v5, go-redis v9, nats.go, godotenv, prometheus/client_golang, testify) in go.mod
- [ ] T002 Create project directory structure per plan.md: cmd/grit/, internal/analysis/core/, internal/cache/, internal/clone/, internal/config/, internal/github/, internal/handler/, internal/job/, internal/metrics/, internal/models/, testdata/fixture-repo/
- [ ] T003 [P] Create .env.example with all environment variables (GITHUB_TOKEN, REDIS_URL, NATS_URL, PORT, CLONE_DIR, CLONE_SIZE_THRESHOLD_KB) per quickstart.md
- [ ] T004 [P] Create Dockerfile with single-stage Go build exposing port 8080
- [ ] T005 [P] Create docker-compose.yml with grit, redis:7, and nats:latest (JetStream enabled) services with pinned image versions

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Implement config loader in internal/config/config.go — parse all env vars via godotenv + os.Getenv, return typed Config struct with fields: Port, RedisURL, NATSURL, GitHubToken, CloneDir, CloneSizeThresholdKB
- [ ] T007 [P] Implement all shared model structs in internal/models/repository.go — Repository, GitHubMetadata, CommitActivity per data-model.md
- [ ] T008 [P] Implement all shared model structs in internal/models/analysis.go — AnalysisResult, FileStats, LanguageBreakdown per data-model.md, with JSON tags matching contracts/api.md response format
- [ ] T009 [P] Implement all shared model structs in internal/models/job.go — AnalysisJob, JobStatus enum (queued/running/completed/failed), JobProgress, SubJobStatus enum (pending/running/completed/failed) per data-model.md
- [ ] T010 Implement Redis cache client in internal/cache/redis.go — Get, Set (with TTL), Delete methods; key format {owner}/{repo}:{sha}:core (24h TTL), job:{job_id} (1h TTL), active:{owner}/{repo}:{sha} (10m TTL); return typed errors for cache miss vs connection failure
- [ ] T011 [P] Implement Prometheus metrics registration in internal/metrics/prometheus.go — register counters (grit_jobs_completed_total, grit_jobs_failed_total, grit_cache_hit_total, grit_cache_miss_total, grit_github_api_requests_total with endpoint label) and histograms (grit_clone_duration_seconds, grit_analysis_duration_seconds)
- [ ] T012 Implement JSON error envelope helper in internal/handler/errors.go — WriteError(w, status, code, message) function producing {"error":"<code>","message":"<msg>"} per contracts/api.md error envelope; map error codes: bad_request(400), private_repo(403), not_found(404), rate_limited(429), service_unavailable(503), analysis_timeout(504)
- [ ] T013 Create testdata/fixture-repo/ — initialize a small Git repository with 5-10 files across 3+ languages (Go, Python, JavaScript), including .gitignore, comments, blank lines, and a binary file; commit it so go-git can clone it in tests
- [ ] T014 Write unit tests for config loader in internal/config/config_test.go — table-driven tests for all env var parsing, defaults, and missing required values
- [ ] T015 [P] Write unit tests for model JSON serialization in internal/models/analysis_test.go — verify AnalysisResult, FileStats, LanguageBreakdown marshal to JSON matching contracts/api.md response schema

**Checkpoint**: Foundation ready — user story implementation can now begin

---

## Phase 3: User Story 1 — Analyze a Public Repository (Priority: P1) 🎯 MVP

**Goal**: End-to-end flow: accept owner/repo, shallow-clone via go-git, walk file tree, count lines per language, fetch GitHub metadata + commit activity, return complete JSON analysis

**Independent Test**: `curl http://localhost:8080/api/sindresorhus/is` returns 202, then polling returns complete analysis with line counts, metadata, and commit activity

### Implementation for User Story 1

- [ ] T016 [P] [US1] Implement language extension map in internal/analysis/core/languages.go — LanguageDef struct (Name, LineCommentPrefixes, BlockCommentStart, BlockCommentEnd), map of 50+ extensions per research.md R7; export LookupLanguage(ext) function
- [ ] T017 [P] [US1] Implement line counter in internal/analysis/core/counter.go — CountLines(reader, langDef) returning (total, code, comment, blank) using state-machine approach per research.md R2; handle single-line and block comments; detect binary files via null-byte scan of first 512 bytes
- [ ] T018 [P] [US1] Implement go-git shallow clone in internal/clone/clone.go — Clone(ctx, owner, repo, token, sizeKB, cloneDir, threshold) returning *git.Repository; use memory.NewStorage() for repos <threshold, filesystem.NewStorage() to cloneDir/{owner}/{repo}/{sha} for larger; depth=1, SingleBranch=true per research.md R1
- [ ] T019 [US1] Implement file tree walker in internal/analysis/core/walker.go — Walk(repo *git.Repository) returning []FileStats; iterate worktree files, respect .gitignore patterns, skip binary files from line counting, lookup language per extension, call CountLines for each text file
- [ ] T020 [P] [US1] Implement GitHub HTTP client base in internal/github/client.go — NewClient(token, httpClient) with configurable timeout (10s); add Authorization header when token provided; read X-RateLimit-Remaining and X-RateLimit-Reset headers; return typed RateLimitError when remaining=0
- [ ] T021 [US1] Implement metadata fetcher in internal/github/metadata.go — FetchMetadata(ctx, owner, repo) returning GitHubMetadata; call GET /repos/{owner}/{repo} and GET /repos/{owner}/{repo}/community/profile; support If-None-Match/ETag conditional requests; 30s total timeout
- [ ] T022 [US1] Implement commit activity fetcher in internal/github/stats.go — FetchCommitActivity(ctx, owner, repo) returning CommitActivity; call GET /repos/{owner}/{repo}/stats/participation for weekly counts; call GET /repos/{owner}/{repo}/commits?per_page=1 and parse Link header for total count; transparent 202 retry with exponential backoff (2s,4s,8s,8s,8s) per research.md R3
- [ ] T023 [US1] Implement core analyzer orchestrator in internal/analysis/core/analyzer.go — Analyze(ctx, owner, repo, token, config) returning AnalysisResult; orchestrate: fetch metadata (get size + default branch SHA) → clone → walk → count → aggregate LanguageBreakdown → assemble AnalysisResult; record Prometheus metrics (clone duration, analysis duration)
- [ ] T024 [US1] Implement NATS JetStream publisher in internal/job/publisher.go — NewPublisher(js JetStream); create stream GRIT with subjects ["grit.jobs.>"], WorkQueuePolicy, MaxAge 1h; Publish(ctx, job) with Nats-Msg-Id={owner}/{repo}:{sha} for dedup per research.md R4; store active:{owner}/{repo}:{sha}→job_id in Redis (10m TTL)
- [ ] T025 [US1] Implement NATS JetStream worker in internal/job/worker.go — NewWorker(js, analyzer, cache, metrics); subscribe to grit.jobs.analysis as durable consumer grit-worker (AckWait 6m, MaxDeliver 3); on message: update job status in Redis to "running", call analyzer.Analyze(), store result in cache, update job status to "completed" or "failed", ack message
- [ ] T026 [US1] Implement GET /api/:owner/:repo handler in internal/handler/analysis.go — validate owner/repo format (regex: ^[a-zA-Z0-9._-]+$), extract optional Bearer token; check Redis cache → if HIT, return 200 with X-Cache:HIT and cached/cached_at fields; if MISS, check active job → if exists return 202 with existing job_id; else create new job, publish to NATS, return 202 with job_id and poll_url; increment cache_hit/miss metrics
- [ ] T027 [US1] Implement GET /api/:owner/:repo/status handler in internal/handler/status.go — lookup job:{job_id} in Redis by active:{owner}/{repo}:{sha} key; return job status with progress object per contracts/api.md; if no active job and no cache, return not_found message
- [ ] T028 [US1] Implement main.go entrypoint in cmd/grit/main.go — load config, connect to Redis, connect to NATS, create JetStream context, register chi routes (GET /api/{owner}/{repo}, GET /api/{owner}/{repo}/status, DELETE /api/{owner}/{repo}/cache, GET /api/{owner}/{repo}/badge, GET /metrics), start NATS worker in goroutine, start HTTP server on configured port, handle graceful shutdown (SIGINT/SIGTERM)
- [ ] T029 [US1] Write unit tests for language map in internal/analysis/core/languages_test.go — verify 40+ extensions resolve to correct language; verify unknown extensions return "Other"
- [ ] T030 [US1] Write unit tests for line counter in internal/analysis/core/counter_test.go — table-driven tests: Go file with // and /* */ comments, Python with # and """ """, blank lines, binary file detection
- [ ] T031 [US1] Write unit tests for file walker in internal/analysis/core/walker_test.go — use testdata/fixture-repo; verify .gitignore exclusion, correct per-file counts, binary file skipping
- [ ] T032 [US1] Write unit tests for clone in internal/clone/clone_test.go — test memory clone for small repos, disk clone for large repos (mock size threshold); verify depth=1 shallow clone
- [ ] T033 [US1] Write unit tests for GitHub client in internal/github/client_test.go — mock HTTP server; test auth header injection, rate limit detection, ETag conditional requests
- [ ] T034 [US1] Write unit tests for stats fetcher in internal/github/stats_test.go — mock HTTP server; test 202 retry with backoff (verify 5 retries), successful response parsing, Link header total commit parsing

**Checkpoint**: User Story 1 complete — full analysis flow works end-to-end

---

## Phase 4: User Story 2 — Async Job Status Polling (Priority: P1)

**Goal**: Client can poll GET /api/:owner/:repo/status to track sub-job progress (clone, file-walk, metadata-fetch, commit-activity-fetch) until complete or failed

**Independent Test**: Request analysis of uncached repo → poll status → see sub-job progress transitions → eventually see "completed"

### Implementation for User Story 2

- [ ] T035 [US2] Enhance worker in internal/job/worker.go — update JobProgress sub-job statuses in Redis as each step starts and completes (clone→running→completed, file_walk→running→completed, etc.); on failure, mark failed sub-job and set overall status to "failed" with error message
- [ ] T036 [US2] Enhance status handler in internal/handler/status.go — return full JobProgress object in response per contracts/api.md; include result_url on completion; include error field on failure
- [ ] T037 [US2] Write unit tests for job progress tracking in internal/job/worker_test.go — mock analyzer with step callbacks; verify Redis job state transitions: queued→running(clone)→running(file_walk)→completed; verify failed sub-job propagates to overall failed status

**Checkpoint**: User Stories 1 AND 2 complete — async polling flow works independently

---

## Phase 5: User Story 3 — Authenticated Access (Priority: P2)

**Goal**: Optional Bearer token flows through to GitHub API calls for higher rate limits; private repos without token get 403

**Independent Test**: Send request with Authorization header → verify GitHub calls use token; send request to private repo without token → get 403

### Implementation for User Story 3

- [ ] T038 [US3] Enhance handler in internal/handler/analysis.go — extract Bearer token from Authorization header; pass token through job payload to worker; pass to GitHub client and go-git clone options (BasicAuth with token as password)
- [ ] T039 [US3] Enhance GitHub client in internal/github/client.go — detect 404 vs 403 from GitHub API; when repo returns 404 with token → not_found; when repo returns 404 without token → check if private by attempting HEAD → return private_repo error
- [ ] T040 [US3] Write unit test for token passthrough in internal/handler/analysis_test.go — mock request with/without Authorization header; verify token reaches GitHub client; verify 403 response for private repo without token

**Checkpoint**: Authenticated access works independently

---

## Phase 6: User Story 4 — Cache Management (Priority: P2)

**Goal**: DELETE /api/:owner/:repo/cache busts cached results; all responses include cached/cached_at and X-Cache headers

**Independent Test**: Analyze repo → verify cached:true → DELETE cache → verify next request returns 202

### Implementation for User Story 4

- [ ] T041 [US4] Implement DELETE /api/:owner/:repo/cache handler in internal/handler/cache.go — validate owner/repo; call cache.Delete for the repo's result key pattern {owner}/{repo}:*:core; return 204 (idempotent, even if no entry existed)
- [ ] T042 [US4] Enhance analysis handler in internal/handler/analysis.go — ensure all 200 responses include cached:true, cached_at timestamp in body; ensure 202 responses include cached:false; set X-Cache header on all analysis responses
- [ ] T043 [US4] Write unit tests for cache handler in internal/handler/cache_test.go — verify DELETE returns 204; verify cache key is removed; verify idempotent behavior on missing key
- [ ] T044 [US4] Write unit tests for cache status fields in internal/handler/analysis_test.go — verify cached/cached_at fields in 200 response; verify X-Cache:HIT header on cached response; verify X-Cache:MISS on 202

**Checkpoint**: Cache management works independently

---

## Phase 7: User Story 5 — Shields.io Badge (Priority: P3)

**Goal**: GET /api/:owner/:repo/badge returns shields.io-compatible JSON with formatted line count and color

**Independent Test**: `curl /api/sindresorhus/is/badge` returns valid shields.io JSON

### Implementation for User Story 5

- [ ] T045 [US5] Implement badge handler in internal/handler/badge.go — check cache for analysis result; if found, format line count (500→"500", 1500→"1.5k", 25000→"25k", 1500000→"1.5M"), select color by threshold per contracts/api.md; return {"schemaVersion":1,"label":"lines of code","message":"<count>","color":"<color>"}; if not cached, return "analyzing..." with color "lightgrey" and trigger background job
- [ ] T046 [US5] Write unit tests for badge handler in internal/handler/badge_test.go — table-driven tests for line count formatting (500, 1500, 25000, 1500000); verify color thresholds (brightgreen, green, yellowgreen, yellow, orange, blue); verify "analyzing..." response when no cache

**Checkpoint**: Badge endpoint works independently

---

## Phase 8: User Story 6 — Graceful Error Handling (Priority: P3)

**Goal**: Structured error responses for all failure modes: 400 bad request, 403 private repo, 404 not found, 429 rate limit, 503 service unavailable, 504 timeout

**Independent Test**: Trigger each error condition and verify structured JSON error response

### Implementation for User Story 6

- [ ] T047 [US6] Implement input validation middleware in internal/handler/analysis.go — reject malformed owner/repo identifiers with 400 bad_request error; regex: owner and repo must match ^[a-zA-Z0-9._-]+$ and combined length ≤256
- [ ] T048 [US6] Implement rate limit forwarding in internal/github/client.go — when GitHub returns 429, extract Retry-After header and return typed RateLimitError; in handler, forward 429 to client with Retry-After header per contracts/api.md
- [ ] T049 [US6] Implement 5-minute analysis timeout in internal/job/worker.go — wrap analyzer.Analyze() call with context.WithTimeout(ctx, 5*time.Minute); on timeout, return 504 analysis_timeout with partial results if any sub-jobs completed
- [ ] T050 [US6] Implement service availability checks in internal/handler/analysis.go — check Redis and NATS connectivity before processing; return 503 service_unavailable with message identifying which service is down; implement Redis fallback: if Redis is down, process without cache and log warning
- [ ] T051 [US6] Write unit tests for error handling in internal/handler/analysis_test.go — table-driven: malformed input→400, nonexistent repo→404, rate limited→429 with Retry-After, service unavailable→503; verify all use consistent error envelope
- [ ] T052 [US6] Write unit test for analysis timeout in internal/job/worker_test.go — mock slow analyzer exceeding 5 min; verify 504 response with partial results

**Checkpoint**: All error conditions handled with structured responses

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T053 [P] Add structured JSON logging across all packages — use log/slog with JSON handler; log all operations (clone start/end, analysis start/end, cache hit/miss, job status transitions, errors) with structured fields (owner, repo, sha, duration, error)
- [ ] T054 [P] Implement clone cleanup goroutine in internal/clone/clone.go — background goroutine that scans CloneDir every 10 minutes and deletes directories older than 1 hour per FR-020
- [ ] T055 [P] Implement job deduplication check in internal/handler/analysis.go — before publishing new job, check active:{owner}/{repo}:{sha} Redis key; if exists, return existing job_id per FR-019; clear active key on job completion in worker
- [ ] T056 Wire Prometheus /metrics endpoint in cmd/grit/main.go — mount promhttp.Handler() on chi router at /metrics path
- [ ] T057 Run quickstart.md verification checklist — start docker compose, execute all 8 curl commands from quickstart.md, verify expected responses
- [ ] T058 [P] Write integration test in internal/analysis/core/analyzer_test.go — end-to-end test using testdata/fixture-repo (no external calls); verify complete AnalysisResult structure, line counts match manual count, language detection correct

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Foundational — this is the MVP
- **US2 (Phase 4)**: Depends on US1 (enhances worker and status handler)
- **US3 (Phase 5)**: Depends on Foundational — can run parallel with US1/US2
- **US4 (Phase 6)**: Depends on US1 (needs cache and analysis handler to exist)
- **US5 (Phase 7)**: Depends on US1 (needs cached analysis to format badge)
- **US6 (Phase 8)**: Depends on US1 (needs handlers and worker to add error paths)
- **Polish (Phase 9)**: Depends on all desired user stories being complete

### User Story Dependencies

- **US1 (P1)**: Can start after Foundational — no dependencies on other stories
- **US2 (P1)**: Enhances US1's worker and status handler — depends on US1
- **US3 (P2)**: Can start after Foundational — independent of US1/US2 but benefits from US1 handler existing
- **US4 (P2)**: Enhances US1's analysis handler — depends on US1
- **US5 (P3)**: Uses cached results from US1 — depends on US1
- **US6 (P3)**: Adds error paths to US1's handlers/worker — depends on US1

### Within Each User Story

- Models before services
- Services before handlers
- Core implementation before integration
- Tests alongside implementation (same phase)

### Parallel Opportunities

- **Phase 1**: T003, T004, T005 can all run in parallel
- **Phase 2**: T007, T008, T009, T011 can all run in parallel; T015 parallel with T014
- **Phase 3**: T016, T017, T018, T020 can all run in parallel; T029, T030 parallel
- **Phase 5+6**: US3 and US4 can run in parallel after US1 completes
- **Phase 7+8**: US5 and US6 can run in parallel after US1 completes
- **Phase 9**: T053, T054, T055, T058 can all run in parallel

---

## Implementation Strategy

### MVP First (User Story 1 + 2)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1 (core analysis flow)
4. Complete Phase 4: User Story 2 (async polling)
5. **STOP and VALIDATE**: Test end-to-end with `curl` per quickstart.md
6. Deploy via Docker Compose if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. US1 + US2 → Core analysis + async polling (MVP!)
3. US3 + US4 → Auth + cache management (parallel)
4. US5 + US6 → Badge + error handling (parallel)
5. Polish → Logging, cleanup, integration tests
6. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All tests use table-driven style per constitution Principle VIII
- External deps (GitHub, Redis, NATS) mocked in unit tests; integration tests use real services via Docker Compose
