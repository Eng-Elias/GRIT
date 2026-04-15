# Data Model: Temporal Intelligence

## Entities

### MonthlySnapshot

A point-in-time LOC measurement at a monthly boundary.

| Field | Type | Description |
|-------|------|-------------|
| date | time.Time | First day of the month (ISO 8601) |
| total_loc | int | Total lines of code at this checkpoint |
| by_language | map[string]int | LOC per language (top 10 + "Other") |

### WeeklyActivity

Aggregated additions and deletions for a single ISO week.

| Field | Type | Description |
|-------|------|-------------|
| week | string | ISO week identifier (e.g., "2026-W15") |
| week_start | time.Time | Monday of this ISO week |
| additions | int | Total lines added across all commits this week |
| deletions | int | Total lines deleted across all commits this week |
| commits | int | Number of commits this week |

### AuthorWeeklyActivity

Per-author weekly activity for a top contributor.

| Field | Type | Description |
|-------|------|-------------|
| email | string | Author email (lowercase) |
| name | string | Author display name (from most recent commit) |
| total_additions | int | Sum of additions over the full period |
| total_deletions | int | Sum of deletions over the full period |
| weeks | []WeeklyActivity | Per-week breakdown for this author |

### CommitCadence

A rolling 4-week window of commit frequency.

| Field | Type | Description |
|-------|------|-------------|
| window_start | time.Time | Start date of the 4-week window |
| window_end | time.Time | End date of the 4-week window |
| commits_per_day | float64 | Average commits per day in this window |

### PRMergeTime

Pull request merge time percentile statistics.

| Field | Type | Description |
|-------|------|-------------|
| median_hours | float64 | Median merge time in hours |
| p75_hours | float64 | 75th percentile merge time in hours |
| p95_hours | float64 | 95th percentile merge time in hours |
| sample_size | int | Number of merged PRs sampled (up to 100) |

### VelocityMetrics

Container for all velocity-related data.

| Field | Type | Description |
|-------|------|-------------|
| weekly_activity | []WeeklyActivity | 52 weekly entries |
| author_activity | []AuthorWeeklyActivity | Top 10 authors by activity |
| commit_cadence | []CommitCadence | Rolling 4-week windows |
| pr_merge_time | *PRMergeTime | PR merge percentiles (nullable) |

### RefactorPeriod

A contiguous span of refactor weeks.

| Field | Type | Description |
|-------|------|-------------|
| start | time.Time | Start date (Monday of first refactor week) |
| end | time.Time | End date (Sunday of last refactor week) |
| net_loc_change | int | Summed net LOC change (negative) |
| weeks | int | Number of consecutive refactor weeks |

### TemporalResult

Complete temporal analysis output cached in Redis.

| Field | Type | Description |
|-------|------|-------------|
| repository | Repository | Repository metadata (shared model) |
| period | string | Period analyzed ("1y", "2y", "3y") |
| loc_over_time | []MonthlySnapshot | Monthly LOC snapshots |
| velocity | VelocityMetrics | Velocity data container |
| refactor_periods | []RefactorPeriod | Detected refactor periods |
| total_months | int | Number of monthly snapshots |
| total_weeks | int | Number of weekly entries (up to 52) |
| analyzed_at | time.Time | Timestamp of analysis |
| cached | bool | Whether result was served from cache |
| cached_at | *time.Time | When the cache was populated |

### TemporalSummary

Abbreviated temporal data embedded in the main analysis response.

| Field | Type | Description |
|-------|------|-------------|
| status | string | "pending" or "complete" |
| current_loc | int | Total LOC at most recent monthly snapshot |
| loc_trend_6m_percent | float64 | LOC growth % over past 6 months |
| avg_weekly_commits | float64 | Average commits per week (past 52 weeks) |
| temporal_url | string | Link to full temporal endpoint |

## Relationships

```
TemporalResult
├── Repository (shared model from internal/models/repository.go)
├── []MonthlySnapshot (one per sampled month)
├── VelocityMetrics
│   ├── []WeeklyActivity (one per week, up to 52)
│   ├── []AuthorWeeklyActivity (top 10 authors)
│   │   └── []WeeklyActivity (per-author per-week)
│   ├── []CommitCadence (rolling 4-week windows)
│   └── *PRMergeTime (nullable, from GitHub GraphQL)
└── []RefactorPeriod (derived from WeeklyActivity data)
```

## State Transitions (Temporal Job)

```
[no job] → queued → running → completed
                           → failed
```

- **queued**: Job published to `grit.jobs.temporal`, active job key set in Redis
- **running**: Worker picked up the message, processing commit log
- **completed**: Results cached at `{owner}/{repo}:{sha}:temporal:{period}`, active job key deleted
- **failed**: Error logged, job status updated, active job key deleted

## Cache States

| State | Endpoint Response |
|-------|------------------|
| Cache HIT | 200 OK + `X-Cache: HIT` + full TemporalResult JSON |
| Job active | 202 Accepted + `{job_id, status, poll_url}` |
| No analysis | 404 Not Found |
| Invalid period | 400 Bad Request |

## Redis Key Patterns

| Key | Value | TTL |
|-----|-------|-----|
| `{owner}/{repo}:{sha}:temporal:{period}` | JSON TemporalResult | 12h |
| `active:{owner}/{repo}:{sha}:temporal` | Job ID string | 30m |
| `job:{job_id}` | JSON AnalysisJob | 1h |
