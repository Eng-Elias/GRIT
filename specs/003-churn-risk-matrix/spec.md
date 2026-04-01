# Feature Specification: Git Churn Analysis, Risk Matrix & Dead Code Estimation

**Feature Branch**: `003-churn-risk-matrix`
**Created**: 2026-04-01
**Status**: Draft
**Input**: User description: "Add Git churn analysis, a risk matrix, and dead code estimation to GRIT. This answers: which files should I be most worried about?"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Per-File Churn Scores (Priority: P1)

A developer or engineering lead requests churn data for a repository. The system walks the default branch commit log and counts how many commits modified each file, producing a churn score per file. The commit window is capped at the most recent 2 years or 5,000 commits (whichever is smaller) to keep analysis fast and relevant.

**Why this priority**: Churn data is the foundational input for the risk matrix and stale-file detection. Without it, neither downstream feature can operate. It also has standalone value — knowing which files change most frequently highlights maintenance hotspots.

**Independent Test**: Request `GET /api/:owner/:repo/churn-matrix` after core + complexity analysis have completed. Verify the response contains a `churn` array with one entry per file that was modified in the commit window, each with a `path` and integer `churn` score. Verify the commit window respects the 2-year / 5,000-commit cap.

**Acceptance Scenarios**:

1. **Given** a repository with 200 commits on its default branch, **When** the churn analysis completes, **Then** every file modified in those 200 commits appears in the `churn` array with the correct modification count.
2. **Given** a repository with 8,000 commits spanning 3 years, **When** the churn analysis runs, **Then** only the most recent 5,000 commits are considered (the 5,000-commit cap is reached before the 2-year cap).
3. **Given** a repository with 1,000 commits spanning 6 months, **When** the churn analysis runs, **Then** all 1,000 commits are considered (both caps are larger than the actual data).
4. **Given** a repository with multiple branches, **When** the churn analysis runs, **Then** only commits reachable from the default branch are counted.

---

### User Story 2 — Risk Matrix (Priority: P1)

For every file at HEAD that has both a churn score and complexity data, the system computes a risk level by combining churn frequency with cyclomatic complexity. Files are classified into risk levels (critical, high, medium, low) using percentile thresholds. The top critical files are returned sorted by `churn × complexity` descending as a "risk zone." The percentile thresholds (p50, p75, p90) are included in the response so frontends can draw reference lines on scatter plots.

**Why this priority**: The risk matrix is the primary deliverable — it directly answers "which files should I be most worried about?" It depends on churn data (US1) and pre-existing complexity cache, making it P1 alongside churn.

**Independent Test**: Seed a repository with known churn and complexity values. Request `GET /api/:owner/:repo/churn-matrix`. Verify the `risk_matrix` array contains correctly classified entries and `risk_zone` is sorted by `churn * complexity` descending. Verify `thresholds` contains p50, p75, p90 for both churn and complexity.

**Acceptance Scenarios**:

1. **Given** files where both churn and complexity exceed p75, **When** the response is returned, **Then** those files have `risk_level: "critical"`.
2. **Given** a file where churn exceeds p90 but complexity is below p50, **When** the response is returned, **Then** that file has `risk_level: "high"`.
3. **Given** a file where both churn and complexity are below p50, **When** the response is returned, **Then** that file has `risk_level: "low"`.
4. **Given** the response is returned, **Then** the `risk_zone` array contains only critical-level files, sorted by `churn * complexity` descending.
5. **Given** the response is returned, **Then** `thresholds` includes `churn_p50`, `churn_p75`, `churn_p90`, `complexity_p50`, `complexity_p75`, `complexity_p90`.

---

### User Story 3 — Dead Code Estimation (Priority: P2)

The system flags files as "potentially unused" (stale) when all of the following are true: the file has zero commits in the past 6 months, the file has a source code extension (not config, documentation, or asset files), and the filename does not appear in any import or require statement across the repository (heuristic string scan). Results are returned as `stale_files` in the churn-matrix response.

**Why this priority**: Dead code detection provides additional housekeeping value but is heuristic-based and less critical than the core risk matrix. It depends on the churn data from US1 and can be developed after the risk matrix.

