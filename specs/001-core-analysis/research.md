# Research: Core Analysis Engine

**Branch**: `001-core-analysis`
**Date**: 2026-03-23

## R1: go-git Shallow Clone — Memory vs Disk Strategy

**Decision**: Use go-git's `memory.NewStorage()` for repositories under
50 MB (GitHub API `size` field in KB), and `filesystem.NewStorage()` to
`/tmp/grit-clones/{owner}/{repo}/{sha}` for larger repositories.

**Rationale**: In-memory cloning avoids disk I/O entirely and is
significantly faster for small-to-medium repos. For larger repos,
in-memory cloning risks OOM on a constrained VPS. The 50 MB threshold
provides a safe margin: most popular open-source repos are under 50 MB
(compressed), and the few that exceed it (e.g., `torvalds/linux` at ~4 GB)
must use disk.

**Alternatives considered**:
- Always disk-clone: Simpler but unnecessarily slow for small repos.
- Always in-memory: Dangerous for large repos; no size limit safety.
- Configurable threshold via env var: Over-engineering for now; can be
  added later if needed.

**Implementation notes**:
- Use `git.CloneContext()` with `&git.CloneOptions{Depth: 1, SingleBranch: true}`
- Determine repo size via GitHub REST API `GET /repos/{owner}/{repo}`
  response field `size` (in KB) before cloning.
- Disk clones use a temp directory with 1-hour TTL cleanup goroutine.

## R2: Line Counting — Code vs Comment vs Blank Detection

**Decision**: Use a simple state-machine approach per file based on
language-specific comment syntax. Detect single-line comments (`//`, `#`,
`--`, `%`, `;`), multi-line comment delimiters (`/* */`, `<!-- -->`,
`""" """`, `=begin/=end`), and blank lines (whitespace-only).

**Rationale**: A full parser (tree-sitter, etc.) would be more accurate
but adds significant complexity and build-time dependencies. The
extension-based approach with comment-syntax tables matches ghloc-web's
accuracy level and is sufficient for the core pillar.

**Alternatives considered**:
- tree-sitter parsing: Accurate but heavy; requires language grammars.
- cloc/scc integration: Shells out to external binary — violates the
  no-exec constraint and adds a Docker dependency.
- Byte-counting only (no code/comment split): Too simplistic; users
  expect the breakdown.

**Implementation notes**:
- `languages.go` maps file extensions to `LanguageDef` struct containing:
  `Name`, `LineCommentPrefixes []string`, `BlockCommentStart string`,
  `BlockCommentEnd string`.
- Files with unrecognized extensions are categorized as "Other".
- Binary files (detected by null-byte scan of first 512 bytes) are
  excluded from line counting but included in file count.

## R3: GitHub REST API — Endpoints and Rate Limiting

**Decision**: Use a typed HTTP client wrapping `net/http` with:
- `GET /repos/{owner}/{repo}` — metadata
- `GET /repos/{owner}/{repo}/stats/participation` — 52-week commit counts
- `GET /repos/{owner}/{repo}/community/profile` — community health %
- `GET /repos/{owner}/{repo}/commits?per_page=1` — total commit count
  via `Link` header `last` page number

**Rationale**: The GitHub REST API v3 is well-documented, stable, and
sufficient for all metadata needs. GraphQL would reduce round trips but
adds query complexity for minimal gain in this feature.

**Alternatives considered**:
- GraphQL API v4: Fewer requests but requires auth for all calls;
  unauthenticated access is REST-only.
- go-github library: Adds a large dependency; a slim typed client
  wrapping net/http is lighter and gives full control over retry logic.

**Implementation notes**:
- GitHub Stats API returns HTTP 202 when stats are being computed.
  Retry with exponential backoff: delays of 2s, 4s, 8s, 8s, 8s
  (max 5 attempts, ~30s total).
- Rate limit: check `X-RateLimit-Remaining` header; if 0, return 429
  to client with `Retry-After` from `X-RateLimit-Reset`.
- Conditional requests: send `If-None-Match` with cached ETag for
  metadata endpoint to reduce rate limit consumption.
- Timeouts: 10s per HTTP request; 30s total for all GitHub API calls.

