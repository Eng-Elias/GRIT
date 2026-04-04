# Feature Specification: Contributor Attribution & Bus Factor Analysis

**Feature Branch**: `004-contributor-bus-factor`  
**Created**: 2026-04-04  
**Status**: Draft  
**Input**: User description: "Add contributor attribution and bus factor analysis to GRIT."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Per-Author Contributor Statistics (Priority: P1)

An engineering lead or developer requests contributor data for a repository. The system runs git blame on every source file at HEAD, attributes each line to an author by email (case-insensitive deduplication), and uses the most recent commit display name for that email as the canonical author name. The response includes per-author statistics: total lines owned, ownership percentage, number of files touched, first and last commit dates, top 3 primary languages by lines of code, and an active/inactive flag (active if last commit within the past 6 months).

**Why this priority**: Per-author ownership data is the foundational input for bus factor calculation and per-file breakdown. It also has standalone value — knowing who owns what code helps teams identify knowledge silos and plan reviews.

**Independent Test**: Request `GET /api/:owner/:repo/contributors` after the blame analysis completes. Verify the response contains an `authors` array where each entry has `email`, `name`, `total_lines_owned`, `ownership_percent`, `files_touched`, `first_commit_date`, `last_commit_date`, `primary_languages`, and `is_active`. Verify that ownership percentages sum to 100% (within rounding tolerance). Verify case-insensitive email deduplication (e.g., `Alice@example.com` and `alice@example.com` merge into one author).

**Acceptance Scenarios**:

1. **Given** a repository with 3 contributors, **When** the blame analysis completes, **Then** the `authors` array contains exactly 3 entries with correct line counts.
2. **Given** a contributor who committed using two email casings (`Alice@example.com` and `alice@example.com`), **When** the response is returned, **Then** a single author entry exists with lines from both casings combined.
3. **Given** a contributor whose display name changed across commits, **When** the response is returned, **Then** the `name` field reflects the display name from their most recent commit.
4. **Given** a contributor whose last commit was more than 6 months ago, **When** the response is returned, **Then** `is_active` is `false`.
5. **Given** a contributor who wrote lines in `.go`, `.py`, and `.js` files, **When** the response is returned, **Then** `primary_languages` lists the top 3 languages ordered by lines of code descending.

---

### User Story 2 — Bus Factor Calculation (Priority: P1)

The system computes the repository's bus factor — the minimum number of authors whose combined line ownership exceeds 80% of the total lines. Authors are ranked by `lines_owned` descending. The system walks the ranked list, accumulating ownership until the 80% threshold is crossed; the count of authors needed is the bus factor integer. The authors in that 80% set are returned as `key_people`. Bus factor and key people are included in the contributors endpoint response.

**Why this priority**: The bus factor is the primary deliverable — it directly answers "how many people could leave before this project is at risk?" It depends on per-author data (US1), making it P1 alongside contributor statistics.

**Independent Test**: Seed a fixture repository with known author line distributions. Request `GET /api/:owner/:repo/contributors`. Verify `bus_factor` is the correct minimum author count to reach 80% ownership. Verify `key_people` contains exactly those authors, ordered by `lines_owned` descending.

**Acceptance Scenarios**:

1. **Given** a repository where one author owns 85% of lines, **When** the response is returned, **Then** `bus_factor` is `1` and `key_people` contains only that author.
2. **Given** a repository where 3 authors each own 30%, 28%, and 25% of lines, **When** the response is returned, **Then** `bus_factor` is `3` (30 + 28 + 25 = 83% exceeds 80%).
3. **Given** a repository with perfectly equal distribution among 10 authors (10% each), **When** the response is returned, **Then** `bus_factor` is `8` (8 × 10% = 80%).
4. **Given** an empty repository with no blameworthy lines, **When** the response is returned, **Then** `bus_factor` is `0` and `key_people` is empty.

---

### User Story 3 — Per-File Contributor Breakdown (Priority: P2)

For each source file at HEAD, the system computes the top 3 authors by line ownership with their ownership percentage for that file. This data is available via a dedicated endpoint. It helps developers quickly see who to ask about any given file.

**Why this priority**: Per-file breakdown adds granularity on top of the repository-level view but is not essential for the bus factor calculation. It can be developed after the core contributor and bus factor data.

**Independent Test**: Request `GET /api/:owner/:repo/contributors/files`. Verify the response contains a `files` array with one entry per source file, each containing `path` and `top_authors` (up to 3 entries with `name`, `email`, `lines_owned`, and `ownership_percent`). Verify ownership percentages are relative to the individual file.

**Acceptance Scenarios**:

1. **Given** a file with 5 contributors, **When** the response is returned, **Then** only the top 3 by line ownership appear in `top_authors`.
2. **Given** a file with 1 contributor, **When** the response is returned, **Then** `top_authors` contains 1 entry with `ownership_percent` equal to 100%.
3. **Given** the response contains a file entry, **Then** each `top_authors` entry includes `name`, `email`, `lines_owned`, and `ownership_percent`.
4. **Given** a binary or non-source file, **When** the response is returned, **Then** that file is excluded from the `files` array.