**Independent Test**: Create a fixture repository containing a source file last modified more than 6 months ago that is never imported by any other file. Request `GET /api/:owner/:repo/churn-matrix`. Verify that file appears in `stale_files` with correct `path`, `last_modified`, and `months_inactive` fields. Verify config files, docs, and imported files are excluded.

**Acceptance Scenarios**:

1. **Given** a `.go` file with no commits in the past 6 months and no import references, **When** the response is returned, **Then** it appears in `stale_files`.
2. **Given** a `.go` file with no recent commits but referenced by an import in another file, **When** the response is returned, **Then** it is NOT in `stale_files`.
3. **Given** a `.md` file with no commits in the past 6 months, **When** the response is returned, **Then** it is NOT in `stale_files` (documentation is excluded).
4. **Given** a `.yaml` config file with no recent commits, **When** the response is returned, **Then** it is NOT in `stale_files`.
5. **Given** the response, **Then** each stale file entry includes `path`, `last_modified` (date of last commit), and `months_inactive` (integer count).

---

### User Story 4 — Async Churn Job Pipeline (Priority: P1)

The churn analysis runs as an asynchronous job triggered after complexity analysis completes. If complexity data is not yet available when the churn job starts, it requeues with a 30-second delay. Results are cached with a 24-hour TTL. The endpoint returns 200 with cached data, 202 if a job is in progress, or 404 if no analysis has been performed.

**Why this priority**: The async pipeline is essential infrastructure — without it the endpoint cannot serve data. It follows the same pattern as existing core and complexity job pipelines.

**Independent Test**: Trigger a core analysis for a repository. Verify that after complexity completes, a churn job is automatically enqueued. Poll the churn-matrix endpoint and verify it transitions from 202 (in progress) to 200 (complete). Verify cache hit on subsequent requests.

**Acceptance Scenarios**:

1. **Given** a complexity job has completed, **When** the complexity worker finishes, **Then** a churn job is automatically published.
2. **Given** a churn job starts but complexity data is not yet cached, **When** the worker checks for complexity data, **Then** the job is requeued with a 30-second delay.
3. **Given** churn analysis has completed and been cached, **When** `GET /api/:owner/:repo/churn-matrix` is called, **Then** the response is 200 with `X-Cache: HIT`.
4. **Given** a churn job is currently running, **When** the endpoint is called, **Then** the response is 202 with job status.
5. **Given** no analysis has been performed for a repository, **When** the endpoint is called, **Then** the response is 404.
6. **Given** the same repository is requested twice while a job is active, **Then** the existing job ID is returned (deduplication).

---

### Edge Cases

- What happens when a repository has zero commits (empty repo)? The churn array is empty, risk matrix is empty, stale files is empty — the response is a valid empty structure with zero counts.
- What happens when the repository has no supported-language files (only config/docs)? The risk matrix is empty (no complexity data to join). Stale files is also empty (no source files). Churn data still reflects all files.
- What happens when complexity analysis is still running when a user requests the churn endpoint? Return 202 with job status if the churn job is queued/running, or 404 if no churn job has been created yet.
- What happens if a file was renamed during the commit window? Each path is treated independently — the old name accumulates churn up to the rename, the new name accumulates from the rename onward.
- What happens when the commit log contains merge commits? Merge commits are included; each file touched in a merge commit counts as one modification.
- What happens when the repository has only one commit? The single commit's changed files each get churn score 1. Risk percentiles degenerate gracefully.
- What happens when the stale-file heuristic has a false positive (file is used but not via import strings)? The feature labels results as "potentially unused" — it is an estimation, not a guarantee. The response should make the heuristic nature clear.

## Requirements *(mandatory)*

### Functional Requirements

#### Churn Analysis

- **FR-001**: System MUST walk the default branch commit log and count the number of commits where each file was modified, producing a per-file churn score.
- **FR-002**: System MUST limit the commit window to the most recent 2 years OR 5,000 commits, whichever yields fewer commits.
- **FR-003**: System MUST only consider commits reachable from the default branch (no feature branches, no stale refs).
- **FR-004**: System MUST treat each file path independently — renamed files accumulate churn under both old and new paths separately.

#### Risk Matrix

