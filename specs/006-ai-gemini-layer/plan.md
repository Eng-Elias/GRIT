# Implementation Plan: AI Layer (Gemini)

**Branch**: `006-ai-gemini-layer` | **Date**: 2026-04-15 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-ai-gemini-layer/spec.md`

## Summary

Add AI-powered intelligence to GRIT using Google Gemini. Three features: streaming AI codebase summary, interactive repo chat, and structured health score — all sharing a common context construction pipeline. The Gemini client uses `google.golang.org/genai` initialized once at startup, with retry/fallback logic (Flash → Flash-Lite on attempt 3). Streaming via SSE using `http.Flusher`. Chat rate-limited via `golang.org/x/time/rate` per IP in a `sync.Map`.

## Technical Context

**Language/Version**: Go 1.22
**Primary Dependencies**: google.golang.org/genai, golang.org/x/time/rate, chi router, go-git v5
**Storage**: Redis 7 (cache layer for AI summaries and health scores)
**Testing**: go test with table-driven tests, httptest for SSE, mock Gemini responses
**Target Platform**: Linux server (Docker Compose), single binary
**Project Type**: web-service (API backend)
**Performance Goals**: SSE first-chunk < 3s, cached results < 100ms, context construction < 2s for 10K files
**Constraints**: Total AI context < 800,000 tokens, chat rate limit 10 req/min/IP, Gemini free tier quota
**Scale/Scope**: Single-instance deployment, moderate concurrent users

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. API-First Design | PASS | All AI features exposed as JSON API endpoints; no SSR |
| II. Modular Analysis Pillars | PASS | AI lives in `internal/ai/` package; reads from models, does not import other pillars |
| III. Async-First Execution | PASS | Summary/chat stream via SSE (non-blocking); health score is synchronous but bounded (30s timeout) |
| IV. Cache-First with Redis | PASS | AI summary TTL 1h, health score TTL 6h, chat uncached (per constitution); key pattern matches `{owner}/{repo}:{sha}:{pillar}` |
| V. Defensive AI Integration | PASS | Gemini 2.5 Flash primary, Flash-Lite fallback on attempt 3; 3 retries with exponential backoff + jitter; rate-limit errors masked; on-demand only |
| VI. Self-Hostable | PASS | Only env var (`GEMINI_API_KEY`) needed; no cloud dependencies beyond Gemini API |
| VII. Clean Handler Separation | PASS | Handlers only parse/validate/respond; AI logic in `internal/ai/` service package |
| VIII. Test Discipline | PASS | Table-driven tests, mocked Gemini API via httptest, no live API calls in CI |

**Gate result**: ALL PASS — no violations.

## Project Structure

### Documentation (this feature)

```text
specs/006-ai-gemini-layer/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── ai-api.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/
├── ai/
│   ├── client.go          # Gemini client wrapper (init, retry, fallback)
│   ├── client_test.go
│   ├── context.go          # Shared context construction (pure function)
│   ├── context_test.go
│   ├── summary.go          # AI summary generation logic
│   ├── summary_test.go
│   ├── chat.go             # Chat logic (turn management, context prepend)
│   ├── chat_test.go
│   ├── health.go           # Health score generation + JSON parsing
│   ├── health_test.go
│   ├── ratelimit.go        # Per-IP token bucket rate limiter
│   └── ratelimit_test.go
├── cache/
│   └── redis.go            # + AI summary/health cache methods
├── config/
│   └── config.go           # + GEMINI_API_KEY env var
├── handler/
│   ├── ai_summary.go       # POST /api/:owner/:repo/ai/summary
│   ├── ai_chat.go          # POST /api/:owner/:repo/ai/chat
│   ├── ai_health.go        # GET /api/:owner/:repo/ai/health
│   └── cache.go            # + AI cache deletion
├── models/
│   └── ai.go               # AISummary, ChatMessage, ChatRequest, HealthScore, etc.
└── metrics/
    └── prometheus.go        # + AI request duration/count metrics

cmd/grit/
└── main.go                  # + Gemini client init, AI handler wiring, routes
```

**Structure Decision**: AI logic in a new `internal/ai/` package following the modular pillar pattern. Handlers in `internal/handler/ai_*.go`. Models in `internal/models/ai.go`. This matches the existing project structure for other pillars (churn, complexity, blame, temporal).

## Complexity Tracking

> No constitution violations — this section is empty.
