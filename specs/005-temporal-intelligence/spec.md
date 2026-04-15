# Feature Specification: Temporal Intelligence

**Feature Branch**: `005-temporal-intelligence`
**Created**: 2026-04-14
**Status**: Draft
**Input**: User description: "Add temporal analysis tracking repository evolution over time with LOC over time, velocity metrics, and refactor detection."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — LOC Over Time (Priority: P1)

An engineering lead wants to understand how a repository has grown over time. The system samples monthly boundary commits going back up to 3 years (or the full history if shorter). At each monthly checkpoint, the system resolves the HEAD commit at that point, walks its tree in memory (no disk checkout), and counts total lines of code plus lines per language for the top 10 languages (all others grouped as "Other"). The response is a chronological array of monthly snapshots.

**Why this priority**: LOC over time is the foundational temporal metric. It provides the baseline growth trajectory that velocity and refactor detection build upon and has standalone value for understanding project scale trends.

**Independent Test**: Request `GET /api/:owner/:repo/temporal` after analysis completes. Verify the response contains a `loc_over_time` array where each entry has `date` (ISO 8601), `total_loc` (integer), and `by_language` (map of language name to line count). Verify entries are sorted chronologically. Verify top 10 languages are broken out with the rest grouped as "Other".

**Acceptance Scenarios**:

1. **Given** a repository with 2 years of history, **When** temporal analysis completes, **Then** `loc_over_time` contains 24 monthly entries sorted by date ascending.
2. **Given** a repository with 5 years of history and `?period=3y` (default), **When** the response is returned, **Then** `loc_over_time` contains 36 entries covering the most recent 3 years.
3. **Given** a repository with 6 months of history, **When** the response is returned, **Then** `loc_over_time` contains 6 entries.
4. **Given** a monthly checkpoint where the repository has 15 distinct languages, **When** the entry is returned, **Then** `by_language` contains the top 10 languages by LOC plus an "Other" entry summing the remaining 5.
5. **Given** a month where no commit exists (repository not yet created), **Then** that month is omitted from the array.

---

### User Story 2 — Velocity Metrics (Priority: P1)

A team lead wants to measure development velocity over the past year. The system computes weekly additions and deletions for the past 52 weeks using commit diff stats. It also computes per-author weekly activity for the top 10 contributors ranked by total activity in the past year. Additionally, it calculates a rolling 4-week commit cadence (average commits per day per window). Finally, it retrieves PR merge time statistics (median, p75, p95) from the last 100 merged pull requests via the GitHub GraphQL API.

**Why this priority**: Velocity metrics answer "how fast are we moving?" — critical for engineering managers and project planning. Weekly granularity provides actionable insight into development pace.

**Independent Test**: Request `GET /api/:owner/:repo/temporal`. Verify the response contains `velocity` with: `weekly_activity` (52-entry array of week/additions/deletions), `author_activity` (up to 10 authors with per-week breakdowns), `commit_cadence` (rolling 4-week windows with commits_per_day), and `pr_merge_time` (median/p75/p95 hours and sample_size).

**Acceptance Scenarios**:

1. **Given** a repository with 1 year of commit history, **When** temporal analysis completes, **Then** `weekly_activity` contains 52 entries with non-negative additions and deletions per week.
2. **Given** a repository with 20 active contributors in the past year, **When** the response is returned, **Then** `author_activity` contains the top 10 by total additions+deletions.
3. **Given** 52 weeks of commit history, **When** the response is returned, **Then** `commit_cadence` contains rolling 4-week windows covering the full period.
4. **Given** a public repository with merged PRs, **When** the response is returned, **Then** `pr_merge_time` contains `median_hours`, `p75_hours`, `p95_hours`, and `sample_size` (up to 100).
5. **Given** a repository with no merged PRs, **When** the response is returned, **Then** `pr_merge_time` is `null`.
6. **Given** a week with zero commits, **When** the response is returned, **Then** that week's entry shows `additions: 0, deletions: 0`.

---

### User Story 3 — Refactor Detection (Priority: P2)

A developer or tech lead wants to identify periods of significant refactoring. The system detects weeks where net LOC change is negative AND the commit count exceeds the median weekly commit count. Consecutive refactor weeks are grouped into refactor periods with start date, end date, and net LOC change. This helps teams visualize cleanup efforts and correlate them with velocity dips.

**Why this priority**: Refactor detection builds on top of the weekly velocity data (US2) and adds interpretive value. It is not required for the core temporal view but adds meaningful insight.

**Independent Test**: Request `GET /api/:owner/:repo/temporal`. Verify the response contains `refactor_periods` — an array of objects each with `start` (ISO 8601), `end` (ISO 8601), and `net_loc_change` (negative integer). Verify that each period corresponds to consecutive weeks meeting the refactor criteria.

**Acceptance Scenarios**:

