# Tasks: AI Layer (Gemini)

**Input**: Design documents from `/specs/006-ai-gemini-layer/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/ai-api.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Create AI model types, add SDK dependency, and extend shared infrastructure (cache, config, metrics)

- [x] T001 Add `google.golang.org/genai` and `golang.org/x/time/rate` dependencies via `go get`
- [x] T002 Create AI model types (AISummary, ChatMessage, ChatRequest, HealthScore, HealthCategories, CategoryScore, ComplexFileSnippet) in `internal/models/ai.go`
- [x] T003 [P] Add `GeminiAPIKey` field to Config struct and load `GEMINI_API_KEY` env var in `internal/config/config.go`
- [x] T004 [P] Add `AISummaryTTL = 1 * time.Hour` and `AIHealthTTL = 6 * time.Hour` constants, cache key helpers, and Get/Set/Delete methods for AI summary and AI health in `internal/cache/redis.go`
- [x] T005 [P] Add AI Prometheus metrics (ai_request_duration_seconds histogram, ai_request_total counter with labels for feature and status) in `internal/metrics/prometheus.go`

---

## Phase 2: Foundational — Gemini Client & Context Builder

**Purpose**: Core AI infrastructure that ALL user stories depend on. MUST be complete before any story phase.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T006 Implement Gemini client wrapper in `internal/ai/client.go` — initialize `genai.NewClient` with APIKey and `BackendGeminiAPI`, expose primary model (`gemini-2.5-flash`) and fallback model (`gemini-2.5-flash-lite`) references, provide `GenerateStream` and `Generate` methods that delegate to the SDK
- [x] T007 Implement `retryWithBackoff` helper in `internal/ai/client.go` — 3 attempts, exponential backoff with jitter (base 1s, max 30s), switch to fallback model on attempt 3, catch 429/5xx errors, return masked errors per R7
- [x] T008 [P] Implement `BuildContext` pure function in `internal/ai/context.go` — accept repo metadata, file tree paths, README content, manifest map, complex file snippets, and directory summary; assemble into `[]genai.Part` with token budget enforcement (800K total, 4 chars/token approximation); truncate per priority order (file tree first, README last)
- [x] T009 [P] Write tests for `BuildContext` in `internal/ai/context_test.go` — test token budget enforcement, truncation priority, missing README handling, empty manifests, large file tree truncation
- [x] T010 Write tests for Gemini client retry logic in `internal/ai/client_test.go` — test successful call, retry on 429, model fallback on attempt 3, error masking, backoff timing

**Checkpoint**: Gemini client and context builder ready — user story implementation can begin.

---

## Phase 3: User Story 4 — Shared Context Construction (Priority: P1) 🎯 MVP Foundation

**Goal**: Assemble a rich context payload from cached analysis data for all AI features.

**Independent Test**: Construct context for a known repository and verify all sections are present, token limits are respected, and truncation works correctly.

> Note: US4 is implemented first because US1, US2, US3 all depend on context assembly from cached data. The `BuildContext` function was created in Phase 2; this phase adds the higher-level `AssembleContext` that reads from cache and the cloned repo.

- [x] T011 [US4] Implement `AssembleContext` in `internal/ai/context.go` — read core analysis from cache to get file tree and metadata, read complexity results for top 5 files, read README.md and package manifests (package.json, go.mod, Cargo.toml, requirements.txt, pyproject.toml, pom.xml, build.gradle) from the cloned repo, compute directory summary, delegate to `BuildContext`
- [x] T012 [US4] Write tests for `AssembleContext` in `internal/ai/context_test.go` — test with full cache data, missing complexity data (proceeds without complex files section), missing README, no manifests found

**Checkpoint**: Context construction fully functional and testable independently.

---

## Phase 4: User Story 1 — AI Codebase Summary (Priority: P1) 🎯 MVP

**Goal**: Stream a structured AI codebase summary via SSE, cache the result.

**Independent Test**: POST to summary endpoint for a repo with cached core analysis → receive SSE stream → verify structured fields → second request returns cached JSON with `X-Cache: HIT`.

- [x] T013 [US1] Implement summary generation logic in `internal/ai/summary.go` — build summary prompt instructing Gemini to return project description, architecture, tech stack, red flags, entry points; call `GenerateStream` via the client wrapper; collect streamed chunks; parse final result into `models.AISummary`
- [x] T014 [US1] Write tests for summary generation in `internal/ai/summary_test.go` — mock Gemini streaming response, verify prompt structure, verify AISummary parsing from streamed output
- [x] T015 [US1] Implement SSE streaming handler in `internal/handler/ai_summary.go` — validate owner/repo, check GEMINI_API_KEY (503 if missing), check core analysis cache (409 if missing), check AI summary cache (return cached with X-Cache HIT), otherwise call summary generation and stream via `http.Flusher`, cache complete result with 1h TTL
- [x] T016 [US1] Wire summary route `POST /api/{owner}/{repo}/ai/summary` in `cmd/grit/main.go` — create AI handler struct, register route

**Checkpoint**: AI summary endpoint fully functional — streams on first request, returns cached on second.

---

## Phase 5: User Story 2 — Ask This Repo Chat (Priority: P2)

**Goal**: Stream AI chat responses with per-IP rate limiting and turn management.

**Independent Test**: POST chat messages → receive SSE stream → verify context is prepended → verify rate limit kicks in after 10 requests/minute.

- [x] T017 [P] [US2] Implement per-IP rate limiter in `internal/ai/ratelimit.go` — `NewRateLimiter` returning struct with `sync.Map` of `rate.Limiter` per IP, `Allow(ip string) bool` method, background cleanup goroutine evicting stale entries every 5 minutes
- [x] T018 [P] [US2] Write tests for rate limiter in `internal/ai/ratelimit_test.go` — test allow within limit, deny over limit, per-IP isolation, cleanup of stale entries
- [x] T019 [US2] Implement chat logic in `internal/ai/chat.go` — validate ChatRequest (non-empty, last role is "user", max 20 turns with oldest truncation), prepend context as system message, call `GenerateStream` with full conversation, collect streamed chunks
- [x] T020 [US2] Write tests for chat logic in `internal/ai/chat_test.go` — test turn truncation at 20, system context prepend, validation errors (empty messages, last role not user, whitespace content)
- [x] T021 [US2] Implement SSE streaming chat handler in `internal/handler/ai_chat.go` — validate owner/repo, check API key (503), check core analysis cache (409), check IP rate limit (429), decode ChatRequest body (400 on invalid), call chat logic and stream via `http.Flusher`
- [x] T022 [US2] Wire chat route `POST /api/{owner}/{repo}/ai/chat` in `cmd/grit/main.go`

**Checkpoint**: Chat endpoint functional — streams responses, enforces rate limit, manages turns.

---

## Phase 6: User Story 3 — AI Health Score (Priority: P3)

**Goal**: Return a structured JSON health assessment with caching.

**Independent Test**: GET health endpoint → receive JSON with overall score (0-100), 5 categories, and improvements list → second request returns cached with `X-Cache: HIT`.

- [x] T023 [US3] Implement health score generation in `internal/ai/health.go` — build health prompt with JSON schema instruction, call `Generate` (non-streaming) with `ResponseMIMEType: "application/json"`, parse response into `models.HealthScore`, on JSON parse failure retry once with stricter prompt including exact schema
- [x] T024 [US3] Write tests for health score generation in `internal/ai/health_test.go` — mock Gemini JSON response, test successful parse, test retry on malformed JSON, test final failure after both attempts
- [x] T025 [US3] Implement health score handler in `internal/handler/ai_health.go` — validate owner/repo, check API key (503), check core analysis cache (409), check health score cache (return cached with X-Cache HIT), otherwise generate and cache with 6h TTL, return JSON
- [x] T026 [US3] Wire health route `GET /api/{owner}/{repo}/ai/health` in `cmd/grit/main.go`

**Checkpoint**: Health score endpoint functional — generates structured JSON, caches results.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Integration, cache cleanup, build verification, and final validation.

- [x] T027 Add AI summary and AI health cache deletion to `HandleDeleteCache` in `internal/handler/cache.go`
- [x] T028 [P] Embed AI summary status into core analysis response via `embedAISummary` method in `internal/handler/analysis.go` — follow existing pattern (complexity_summary, churn_summary), show status pending/complete with `ai_summary_url`
- [x] T029 Increase `WriteTimeout` to 120s in `cmd/grit/main.go` for long SSE streaming connections
- [x] T030 Verify all existing tests still pass (`go test ./...`)
- [x] T031 Verify build succeeds (`go build ./...`)
- [x] T032 Run quickstart.md validation — confirm 503 when no API key, confirm 409 when no core analysis

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 completion — BLOCKS all user stories
- **US4 Context (Phase 3)**: Depends on Phase 2 — provides context assembly for US1/US2/US3
- **US1 Summary (Phase 4)**: Depends on Phases 2 + 3
- **US2 Chat (Phase 5)**: Depends on Phases 2 + 3; independent of US1
- **US3 Health (Phase 6)**: Depends on Phases 2 + 3; independent of US1/US2
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **US4 (P1)**: Foundation for all AI features — must complete first after Phase 2
- **US1 (P1)**: Depends on US4 context construction only — can run after Phase 3
- **US2 (P2)**: Depends on US4 context construction only — can run in parallel with US1
- **US3 (P3)**: Depends on US4 context construction only — can run in parallel with US1/US2

### Within Each User Story

- Logic/service before handler
- Handler before route wiring
- Tests alongside implementation (same phase)
- Story complete before moving to next priority (sequential) or all in parallel (if staffed)

### Parallel Opportunities

- T003, T004, T005 can run in parallel (Phase 1 — different files)
- T008, T009 can run in parallel with T006, T007 (Phase 2 — different files)
- T017, T018 can run in parallel with T019, T020 (Phase 5 — different files)
- US1, US2, US3 can all run in parallel after US4 completes (if team capacity allows)

---

## Parallel Example: Phase 1 Setup

```
# All three can run simultaneously (different files):
T003: Add GeminiAPIKey to config.go
T004: Add AI cache methods to redis.go
T005: Add AI metrics to prometheus.go
```

## Parallel Example: Phase 2 Foundational

```
# Client and context can be developed in parallel:
Stream 1: T006 → T007 → T010 (client + retry + tests)
Stream 2: T008 → T009 (context builder + tests)
```

## Parallel Example: User Stories (after Phase 3)

```
# All three stories can proceed simultaneously:
Stream A: T013 → T014 → T015 → T016 (US1 Summary)
Stream B: T017 → T018 → T019 → T020 → T021 → T022 (US2 Chat)
Stream C: T023 → T024 → T025 → T026 (US3 Health)
```

---

## Implementation Strategy

### MVP First (US4 + US1 Only)

1. Complete Phase 1: Setup (T001–T005)
2. Complete Phase 2: Foundational — Gemini client + context builder (T006–T010)
3. Complete Phase 3: US4 — Context assembly from cache (T011–T012)
4. Complete Phase 4: US1 — AI Summary with SSE (T013–T016)
5. **STOP and VALIDATE**: Test summary endpoint end-to-end

### Incremental Delivery

1. Setup + Foundational + US4 → Context construction works
2. Add US1 → Summary streaming works → **MVP deployed**
3. Add US2 → Chat with rate limiting works
4. Add US3 → Health score with JSON parsing works
5. Polish → Cache cleanup, embedding, timeout adjustment

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- US4 (context construction) is prerequisite for all other stories
- US1/US2/US3 are independently testable after US4
- Commit after each task or logical group
- The `google.golang.org/genai` SDK must be added before any AI code compiles
- SSE handlers need `http.Flusher` assertion and proper Content-Type headers
- Never expose raw Gemini errors — always mask through client wrapper