---

### User Story 4 — Async Blame Job Pipeline (Priority: P1)

The blame analysis runs as an asynchronous job triggered after the core analysis completes, independent of complexity and churn pipelines. Results are cached with a 48-hour TTL. The contributors endpoint returns 200 with cached data, 202 if a job is in progress, or 404 if no analysis has been performed. The worker pool uses `min(4, NumCPU())` goroutines for parallel blame processing with a 10-minute hard timeout. If the timeout is reached, partial results are stored with a `partial: true` flag.

**Why this priority**: The async pipeline is essential infrastructure — without it the endpoints cannot serve data. It follows the same pattern as existing core, complexity, and churn job pipelines.

**Independent Test**: Trigger a core analysis for a repository. Verify that after core completes, a blame job is automatically enqueued (in parallel with complexity). Poll the contributors endpoint and verify it transitions from 202 (in progress) to 200 (complete). Verify cache hit on subsequent requests. Verify that a 10-minute timeout produces partial results with the `partial` flag.

**Acceptance Scenarios**:

1. **Given** a core analysis job has completed, **When** the core worker finishes, **Then** a blame job is automatically published to the NATS queue (alongside the complexity job).
2. **Given** blame analysis has completed and been cached, **When** `GET /api/:owner/:repo/contributors` is called, **Then** the response is 200 with `X-Cache: HIT`.
3. **Given** a blame job is currently running, **When** the endpoint is called, **Then** the response is 202 with job status.
4. **Given** no analysis has been performed for a repository, **When** the endpoint is called, **Then** the response is 404.
5. **Given** the same repository is requested twice while a job is active, **Then** the existing job ID is returned (deduplication).
6. **Given** blame processing exceeds the 10-minute hard timeout, **When** the timeout fires, **Then** partial results are stored with `partial: true` and the job completes with a success status.

---

### User Story 5 — Contributor Summary in Main Analysis (Priority: P2)

Once contributor data is available, a `contributor_summary` object is embedded in the main `GET /api/:owner/:repo` response. This summary includes the bus factor, top 3 contributors by line ownership, total author count, active author count, and a link to the full contributors endpoint.

**Why this priority**: Embedding a summary in the main response provides at-a-glance contributor health without requiring a separate request, but it depends on the full contributor pipeline being operational first.

**Independent Test**: Request `GET /api/:owner/:repo` after all analysis pipelines have completed. Verify the response contains a `contributor_summary` object with `bus_factor`, `top_contributors`, `total_authors`, `active_authors`, and `contributors_url`.

**Acceptance Scenarios**:

1. **Given** contributor analysis has completed, **When** `GET /api/:owner/:repo` is called, **Then** the response includes `contributor_summary` with correct values.
2. **Given** contributor analysis has NOT completed, **When** `GET /api/:owner/:repo` is called, **Then** `contributor_summary` is either absent or shows `status: "pending"`.
3. **Given** the response includes `contributor_summary`, **Then** `top_contributors` contains at most 3 entries sorted by `lines_owned` descending.

---

### Edge Cases

- What happens when a repository has zero blameworthy lines (empty repo or only binary files)? The `authors` array is empty, `bus_factor` is `0`, `key_people` is empty, and `files` is empty — the response is a valid empty structure.
- What happens when git blame encounters a binary file? Binary files are skipped — only source files with recognized extensions are blamed.
- What happens when a file was added via a merge commit with no individual author? Git blame still attributes each line to the commit that last touched it; merge commits that don't modify line content are transparent to blame.
- What happens when an author uses multiple email addresses (not just different casings)? Each distinct email (after case normalization) is treated as a separate author. Mailmap-style merging is out of scope for this feature.
- What happens when the repository has only one contributor? `bus_factor` is `1`, that contributor appears in `key_people`, and all files show a single author with 100% ownership.
- What happens when the blame job times out after 10 minutes? Partial results collected so far are stored with `partial: true`. The per-file breakdown may be incomplete. The bus factor is computed from whatever lines were attributed.
- What happens when a file has zero lines (empty source file)? It is skipped in the blame analysis — it contributes nothing to line counts or ownership.

## Requirements *(mandatory)*

### Functional Requirements

#### Blame Attribution

- **FR-001**: System MUST run git blame on every source file at HEAD to attribute each line to an author by email.
- **FR-002**: System MUST deduplicate authors by email using case-insensitive comparison (e.g., `Alice@example.com` and `alice@example.com` are the same author).
- **FR-003**: System MUST use the display name from the author's most recent commit as the canonical `name` for that email.
- **FR-004**: System MUST only blame files with source code extensions recognized by the language registry (same set used by complexity analysis), skipping binary and non-source files.

