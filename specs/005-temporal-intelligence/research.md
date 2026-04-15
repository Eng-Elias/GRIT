# Research: Temporal Intelligence

## R1 — Monthly LOC Snapshot via go-git

**Decision**: Use `git.Log` with `--first-parent` semantics (`git.LogOptions{From: HEAD, Order: LogOrderCommitterTime}`) to iterate commits chronologically. For each month boundary, find the latest commit on or before the 1st of that month. Then call `commit.Tree()` and walk `tree.Files().ForEach()` to count lines per file — reusing the existing `core.CountLines` pattern from `internal/analysis/core/walker.go`.

**Rationale**: The existing codebase already walks commit trees in-memory via go-git (`core.Walk`, `blame_worker.listSourceFiles`, `churn.readSourceContents`). No disk checkout is needed — `object.File.Reader()` provides file content directly from the git object store. This is the same approach the core walker uses.

**Alternatives considered**:
- **Full disk checkout per month**: Too slow and disk-heavy for 36 checkpoints. Rejected.
- **`git log --numstat` via exec**: Breaks `CGO_ENABLED=0` purity and is less portable. Rejected.
- **go-git `DiffTree` for LOC deltas**: Would accumulate errors over time; direct tree counting is more accurate for absolute LOC. Rejected for LOC (used for velocity instead).

**Key API calls**:
- `repo.Log(&git.LogOptions{From: hash, Order: git.LogOrderCommitterTime})` — iterate commits
- `commit.Tree()` → `tree.Files().ForEach(func(f *object.File))` — walk tree
- `f.Reader()` → `io.ReadAll()` → count newlines — count lines
- Language detection via `blame.LanguageForFile(path)` — reuse existing extension registry

## R2 — Weekly Velocity via Commit Stats

**Decision**: Walk the commit log once for the past 52 weeks. For each commit, use `commit.Stats()` to get per-file additions/deletions. Aggregate by ISO week number. This is the same API the churn pillar uses (`churn.WalkCommitLog` calls `c.Stats()`).

**Rationale**: `commit.Stats()` is already proven in the churn pillar. It computes diff stats by comparing a commit to its parent. Aggregating by week boundary (Monday-to-Sunday ISO weeks) gives consistent weekly buckets.

**Alternatives considered**:
- **`DiffTree` between week-boundary commits**: Would miss intermediate commits and only show net change between two points. Rejected — per-commit stats give accurate total additions/deletions.
- **Per-file tracking**: Unnecessary for weekly aggregates. Per-commit `Stats()` already provides file-level additions/deletions which we sum.

**Key API calls**:
- `commit.Stats()` → `[]object.FileStat{Name, Addition, Deletion}` — per-commit diff stats
- Aggregate `Addition`/`Deletion` into weekly buckets keyed by `time.ISOWeek()`

## R3 — Per-Author Weekly Activity

**Decision**: While walking the commit log for velocity (R2), also record per-author additions/deletions per week. Track author by `commit.Author.Email` (lowercase). After the walk, rank authors by total activity (additions + deletions) and keep the top 10.

**Rationale**: Piggybacks on the same commit log walk as velocity — zero extra cost. Author attribution uses the commit's author email, which is simpler and more reliable than blame-level attribution for velocity purposes.

**Alternatives considered**:
- **Separate walk for author activity**: Redundant. Rejected.
- **Blame-based author attribution**: Too expensive for velocity; blame is for ownership, not activity. Rejected.

## R4 — Rolling 4-Week Commit Cadence

**Decision**: Count commits per day during the 52-week walk. Then compute a sliding 4-week (28-day) window: for each window position, `total_commits_in_window / 28` gives `commits_per_day`. Windows slide by 1 week (7 days).

**Rationale**: A 4-week rolling average smooths out vacation weeks and holidays. Weekly sliding gives enough resolution without overwhelming the response.

**Alternatives considered**:
- **2-week window**: Too noisy. Rejected.
- **Monthly window**: Inconsistent lengths (28-31 days). Rejected.

