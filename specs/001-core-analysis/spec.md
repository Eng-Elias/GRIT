# Feature Specification: Core Analysis Engine

**Feature Branch**: `001-core-analysis`
**Created**: 2026-03-22
**Status**: Draft
**Input**: User description: "Build the core analysis engine for GRIT — the foundation pillar that all others depend on."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Analyze a Public Repository (Priority: P1)

A developer enters a GitHub repository identifier (e.g., `facebook/react`)
into the GRIT web UI or calls the API directly. The system clones/fetches the
repository, walks the file tree, counts lines (total, code, comment, blank)
per file grouped by language, fetches GitHub metadata and commit activity,
and returns a comprehensive JSON analysis result.

**Why this priority**: This is the fundamental value proposition of GRIT —
without line counting and metadata retrieval, no other pillar can function.
This story delivers a working end-to-end flow from request to result.

**Independent Test**: Can be fully tested by calling
`GET /api/:owner/:repo` for any public GitHub repository and verifying
the response contains file tree statistics, language breakdown, GitHub
metadata, and commit activity data.

**Acceptance Scenarios**:

1. **Given** a valid public repository identifier `torvalds/linux`,
   **When** the user requests `GET /api/torvalds/linux`,
   **Then** the system returns HTTP 202 with a job ID (since the repo is
   not cached), and subsequent polling of the status endpoint eventually
   returns a complete analysis with line counts per language, GitHub
   metadata (stars, forks, license, etc.), and 52-week commit activity.

2. **Given** a valid public repository identifier `sindresorhus/is`,
   **When** the user requests `GET /api/sindresorhus/is`,
   **Then** the system returns a complete analysis with per-file and
   per-language breakdowns of total lines, code lines, comment lines,
   and blank lines, with files matching `.gitignore` patterns excluded.

3. **Given** a repository that was previously analyzed at the same commit
   SHA, **When** the user requests `GET /api/:owner/:repo`,
   **Then** the system returns the cached result immediately (HTTP 200)
   with `cached: true` and `cached_at` timestamp in the response body
   and an `X-Cache: HIT` header.

---

### User Story 2 - Async Job Status Polling (Priority: P1)

When analysis is not cached, the system enqueues the work as a background
job and the client polls a status endpoint to track progress until the
analysis completes or fails.

**Why this priority**: The async pattern is constitutionally mandated
(Principle III) and is required for the primary analysis flow to work for
any non-trivial repository. Without it, HTTP connections would block for
minutes on large repos.

**Independent Test**: Can be tested by requesting analysis of a repository
not in cache and verifying the 202 → polling → complete flow works
end-to-end, including progress reporting for sub-jobs.

**Acceptance Scenarios**:

1. **Given** a repository not in cache, **When** the user requests
   `GET /api/:owner/:repo`, **Then** the system returns HTTP 202 with
   `{ "job_id": "<uuid>", "status": "queued" }`.

2. **Given** a queued job ID, **When** the user polls
   `GET /api/:owner/:repo/status`, **Then** the response includes
   `status` (one of `queued`, `running`, `completed`, `failed`) and
   a `progress` object with sub-job status for clone, file-walk,
   metadata-fetch, and commit-activity-fetch.

3. **Given** a completed job, **When** the user polls the status endpoint,
   **Then** the response includes `status: "completed"` and the full
   analysis result is available at `GET /api/:owner/:repo`.

---

### User Story 3 - Authenticated Access for Higher Rate Limits (Priority: P2)

A developer provides a GitHub personal access token via the
`Authorization: Bearer <token>` header to get higher GitHub API rate
limits and optionally analyze repositories they have access to.

**Why this priority**: Public API access works without authentication, but
authenticated access significantly increases rate limits (from 60 to 5000
requests/hour) and is essential for power users.

**Independent Test**: Can be tested by sending a request with and without
an `Authorization` header and verifying that the GitHub API calls use the
token when provided.

**Acceptance Scenarios**:

1. **Given** a request with a valid `Authorization: Bearer <token>` header,
   **When** the system makes GitHub API calls, **Then** it uses the
   provided token for authentication, resulting in higher rate limits.

2. **Given** a request without an `Authorization` header for a public
   repository, **When** the system processes the request, **Then** it
   proceeds using unauthenticated GitHub API access.

3. **Given** a request without a token for a private repository,
   **When** the system attempts to access the repo, **Then** it returns
   HTTP 403 with `{ "error": "private_repo", "message": "This repository
   is private. Provide a GitHub token via Authorization header." }`.