- **FR-005**: System MUST compute a risk level for every file at HEAD that has both churn data and complexity data, using the formula: join on file path between churn scores and per-file cyclomatic complexity.
- **FR-006**: System MUST classify risk levels using percentile thresholds:
  - `critical`: both churn AND complexity above p75
  - `high`: either churn OR complexity above p90
  - `medium`: either churn OR complexity above p50
  - `low`: everything else
- **FR-007**: System MUST return a `risk_zone` array containing only critical-level files, sorted by `churn × complexity` descending.
- **FR-008**: System MUST include percentile thresholds in the response: `churn_p50`, `churn_p75`, `churn_p90`, `complexity_p50`, `complexity_p75`, `complexity_p90`.
- **FR-009**: Each risk matrix entry MUST include: `path`, `churn`, `complexity_cyclomatic`, `language`, `loc`, `risk_level`.

#### Dead Code Estimation

- **FR-010**: System MUST flag a file as "potentially unused" (stale) only when ALL three conditions are met: (a) zero commits in the past 6 months, (b) file has a source code extension, (c) filename does not appear in any import/require string in the repository.
- **FR-011**: System MUST exclude non-source files (config, documentation, assets, build artifacts) from stale file detection using a defined set of excluded extensions.
- **FR-012**: The import/require scan MUST use a simple string-contains heuristic (scanning file contents for the basename), not full dependency resolution.
- **FR-013**: Each stale file entry MUST include: `path`, `last_modified` (date of last commit touching the file), `months_inactive` (integer months since last commit).

#### Async Job Pipeline

- **FR-014**: System MUST publish a churn analysis job to a dedicated NATS subject after complexity analysis completes.
- **FR-015**: If complexity data is not yet cached when the churn worker starts, the worker MUST requeue the job with a 30-second delay.
- **FR-016**: System MUST cache churn analysis results with a 24-hour TTL.
- **FR-017**: System MUST deduplicate churn jobs — if an active job already exists for the same owner/repo/sha, return the existing job ID.
- **FR-018**: The `GET /api/:owner/:repo/churn-matrix` endpoint MUST return 200 with cached data and `X-Cache: HIT`, 202 with job status if a job is in progress, or 404 if no analysis exists.
- **FR-019**: System MUST embed a `churn_summary` in the main analysis response (similar to `complexity_summary`) with status and key aggregate fields.
- **FR-020**: Cache deletion for a repository MUST clear churn cache entries alongside core and complexity cache entries.

### Key Entities

- **FileChurn**: Per-file churn metric — path, churn score (commit count), last modified date.
- **RiskEntry**: Per-file risk classification — path, churn, cyclomatic complexity, language, LOC, computed risk level.
- **StaleFile**: Potentially unused file — path, last modified date, months inactive.
- **ChurnMatrixResult**: Complete churn analysis output — repository info, churn array, risk matrix, risk zone, thresholds, stale files, aggregate stats, timestamp.
- **ChurnSummary**: Abbreviated churn data embedded in main analysis response — status, total files analyzed, critical file count, stale file count, churn matrix URL.

### Assumptions

- The existing go-git clone infrastructure (used by core analysis) provides access to the default branch commit log via `CommitIterator`.
- Complexity data is available in Redis cache from the complexity analysis pipeline before churn analysis needs it for the risk matrix.
- The "source code extension" list for stale file detection reuses the same extensions supported by the complexity language registry, plus common extensions like `.sh`, `.sql`, `.proto`.
- The string-contains import scan reads files at HEAD — it does not need to parse AST or resolve transitive dependencies.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Churn analysis for a repository with 5,000 commits completes within 60 seconds.
- **SC-002**: The risk matrix correctly classifies 100% of files when compared against manually computed percentile thresholds on a known fixture repository.
- **SC-003**: The `GET /api/:owner/:repo/churn-matrix` endpoint returns a complete response (churn + risk matrix + stale files) in a single request.
- **SC-004**: Dead code estimation correctly identifies files with zero recent commits that are not imported, and excludes all config/doc/asset files, verified against a fixture repository.
- **SC-005**: The churn job pipeline follows the same reliability patterns as core and complexity pipelines: deduplication, cache TTL, job status tracking, and automatic triggering.
- **SC-006**: All existing core and complexity tests continue to pass after churn feature integration.