## R4: NATS JetStream — Job Queue Pattern

**Decision**: Use a single JetStream stream `GRIT` with subject
`grit.jobs.analysis`. Publisher uses message deduplication via NATS
`Nats-Msg-Id` header set to `{owner}/{repo}:{sha}` to prevent
duplicate jobs.

**Rationale**: NATS JetStream provides built-in message deduplication
(dedup window), durable consumers, and at-least-once delivery — all
needed for reliable job processing. The dedup window handles the
concurrent-request scenario natively.

**Alternatives considered**:
- Redis Streams: Would work but adds dual-purpose Redis usage (cache +
  queue), complicating failure modes.
- Simple NATS (no JetStream): No persistence, no dedup, no retry.
- In-process goroutine pool: Violates horizontal scalability principle.

**Implementation notes**:
- Stream config: `MaxAge: 1h`, `Subjects: ["grit.jobs.>"]`,
  `Retention: WorkQueuePolicy`, `Discard: DiscardOld`.
- Dedup window: 5 minutes (covers the max analysis time).
- Consumer: durable pull consumer `grit-worker`, `AckWait: 6m`,
  `MaxDeliver: 3`.
- Job payload: JSON `{ "owner": "...", "repo": "...", "sha": "...",
  "token": "..." (optional) }`.
- Job status stored in Redis: key `job:{job_id}`, TTL 1h.

## R5: Redis Caching Strategy

**Decision**: Cache the full `AnalysisResult` as JSON in Redis, keyed
by `{owner}/{repo}:{sha}:core` with 24-hour TTL.

**Rationale**: Storing the full result (not individual sub-components)
keeps the cache simple — a single GET returns everything needed for the
API response. The `:core` suffix follows the constitution's per-pillar
keying pattern.

**Alternatives considered**:
- Separate keys per sub-result (metadata, lines, commits): More
  granular invalidation but unnecessary complexity for the core pillar.
- msgpack/protobuf serialization: Faster but JSON is debuggable and
  the performance difference is negligible for this payload size.

**Implementation notes**:
- Job status keys: `job:{job_id}` → JSON `{ status, progress, result }`,
  TTL 1h.
- Active job dedup keys: `active:{owner}/{repo}:{sha}` → `{job_id}`,
  TTL 10m (cleared on completion).
- Cache check flow: handler checks `{owner}/{repo}:{sha}:core` →
  if HIT, return immediately; if MISS, check `active:*` for existing
  job → if found, return existing job_id; if not, publish new job.

## R6: Prometheus Metrics

**Decision**: Use `prometheus/client_golang` to expose a `/metrics`
endpoint with counters and histograms.

**Rationale**: Lightweight, standard, and compatible with any Prometheus
scraper or Grafana dashboard. No additional infrastructure required
beyond the existing Docker Compose stack.

**Implementation notes**:
- Counters: `grit_jobs_completed_total`, `grit_jobs_failed_total`,
  `grit_cache_hit_total`, `grit_cache_miss_total`,
  `grit_github_api_requests_total` (labeled by endpoint).
- Histograms: `grit_clone_duration_seconds`,
  `grit_analysis_duration_seconds`.
- Register on a separate `http.ServeMux` if needed, or mount on the
  chi router at `/metrics`.

## R7: Language Extension Map — Coverage

**Decision**: Ship a built-in map of 50+ language extensions in
`languages.go`. No external dependency for language detection.

**Rationale**: File-extension-based detection is simple, fast, and
matches ghloc-web's approach. Covering 50+ languages ensures the
SC-003 success criterion (≥40 languages) is met with margin.

**Languages covered** (non-exhaustive):
Go, Python, JavaScript, TypeScript, Java, Kotlin, C, C++, C#, Rust,
Swift, Ruby, PHP, Perl, Lua, R, Scala, Haskell, Elixir, Erlang,
Clojure, Dart, Zig, Nim, OCaml, F#, COBOL, Fortran, Assembly,
Shell/Bash, PowerShell, SQL, HTML, CSS, SCSS, LESS, XML, JSON, YAML,
TOML, Markdown, Dockerfile, Makefile, CMake, Gradle, Terraform, HCL,
Protobuf, GraphQL, Svelte, Vue.