#### Per-Author Statistics

- **FR-005**: For each deduplicated author, the system MUST compute: `total_lines_owned`, `ownership_percent` (relative to total repository lines), `files_touched` (count of distinct files).
- **FR-006**: For each author, the system MUST record `first_commit_date` and `last_commit_date`.
- **FR-007**: For each author, the system MUST compute `primary_languages` — the top 3 languages by lines of code owned, identified by file extension.
- **FR-008**: For each author, the system MUST set `is_active` to `true` if their `last_commit_date` is within the past 6 months, `false` otherwise.

#### Bus Factor

- **FR-009**: System MUST rank authors by `lines_owned` descending and find the minimum count whose combined ownership exceeds 80% of total lines — this count is the `bus_factor` integer.
- **FR-010**: System MUST return the authors comprising the 80% threshold as `key_people`, ordered by `lines_owned` descending.
- **FR-011**: If there are zero blameworthy lines, `bus_factor` MUST be `0` and `key_people` MUST be empty.

#### Per-File Breakdown

- **FR-012**: For each source file at HEAD, the system MUST compute the top 3 authors by line ownership with their `name`, `email`, `lines_owned`, and `ownership_percent` (relative to that file).
- **FR-013**: Files with zero blameworthy lines MUST be excluded from the per-file breakdown.

#### Endpoints

- **FR-014**: `GET /api/:owner/:repo/contributors` MUST return the full contributor data including `authors`, `bus_factor`, `key_people`, repository metadata, and analysis timestamp.
- **FR-015**: `GET /api/:owner/:repo/contributors/files` MUST return the per-file contributor breakdown.
- **FR-016**: `GET /api/:owner/:repo` MUST include a `contributor_summary` object (bus factor, top 3 contributors, total/active author counts, contributors URL) once contributor data is available.
- **FR-017**: Contributor endpoints MUST return 200 with cached data and `X-Cache: HIT`, 202 with job status if a job is in progress, or 404 if no analysis exists.

#### Async Job Pipeline

- **FR-018**: System MUST publish a blame analysis job to a dedicated NATS subject (`grit.jobs.blame`) after core analysis completes, independent of complexity and churn pipelines.
- **FR-019**: The blame worker MUST use a pool of `min(4, NumCPU())` goroutines for parallel blame processing.
- **FR-020**: The blame worker MUST enforce a 10-minute hard timeout on the analysis. If the timeout is reached, partial results MUST be stored with `partial: true`.
- **FR-021**: System MUST cache blame analysis results with a 48-hour TTL.
- **FR-022**: System MUST deduplicate blame jobs — if an active job already exists for the same owner/repo/sha, return the existing job ID.
- **FR-023**: Cache deletion for a repository MUST clear blame/contributor cache entries alongside core, complexity, and churn cache entries.

### Key Entities

- **Author**: A deduplicated contributor — canonical email, display name, total lines owned, ownership percentage, files touched, first/last commit dates, primary languages, active flag.
- **BusFactor**: Repository-level metric — integer bus factor value, list of key people (Author references).
- **FileContributors**: Per-file ownership breakdown — file path, array of top 3 authors with per-file line counts and ownership percentages.
- **ContributorResult**: Complete blame analysis output — repository info, authors array, bus factor, key people, per-file breakdown, total lines analyzed, partial flag, analysis timestamp.
- **ContributorSummary**: Abbreviated contributor data embedded in main analysis response — bus factor, top 3 contributors, total author count, active author count, contributors URL, status.

### Assumptions

- The existing go-git clone infrastructure (used by core analysis) provides access to the repository at HEAD for blame operations.
- Core analysis completes before blame analysis is triggered, as the blame job is auto-published after the core worker finishes. Blame runs in parallel with complexity and churn.
- The source file extension list for blame reuses the same extensions supported by the complexity language registry.
- Git blame via go-git operates at the file level and returns line-by-line author attribution with email and timestamp.
- Author email deduplication is limited to case-insensitive normalization. Full mailmap support is out of scope.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Blame analysis for a repository with 500 source files completes within 10 minutes (the hard timeout).
- **SC-002**: The bus factor is correctly computed when compared against manually verified line counts on a known fixture repository.
- **SC-003**: `GET /api/:owner/:repo/contributors` returns a complete response (authors + bus factor + key people) in a single request.
- **SC-004**: `GET /api/:owner/:repo/contributors/files` returns per-file top-3 breakdowns for every source file at HEAD.
- **SC-005**: Author email deduplication correctly merges entries that differ only by case, verified against a fixture repository with intentional casing variations.
- **SC-006**: The blame job pipeline follows the same reliability patterns as core, complexity, and churn pipelines: deduplication, cache TTL, job status tracking, partial results on timeout, and automatic triggering.
- **SC-007**: All existing core, complexity, and churn tests continue to pass after contributor feature integration.