## R5 — PR Merge Time via GitHub GraphQL

**Decision**: Use the GitHub GraphQL API (`https://api.github.com/graphql`) with a raw `net/http` POST request (no third-party GraphQL library). Query the last 100 merged PRs with cursor-based pagination (page size 25, up to 4 pages). Extract `createdAt` and `mergedAt` timestamps, compute duration, then calculate median, p75, and p95 percentiles.

**Rationale**: The existing `internal/github/client.go` already uses `net/http` with manual JSON handling. Adding a GraphQL method to the same `Client` struct keeps the pattern consistent. No new dependency needed — just a `POST` with a JSON body containing the GraphQL query. Cursor-based pagination is GitHub's standard for GraphQL.

**Alternatives considered**:
- **`github.com/shurcooL/graphql`**: Adds a dependency for a single query. Rejected to minimize deps.
- **GitHub REST API (`GET /repos/:owner/:repo/pulls?state=closed`)**: Requires multiple pages, doesn't distinguish merged from closed, less efficient. Rejected.
- **No PR data**: Would leave velocity incomplete. PR merge time is a key engineering metric. Rejected.

**GraphQL query**:
```graphql
query($owner: String!, $repo: String!, $cursor: String) {
  repository(owner: $owner, name: $repo) {
    pullRequests(states: MERGED, first: 25, after: $cursor, orderBy: {field: CREATED_AT, direction: DESC}) {
      pageInfo { hasNextPage endCursor }
      nodes { createdAt mergedAt }
    }
  }
}
```

**Graceful degradation**: If the token is empty, the GraphQL API is unreachable, or rate-limited, `pr_merge_time` is returned as `null`. The rest of the temporal result is unaffected.

## R6 — Refactor Detection Algorithm

**Decision**: After computing weekly velocity (R2), identify "refactor weeks" where: (1) net LOC change (additions − deletions) < 0, AND (2) commit count > median weekly commit count. Group consecutive refactor weeks into periods.

**Rationale**: Negative net LOC indicates code removal. Above-median commit activity distinguishes intentional refactoring from low-activity weeks where a single file deletion might cause negative LOC. This two-condition heuristic avoids false positives.

**Alternatives considered**:
- **Only negative net LOC**: Too many false positives (one file delete). Rejected.
- **Machine learning detection**: Overkill for a heuristic. Rejected.
- **Configurable thresholds**: Premature complexity. Fixed 50th percentile (median) is a reasonable default. Rejected for now.

## R7 — NATS Integration & Pipeline Trigger

**Decision**: Add `TemporalSubject = "grit.jobs.temporal"` to `publisher.go`. Add `PublishTemporal()` method following the exact pattern of `PublishBlame()`. The core worker publishes the temporal job alongside blame after core analysis completes. The existing NATS stream wildcard `grit.jobs.>` already covers the new subject.

**Rationale**: Identical pattern to blame, churn, and complexity pipelines. No stream configuration changes needed.

## R8 — Redis Caching

**Decision**: Cache key `{owner}/{repo}:{sha}:temporal`, TTL 12h (per constitution). Active job key `active:{owner}/{repo}:{sha}:temporal`. Same pattern as contributors/churn/complexity.

**Rationale**: Constitution mandates 12h TTL for temporal data. The existing cache pattern in `redis.go` is directly reusable.

## R9 — Period Parameter Handling

**Decision**: The `?period` query parameter (1y, 2y, 3y; default 3y) controls only the LOC over time window. It is parsed in the handler and passed to the analyzer. The 52-week velocity window is always fixed regardless of period. The period is stored in the cached result so the handler can validate cache relevance.

**Rationale**: LOC trends over different periods serve different audiences (quick check vs. long-term view). Velocity is always 1 year — consistent with typical sprint/quarter planning cycles.

**Note**: Each distinct period value produces a separate cache entry since the LOC data differs. Cache key includes the period: `{owner}/{repo}:{sha}:temporal:{period}`.
