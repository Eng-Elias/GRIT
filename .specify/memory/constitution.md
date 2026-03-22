<!--
  Sync Impact Report
  ==================
  Version change: N/A → 1.0.0 (initial creation)
  Modified principles: N/A (first ratification)
  Added sections:
    - Core Principles (8 principles)
    - Technology Stack (non-negotiable)
    - Caching & Performance Standards
    - Governance
  Removed sections: N/A
  Templates requiring updates:
    - .specify/templates/plan-template.md ✅ no updates needed (generic)
    - .specify/templates/spec-template.md ✅ no updates needed (generic)
    - .specify/templates/tasks-template.md ✅ no updates needed (generic)
  Follow-up TODOs: None
-->

# GRIT (Git Repo Intelligence Tool) Constitution

## Core Principles

### I. API-First Design

All analysis results MUST be exposed as structured JSON endpoints
consumed by the React frontend. There MUST be no server-side
rendering. The Go backend serves both the JSON API and the compiled
React static files from a single binary.

- Every feature MUST start as a JSON API endpoint before any UI work.
- API response schemas MUST be documented and stable within a major
  version.
- The frontend MUST be a pure consumer of the API — no direct
  database or cache access.

### II. Modular Analysis Pillars

Each analysis pillar (core, complexity, churn, contributor, temporal,
AI) MUST be independently runnable and testable in isolation.

- Every pillar MUST reside in its own Go package under a shared
  `internal/analysis/` directory.
- Pillars MUST NOT import from each other; shared types belong in a
  dedicated `internal/models/` package.
- A pillar MUST be executable via its own integration test without
  requiring other pillars to be running.
- Adding or removing a pillar MUST NOT break the build or other
  pillars.

### III. Async-First Execution

Heavy analysis jobs MUST run in a background job queue. The frontend
MUST poll a status endpoint until the job completes.

- NATS with JetStream is the job queue; no other queue system is
  permitted.
- Every analysis request MUST return an immediate `202 Accepted`
  response with a job ID.
- The `/api/v1/jobs/{id}` status endpoint MUST report `queued`,
  `running`, `completed`, or `failed`.
- Workers MUST be horizontally scalable — no shared in-process state.

### IV. Cache-First with Redis

Redis is the sole cache layer. The system MUST never re-analyze a
repository at the same commit SHA if a cached result exists.

- Cache keys MUST follow the format `{owner}/{repo}:{commit_sha}:{pillar}`.
- TTLs are non-negotiable per pillar:
  - **Core analysis**: 24 h
  - **Complexity + churn**: 24 h
  - **Contributor (blame)**: 48 h
  - **Temporal**: 12 h
  - **AI summary**: 1 h
  - **AI health score**: 6 h
- Cache misses MUST trigger an async job; cache hits MUST return
  immediately.
- Redis MUST be the only cache layer — no in-process caches.

### V. Defensive AI Integration

Gemini 2.5 Flash (via `google.golang.org/genai`) is the AI provider.
AI features MUST be on-demand only — never triggered automatically.

- The system MUST never call Gemini twice for the same
  `{owner}/{repo}:{commit_sha}` combination.
- HTTP 429 rate-limit errors MUST be handled gracefully: return
  `{ "error": "rate_limited", "retry_after": 60 }` to the client.
  Raw Gemini errors MUST NOT be exposed.
- Retries MUST use exponential backoff with jitter (max 3 attempts).
- On the final retry, the system MUST automatically fall back to
  Gemini 2.5 Flash-Lite.
- All AI results MUST be cached per the TTLs defined in Principle IV.

### VI. Self-Hostable by Default

GRIT MUST run on a single VPS via Docker Compose with no hard
dependencies on external cloud services beyond the GitHub API and
Gemini API.

- All configuration MUST be environment-variable-based; no hardcoded
  secrets or config files committed to the repository.
- `docker compose up` MUST be sufficient to start the entire stack
  (Go backend, Redis, NATS, Caddy).
- The Docker Compose file MUST pin all image versions explicitly.
- Secrets (API keys) MUST be injected via `.env` files or Docker
  secrets — never baked into images.

### VII. Clean Handler Separation

No business logic is permitted in HTTP handlers.

- Handlers MUST only perform: request parsing, input validation,
  calling a service method, and writing the response.
- All domain logic MUST reside in service packages under `internal/`.
- All external API calls (GitHub, Gemini) MUST have configurable
  timeout and retry logic.
- Error responses MUST use a consistent JSON envelope:
  `{ "error": "<code>", "message": "<human-readable>" }`.

### VIII. Test Discipline

Every service package MUST have unit tests. Tests MUST exercise
real behavior, not implementation details.

- Table-driven tests are the default style for Go service packages.
- External dependencies (GitHub API, Gemini API, Redis, NATS) MUST
  be mocked or use test containers — never hit live services in CI.
- Integration tests MUST exist for each analysis pillar verifying
  end-to-end correctness against a known fixture repository.
- Test coverage MUST NOT decrease on any PR.

## Technology Stack (Non-Negotiable)

The following stack is fixed and MUST NOT be substituted without a
constitutional amendment:

| Layer          | Technology                                         |
|----------------|----------------------------------------------------|
| Backend        | Go 1.22, chi router, go-git v5                     |
| AI SDK         | `google.golang.org/genai` (Gemini 2.5 Flash)       |
| Frontend       | React 18 + TypeScript, Recharts, D3 v7, TailwindCSS|
| Cache          | Redis 7                                            |
| Job Queue      | NATS with JetStream                                |
| Reverse Proxy  | Caddy                                              |
| Orchestration  | Docker Compose                                     |
| AI Fallback    | Gemini 2.5 Flash-Lite                              |

- Dependency additions MUST be justified and reviewed.
- Major version upgrades (e.g., React 19, Go 1.23) require explicit
  review against this constitution before adoption.

## Caching & Performance Standards

- All analysis endpoints MUST check the Redis cache before
  dispatching a job.
- API responses for cached results MUST include an `X-Cache: HIT`
  header; cache misses MUST include `X-Cache: MISS`.
- The backend MUST serve the compiled React frontend as static files
  — no separate frontend server in production.
- Long-running analysis requests MUST NOT block HTTP connections;
  the async job pattern (Principle III) is mandatory.
- GitHub API calls MUST respect rate limits and use conditional
  requests (`If-None-Match` / ETags) where possible.

## Governance

- This constitution supersedes all other development practices,
  conventions, or ad-hoc decisions for the GRIT project.
- Amendments MUST be documented with a version bump, rationale, and
  migration plan for any affected code.
- Version increments follow Semantic Versioning:
  - **MAJOR**: Principle removal or backward-incompatible redefinition.
  - **MINOR**: New principle, new section, or material expansion.
  - **PATCH**: Wording clarifications, typo fixes, non-semantic
    refinements.
- All pull requests MUST verify compliance with this constitution
  before merge.
- Complexity beyond what the constitution prescribes MUST be
  justified in writing and linked to a specific requirement.

**Version**: 1.0.0 | **Ratified**: 2026-03-22 | **Last Amended**: 2026-03-22
