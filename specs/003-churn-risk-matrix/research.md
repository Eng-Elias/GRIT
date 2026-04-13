# Research: Git Churn Analysis, Risk Matrix & Dead Code Estimation

**Feature**: 003-churn-risk-matrix
**Date**: 2026-04-01

## R1: Commit Log Walking with go-git

**Decision**: Use go-git's `repository.Log()` returning a `CommitIterator` to walk the default branch. Build `map[string]int` from each commit's `Stats()` diff.

**Rationale**:
- go-git v5 is already a project dependency used by core analysis for cloning
- `repo.Log(&git.LogOptions{From: headRef})` iterates commits in reverse chronological order
- Each `commit.Stats()` returns `[]FileStat` with `Name` field — increment `churnMap[stat.Name]++`
- Single-pass iteration: O(commits × avg_files_per_commit) — no need for multiple passes
- `CommitIterator.ForEach()` handles traversal; break early once the commit window cap is reached

**Commit window logic**:
```
cutoffDate = now - 2 years
maxCommits = 5000
commitCount = 0

for each commit in log (newest first):
    if commitCount >= maxCommits: break
    if commit.Author.When.Before(cutoffDate): break
    for each fileStat in commit.Stats():
        churnMap[fileStat.Name]++
    commitCount++
```

**Alternatives considered**:
- `git log --numstat` via exec: Requires git binary in Docker image, parsing stdout. Rejected — go-git provides native Go API without external process.
- `commit.Patch()` for full diffs: More expensive than `Stats()` which only provides file-level summary. Rejected for performance.

## R2: Default Branch Detection

**Decision**: Use `repo.Head()` to get the HEAD reference of the cloned repository, which points to the default branch after a standard clone.

**Rationale**:
- Core analysis already clones the repo with go-git; HEAD points to the default branch
- No need to query GitHub API for default branch name — the clone's HEAD is authoritative
- If needed as fallback: `repo.References()` can find `refs/remotes/origin/HEAD`

## R3: Risk Level Classification

**Decision**: Sort-based percentile calculation in Go. Join churn and complexity data by file path. Classify using p50, p75, p90 thresholds.

**Rationale**:
- Percentiles require sorted data — `sort.Float64s()` on the value arrays, then index-based lookup
- p50 = value at index `len/2`, p75 = value at `len*3/4`, p90 = value at `len*9/10`
- For small arrays (< 4 elements), degenerate gracefully: all values are the same percentile

**Classification rules** (evaluated in order):
1. `critical`: churn > churn_p75 AND complexity > complexity_p75
2. `high`: churn > churn_p90 OR complexity > complexity_p90
3. `medium`: churn > churn_p50 OR complexity > complexity_p50
4. `low`: everything else

**Risk zone**: Filter critical entries, sort by `churn * complexity` descending.

**Data join**: Complexity data comes from Redis cache (`GetComplexity`), unmarshaled to `ComplexityResult`. Iterate `ComplexityResult.Files` to build `map[string]int` of path → cyclomatic. Join with churn `map[string]int` on path.

**Alternatives considered**:
- Statistical library for percentiles: Unnecessary overhead for a simple sorted-index calculation. Rejected.
- Weighted risk scoring: More complex, harder to explain to users. The percentile-based approach is transparent and the thresholds are returned in the response for frontend use.

## R4: Dead Code (Stale File) Detection

**Decision**: Three-condition heuristic: no recent commits + source extension + no import references.

**Rationale**:
- Heuristic approach is explicitly stated in the spec — no full dependency resolution needed
- Condition 1 (recency): Files with zero churn entries within the past 6 months. Computed from the commit log by tracking `lastModified[path] = commit.Author.When` for each file.
- Condition 2 (source extension): Allowlist of source code extensions. Reuse complexity language registry extensions plus `.sh`, `.sql`, `.proto`, `.graphql`. Excludelist: `.md`, `.txt`, `.yaml`, `.yml`, `.json`, `.toml`, `.xml`, `.csv`, `.svg`, `.png`, `.jpg`, `.gif`, `.ico`, `.woff`, `.ttf`, `.lock`, `.sum`, `.mod`.
- Condition 3 (import scan): For each stale candidate, check if `filepath.Base(path)` (without extension) appears as a substring in any source file's content. Read file contents at HEAD via `worktree.Filesystem`.

**Import scan implementation**:
```
for each staleCandidate:
    baseName = strings.TrimSuffix(filepath.Base(candidate.Path), filepath.Ext(candidate.Path))
    for each sourceFile in repo:
        content = readFile(sourceFile)
        if strings.Contains(content, baseName):
            mark candidate as referenced; break
```

**Performance**: The import scan is O(stale_candidates × source_files × avg_file_size). For most repos this is small because stale candidates are typically few. For very large repos, the source file contents can be read lazily and cached in memory during the scan.

**Alternatives considered**:
- AST-based import resolution: Accurate but language-specific and vastly more complex. Rejected per spec's explicit heuristic approach.
- Git blame for recency: More granular (line-level) but much slower. File-level commit recency is sufficient for this heuristic.

## R5: NATS Subject and Auto-Trigger

**Decision**: Use NATS subject `grit.jobs.churn` on the existing `GRIT` stream. Complexity worker publishes churn job after successful complexity completion.

**Rationale**:
- The `GRIT` stream already uses `grit.jobs.>` wildcard — covers `grit.jobs.churn` automatically
- Separate subject allows independent consumer group: `PullSubscribe("grit.jobs.churn", "grit-churn-worker")`
- Auto-trigger from complexity worker mirrors how core worker triggers complexity
- 30-second requeue: If complexity cache not found, `msg.NakWithDelay(30 * time.Second)` requeues the message

**Churn job payload**: Same structure as existing `JobPayload` — `job_id`, `owner`, `repo`, `sha`, `token`.

## R6: Cache Key Strategy

**Decision**: Separate Redis key `{owner}/{repo}:{sha}:churn` with 24h TTL (per Constitution Principle IV).

**Rationale**:
- Separate from core (`:core`) and complexity (`:complexity`) per pillar isolation
- 24h TTL matches constitution's churn pillar TTL
- `findChurn` scan pattern: `{owner}/{repo}:*:churn` for SHA-unknown lookups
- `DeleteChurn` follows same SCAN-based pattern as `DeleteAnalysis` and `DeleteComplexity`
- Active job key: `active:{owner}/{repo}:{sha}:churn`

## R7: Churn Summary Embedding

**Decision**: When the analysis handler returns a cached core result, also look up the churn cache. If found, attach a `churn_summary` field to the response.

**Rationale**:
- Same pattern as `complexity_summary` — consistent API experience
- Summary is lightweight: status, total files analyzed, critical file count, stale file count, churn matrix URL
- If churn cache miss: embed `churn_summary: { "status": "pending" }` to indicate not yet available
- Full churn data remains at the dedicated `/churn-matrix` endpoint

## R8: Months Inactive Calculation

**Decision**: `months_inactive = int(time.Since(lastModified).Hours() / (24 * 30))`

**Rationale**:
- Approximate month calculation using 30-day months is sufficient for this heuristic
- Consistent with the spec's "months_inactive" integer field
- No need for calendar-aware month counting — the precision difference is negligible for a staleness heuristic
