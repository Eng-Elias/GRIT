# Research: Contributor Attribution & Bus Factor Analysis

**Date**: 2026-04-04  
**Status**: Complete  
**Feature**: [spec.md](spec.md) | [plan.md](plan.md)

## R1: go-git Blame API

**Decision**: Use `go-git/go-git/v5/git.Blame()` to attribute lines per file.

**Rationale**: go-git is already a project dependency (used by core analysis for cloning and commit walking). Its `git.Blame()` function accepts a `*git.BlameOptions` with `Hash` (commit) and `Path` (file path) and returns a `*git.BlameResult` containing a slice of `git.Line` structs. Each `Line` has:
- `Author`: string (name)
- `AuthorMail`: string (email, angle-bracket-wrapped)  
- `Date`: `time.Time`
- `Hash`: `plumbing.Hash` (commit that last touched this line)
- `Text`: string (line content)

The email field needs angle-bracket stripping and case-normalization (`strings.ToLower`).

**Alternatives considered**:
- Shelling out to `git blame --porcelain`: Adds a process-per-file overhead and requires git binary on the Docker image. Rejected for consistency with existing pure-Go approach.

## R2: Goroutine Pool Design

**Decision**: Use a fixed-size worker pool with a buffered channel of file paths, `sync.WaitGroup` for completion, and `sync.Mutex`-protected result map. Pool size: `min(4, runtime.NumCPU())`.

**Rationale**: The existing complexity pillar uses a similar pool pattern (see `internal/analysis/complexity/pool.go`). A channel-based pool is idiomatic Go and avoids external dependencies. The mutex-protected map collects per-file blame results as they complete. On context cancellation (timeout), workers check `ctx.Err()` and stop, leaving partial results in the map.

**Alternatives considered**:
- `errgroup.Group` with `SetLimit()`: Simpler API but harder to collect partial results on timeout — `errgroup.Wait()` blocks until all goroutines finish or one errors. Our requirement is to snapshot whatever is done at timeout.
- `sync.Pool`: Wrong abstraction — `sync.Pool` is for object reuse, not work distribution.

## R3: Partial Results on Timeout

**Decision**: Use `context.WithTimeout(ctx, 10*time.Minute)` for the entire blame analysis. Workers check `ctx.Err()` before starting each file. When the context expires, the orchestrator snapshots the mutex-protected result map, computes bus factor from whatever lines were attributed, and stores the result with `Partial: true`.

**Rationale**: This matches the spec requirement (FR-020). The bus factor computed from partial data is still meaningful — it reflects the ownership distribution of the files that were processed. The `partial` flag lets consumers know the data may be incomplete.

**Alternatives considered**:
- Per-file timeout: Would add complexity without solving the global deadline requirement. A file that takes too long would still not be killed by a global timeout.

## R4: Pipeline Trigger — Core Worker (not Churn)

**Decision**: The blame job is enqueued by the **core analysis worker** (`internal/job/worker.go`) after core analysis completes, published alongside the complexity job. It runs **independent of** the complexity and churn pipelines.

**Rationale**: User explicitly specified this architecture. Blame only needs the file list from core analysis (to know which files to blame) and the cloned repository on disk. It does not depend on complexity or churn data. This allows blame, complexity, and churn to run in parallel after core completes, reducing total pipeline latency.

**Spec update needed**: The spec (FR-018) says "after churn analysis completes." This plan overrides that to "after core analysis completes" per user direction. The spec should be updated to reflect this.

## R5: Redis Key Pattern & TTL

**Decision**: 
- Data key: `{owner}/{repo}:{sha}:contributors` (TTL 48h)
- Active job key: `active:{owner}/{repo}:{sha}:blame` (TTL 10min)

**Rationale**: Follows existing key patterns: `{owner}/{repo}:{sha}:{pillar}` for data, `active:{owner}/{repo}:{sha}:{pillar}` for active job tracking. The 48h TTL matches the constitution's specification for contributor (blame) data.

## R6: Source File Extension Filtering

**Decision**: Reuse the complexity language registry (`internal/analysis/complexity/languages.go`) to determine which file extensions are "source code." Only files with recognized extensions are blamed.

**Rationale**: The spec (FR-004) says to use the same set as complexity analysis. This avoids maintaining a duplicate extension list and ensures consistency across pillars. The language registry already covers `.go`, `.py`, `.js`, `.ts`, `.java`, `.rb`, `.rs`, `.c`, `.cpp`.

**Implementation note**: The blame package will import a helper from complexity or, to avoid cross-pillar imports (constitution Principle II), the supported extensions set will be extracted to `internal/models/` or the blame package will define its own compatible set.

**Refined decision**: To comply with Principle II (no cross-pillar imports), define a `SupportedSourceExtensions` set in `internal/analysis/blame/extensions.go` that mirrors the complexity registry. This is a small static map and avoids coupling.