---

### User Story 4 - Cache Management (Priority: P2)

A user can manually bust the cache for a specific repository to force
re-analysis, and all API responses clearly indicate cache status.

**Why this priority**: Cache management is important for users who know a
repository has been updated and want fresh results without waiting for TTL
expiry.

**Independent Test**: Can be tested by analyzing a repository, then calling
`DELETE /api/:owner/:repo/cache` and verifying subsequent requests trigger
a fresh analysis.

**Acceptance Scenarios**:

1. **Given** a cached analysis for `owner/repo`, **When** the user sends
   `DELETE /api/:owner/:repo/cache`, **Then** the system removes the
   cached entry and returns HTTP 204.

2. **Given** a cache bust was performed, **When** the user requests
   `GET /api/:owner/:repo`, **Then** a fresh analysis is enqueued
   (HTTP 202) and the eventual result has `cached: false`.

3. **Given** any analysis response, **When** the result is returned,
   **Then** it includes `cached` (boolean) and `cached_at` (ISO 8601
   timestamp or null) fields, plus an `X-Cache: HIT` or `X-Cache: MISS`
   response header.

---

### User Story 5 - Shields.io Compatible Badge (Priority: P3)

A user can embed a lines-of-code badge in their README using the badge
endpoint, which returns a shields.io-compatible JSON response (same format
as ghloc-web).

**Why this priority**: Badges are a highly visible feature that drives
adoption, but they depend on the core analysis being functional first.

**Independent Test**: Can be tested by calling `GET /api/:owner/:repo/badge`
and verifying the response matches the shields.io JSON endpoint schema.

**Acceptance Scenarios**:

1. **Given** a repository with a cached analysis, **When** the user
   requests `GET /api/:owner/:repo/badge`, **Then** the system returns
   a JSON object with `schemaVersion`, `label`, `message` (formatted
   line count), and `color` fields compatible with shields.io.

2. **Given** a repository without a cached analysis, **When** the user
   requests the badge endpoint, **Then** the system returns a badge
   with `message: "analyzing..."` and `color: "lightgrey"` while
   triggering a background analysis job.

---

### User Story 6 - Graceful Error Handling (Priority: P3)

The system handles all expected failure modes gracefully, returning
structured error responses that help the client understand what went
wrong and what to do about it.

**Why this priority**: Robust error handling is essential for a production
system but depends on the primary analysis flow being implemented first.

**Independent Test**: Can be tested by triggering each error condition
(nonexistent repo, rate limit, timeout) and verifying the structured
error response.

**Acceptance Scenarios**:

1. **Given** a nonexistent repository identifier, **When** the user
   requests `GET /api/nonexistent/repo`, **Then** the system returns
   HTTP 404 with `{ "error": "not_found", "message": "Repository
   nonexistent/repo does not exist." }`.

2. **Given** the GitHub API returns HTTP 429, **When** the system
   receives the rate-limit response, **Then** it returns HTTP 429
   to the client with the `Retry-After` header value forwarded from
   GitHub.

3. **Given** an analysis that exceeds 5 minutes, **When** the timeout
   is reached, **Then** the system returns HTTP 504 with partial
   results if any sub-jobs completed, or a timeout error if none did.

---

### Edge Cases

- What happens when a repository has no files (empty repo)?
  The system MUST return a valid analysis with zero line counts.
- What happens when a repository has an extremely large file tree (>100k files)?
  The system MUST handle it within the 5-minute timeout or return
  partial results.
- What happens when `.gitignore` references patterns that match all files?
  The system MUST still return a valid response with zero counts.
- What happens when the Redis cache is unavailable?
  The system MUST fall back to processing without cache and log a
  warning — never crash or hang.
- What happens when NATS is unavailable?
  The system MUST return HTTP 503 with a clear service-unavailable
  message.
- What happens when a repository name contains special characters?
  The system MUST validate the `owner/repo` format and reject
  malformed identifiers with HTTP 400.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept a GitHub repository identifier
  (`owner/repo`) via URL path parameters on all API endpoints.
- **FR-002**: System MUST optionally accept a GitHub personal access
  token via `Authorization: Bearer <token>` header and use it for
  all GitHub API calls when provided.
- **FR-003**: System MUST clone or shallow-fetch repositories using
  go-git (no shelling out to the `git` CLI).
- **FR-004**: System MUST walk the entire file tree and count total
  lines, code lines, comment lines, and blank lines per file.