1. **Given** a 3-week stretch where each week has negative net LOC and above-median commit counts, **When** the response is returned, **Then** `refactor_periods` contains one entry spanning those 3 weeks with the summed `net_loc_change`.
2. **Given** two separate refactor weeks separated by a non-refactor week, **When** the response is returned, **Then** `refactor_periods` contains two separate entries.
3. **Given** a repository with no weeks meeting the refactor criteria, **When** the response is returned, **Then** `refactor_periods` is an empty array.
4. **Given** a week with negative net LOC but below-median commit count, **Then** that week is NOT classified as a refactor week.

---

### User Story 4 — Async Temporal Job Pipeline (Priority: P1)

The temporal analysis runs as an asynchronous job triggered after core analysis completes, independent of complexity, churn, and blame pipelines. Results are cached with a 12-hour TTL. The temporal endpoint returns 200 with cached data, 202 if a job is in progress, or 404 if no analysis has been performed. The worker processes monthly checkpoints in a single sequential pass through the commit log.

**Why this priority**: The async pipeline is essential infrastructure — without it the endpoint cannot serve data. It follows the same pattern as existing core, complexity, churn, and blame job pipelines.

**Independent Test**: Trigger a core analysis for a repository. Verify that after core completes, a temporal job is automatically enqueued. Poll the temporal endpoint and verify it transitions from 202 to 200. Verify cache hit on subsequent requests.

**Acceptance Scenarios**:

1. **Given** a core analysis job has completed, **When** the core worker finishes, **Then** a temporal job is automatically published to the NATS queue alongside complexity and blame jobs.
2. **Given** temporal analysis has completed and been cached, **When** `GET /api/:owner/:repo/temporal` is called, **Then** the response is 200 with `X-Cache: HIT`.
3. **Given** a temporal job is currently running, **When** the endpoint is called, **Then** the response is 202 with job status.
4. **Given** no analysis has been performed for a repository, **When** the endpoint is called, **Then** the response is 404.
5. **Given** the same repository is requested twice while a job is active, **Then** the existing job ID is returned (deduplication).

---

### User Story 5 — Temporal Summary in Main Analysis (Priority: P2)

Once temporal data is available, a `temporal_summary` object is embedded in the main `GET /api/:owner/:repo` response. This summary includes the current month's total LOC, LOC trend (growth percentage over past 6 months), weekly commit average, and a link to the full temporal endpoint.

**Why this priority**: Embedding a summary in the main response provides at-a-glance temporal health without requiring a separate request, but it depends on the full temporal pipeline being operational.

**Independent Test**: Request `GET /api/:owner/:repo` after all pipelines complete. Verify the response contains a `temporal_summary` object with `status`, `current_loc`, `loc_trend_6m_percent`, `avg_weekly_commits`, and `temporal_url`.

**Acceptance Scenarios**:

1. **Given** temporal analysis has completed, **When** `GET /api/:owner/:repo` is called, **Then** the response includes `temporal_summary` with correct values.
2. **Given** temporal analysis has NOT completed, **When** `GET /api/:owner/:repo` is called, **Then** `temporal_summary` shows `status: "pending"`.
3. **Given** a repository with 6 months of LOC data, **Then** `loc_trend_6m_percent` reflects the percentage change from 6 months ago to now.

---

### Edge Cases

- What happens when a repository has fewer than 1 month of history? The `loc_over_time` array contains a single entry for the current month. Velocity covers only the available weeks.
- What happens when a monthly boundary has no commit? That month is skipped — only months with at least one commit before the boundary are sampled.
- What happens when `?period=1y` is specified? The system samples the most recent 12 months instead of 36.
- What happens when the GitHub GraphQL API is unavailable or rate-limited? `pr_merge_time` is returned as `null` with no error — the rest of the temporal data is still valid.
- What happens when all weeks have positive net LOC? `refactor_periods` is an empty array.
- What happens when a repository has no source files at an early monthly checkpoint? That checkpoint shows `total_loc: 0` and an empty `by_language` map.
- What happens when the period parameter is invalid (not 1y, 2y, or 3y)? The system returns a 400 Bad Request with an error message.

## Requirements *(mandatory)*

### Functional Requirements

#### LOC Over Time

- **FR-001**: System MUST sample monthly boundary commits going back up to the requested period (1y, 2y, or 3y; default 3y) or the full repository history if shorter.
- **FR-002**: At each monthly checkpoint, the system MUST resolve the latest commit on or before the first day of that month and count total lines of code by walking the commit tree in memory.
- **FR-003**: The system MUST break down LOC by language for the top 10 languages by line count at each checkpoint, grouping the remainder as "Other".
- **FR-004**: The `loc_over_time` array MUST be sorted chronologically (oldest first).

#### Velocity Metrics