- **FR-005**: System MUST group and aggregate line counts by
  programming language, detected from file extensions using a
  built-in map covering 40+ languages.
- **FR-006**: System MUST respect `.gitignore` patterns when walking
  the file tree, excluding matched files from analysis.
- **FR-007**: System MUST fetch GitHub repository metadata via the
  REST API: name, description, homepage, stars, forks, watchers,
  open issues, size (KB), default branch, primary language, license
  (name + SPDX ID), topics, created/updated/pushed dates, has_wiki,
  has_projects, has_discussions, and community health percentage.
- **FR-008**: System MUST fetch weekly commit counts for the past 52
  weeks via the GitHub Stats API, plus total commit count on the
  default branch.
- **FR-009**: System MUST cache all analysis results in Redis keyed
  by `{owner}/{repo}:{default_branch_sha}` with a 24-hour TTL.
- **FR-010**: System MUST include `cached` (boolean) and `cached_at`
  (ISO 8601 timestamp or null) in every analysis response body.
- **FR-011**: System MUST set `X-Cache: HIT` or `X-Cache: MISS`
  response headers on all analysis responses.
- **FR-012**: System MUST support manual cache invalidation via
  `DELETE /api/:owner/:repo/cache`.
- **FR-013**: System MUST return HTTP 202 with a job ID for
  uncached analysis requests and process them asynchronously via
  the NATS JetStream job queue.
- **FR-014**: System MUST expose a status polling endpoint at
  `GET /api/:owner/:repo/status` reporting job progress including
  sub-job status.
- **FR-015**: System MUST expose a shields.io-compatible badge
  endpoint at `GET /api/:owner/:repo/badge`.
- **FR-016**: System MUST return HTTP 403 for private repositories
  accessed without a token, HTTP 404 for nonexistent repositories,
  HTTP 429 with forwarded `Retry-After` for GitHub rate limits, and
  HTTP 504 with partial results for analysis timeouts exceeding 5
  minutes.
- **FR-017**: System MUST use a consistent JSON error envelope:
  `{ "error": "<code>", "message": "<human-readable>" }` for all
  error responses.
- **FR-018**: System MUST validate the `owner/repo` format and
  reject malformed identifiers with HTTP 400.

### Key Entities

- **Repository**: Represents a GitHub repository identified by
  `owner/repo`. Key attributes: owner, name, metadata (stars,
  forks, license, etc.), default branch, latest commit SHA.
- **AnalysisResult**: The output of the core analysis for a
  repository at a specific commit. Key attributes: file tree
  statistics (per-file and per-language line counts), GitHub
  metadata snapshot, commit activity (52-week weekly counts),
  timestamps (analyzed_at, cached_at).
- **FileStats**: Per-file line count breakdown. Key attributes:
  file path, language, total lines, code lines, comment lines,
  blank lines, byte size.
- **LanguageBreakdown**: Aggregated statistics per detected
  language. Key attributes: language name, file count, total
  lines, code lines, comment lines, blank lines, percentage of
  total.
- **AnalysisJob**: A background job representing an in-progress
  analysis. Key attributes: job ID, repository identifier,
  status (queued/running/completed/failed), sub-job progress,
  created_at, completed_at, error (if failed).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Analysis of a repository with up to 10,000 files
  completes within 2 minutes end-to-end (clone + walk + metadata +
  commit activity).
- **SC-002**: Cached analysis results are returned within 50ms
  (p95 latency).
- **SC-003**: The system correctly identifies and categorizes files
  for at least 40 programming languages by file extension.
- **SC-004**: Line count accuracy matches or exceeds ghloc-web for
  the same repository at the same commit — verified against 5
  reference repositories.
- **SC-005**: All four API endpoints return valid, parseable JSON
  responses for every request (success and error cases).
- **SC-006**: The badge endpoint produces valid shields.io JSON
  that renders correctly when used in a GitHub README.
- **SC-007**: The system handles GitHub API rate limiting without
  crashing or exposing raw upstream errors, returning structured
  429 responses with `Retry-After`.
- **SC-008**: The async job flow (202 → polling → complete)
  functions correctly for repositories of all sizes, with status
  updates reflecting real sub-job progress.
- **SC-009**: Cache invalidation via DELETE endpoint causes the next
  GET to trigger a fresh analysis.
- **SC-010**: The system operates correctly in Docker Compose with
  Redis and NATS as the only infrastructure dependencies.