- **FR-005**: System MUST compute weekly additions and deletions for the past 52 weeks by aggregating commit diff stats per week boundary.
- **FR-006**: System MUST compute per-author weekly activity for the top 10 contributors ranked by total additions+deletions in the past year.
- **FR-007**: System MUST compute a rolling 4-week commit cadence: for each 4-week window, the average number of commits per day.
- **FR-008**: System MUST retrieve PR merge time statistics (median, p75, p95) from the last 100 merged pull requests via the GitHub GraphQL API.
- **FR-009**: If the GitHub API is unavailable or returns an error, `pr_merge_time` MUST be `null` — the remaining temporal data MUST still be returned.

#### Refactor Detection

- **FR-010**: A "refactor week" is defined as a week where net LOC change (additions minus deletions) is negative AND the commit count exceeds the median weekly commit count over the analysis period.
- **FR-011**: Consecutive refactor weeks MUST be grouped into refactor periods with `start`, `end`, and `net_loc_change`.
- **FR-012**: If no weeks meet the refactor criteria, `refactor_periods` MUST be an empty array.

#### Endpoint

- **FR-013**: `GET /api/:owner/:repo/temporal` MUST return the full temporal data including `loc_over_time`, `velocity`, and `refactor_periods`.
- **FR-014**: The endpoint MUST accept an optional `?period=1y|2y|3y` query parameter (default `3y`) controlling the LOC time range.
- **FR-015**: The endpoint MUST return 200 with cached data and `X-Cache: HIT`, 202 with job status if a job is in progress, or 404 if no analysis exists.
- **FR-016**: `GET /api/:owner/:repo` MUST include a `temporal_summary` object once temporal data is available.
- **FR-017**: Invalid `period` values MUST return 400 Bad Request.

#### Async Job Pipeline

- **FR-018**: System MUST publish a temporal analysis job to a dedicated NATS subject (`grit.jobs.temporal`) after core analysis completes, independent of complexity, churn, and blame pipelines.
- **FR-019**: The temporal worker MUST process monthly checkpoints in a single sequential pass through the commit log (not parallelized — sequential to maintain chronological order).
- **FR-020**: System MUST cache temporal analysis results with a 12-hour TTL.
- **FR-021**: System MUST deduplicate temporal jobs — if an active job already exists for the same owner/repo/sha, return the existing job ID.
- **FR-022**: Cache deletion for a repository MUST clear temporal cache entries alongside other pillar cache entries.

### Key Entities

- **MonthlySnapshot**: A point-in-time measurement — date, total LOC, per-language LOC breakdown (top 10 + "Other").
- **WeeklyActivity**: Weekly aggregation — week identifier, total additions, total deletions.
- **AuthorWeeklyActivity**: Per-author weekly breakdown — author email/name, array of weekly addition/deletion entries.
- **CommitCadence**: Rolling 4-week window — window start, window end, average commits per day.
- **PRMergeTime**: Pull request merge time statistics — median hours, p75 hours, p95 hours, sample size.
- **RefactorPeriod**: A contiguous span of refactor weeks — start date, end date, net LOC change.
- **TemporalResult**: Complete temporal analysis output — repository info, loc_over_time array, velocity object, refactor_periods array, analysis timestamp.
- **TemporalSummary**: Abbreviated temporal data for the main analysis response — current LOC, 6-month LOC trend percentage, average weekly commits, temporal URL, status.

### Assumptions

- The existing go-git clone infrastructure provides access to the full commit history for walking monthly boundaries.
- Core analysis completes before temporal analysis is triggered, as the temporal job is auto-published after the core worker finishes. Temporal runs in parallel with complexity, churn, and blame.
- Language detection for LOC counting reuses the same source file extension registry used by the blame pillar.
- The GitHub GraphQL API is available for PR merge time data. A valid token is required; if unavailable, PR data is gracefully omitted.
- Walking a commit tree in memory (via go-git) to count lines is feasible without a full disk checkout.
- The 52-week velocity window is always relative to the analysis timestamp, not the period parameter. The period parameter only affects LOC over time.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: LOC over time data for a 3-year period on a repository with 1000 files completes within 5 minutes.
- **SC-002**: The `loc_over_time` array accurately reflects total LOC at each monthly boundary when compared against manual `cloc` counts on known fixture commits.
- **SC-003**: Weekly additions and deletions match the output of `git log --stat` aggregated by week on a fixture repository.
- **SC-004**: Refactor periods correctly identify consecutive weeks with negative net LOC and above-median commit activity on a fixture repository.
- **SC-005**: PR merge time percentiles are within 1% of manually computed values from the same 100 PRs.
- **SC-006**: `GET /api/:owner/:repo/temporal` returns a complete response in a single request after analysis completes.
- **SC-007**: The temporal job pipeline follows the same reliability patterns as other pillars: deduplication, cache TTL, job status tracking, and automatic triggering.
- **SC-008**: All existing core, complexity, churn, and contributor tests continue to pass after temporal feature integration.
