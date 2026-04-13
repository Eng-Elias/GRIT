# Data Model: Git Churn Analysis, Risk Matrix & Dead Code Estimation

**Feature**: 003-churn-risk-matrix
**Date**: 2026-04-01

## Entities

### FileChurn

Per-file churn metric derived from commit log analysis.

| Field         | Type      | Description                                          |
|---------------|-----------|------------------------------------------------------|
| path          | string    | File path relative to repository root                |
| churn         | int       | Number of commits that modified this file            |
| last_modified | time.Time | Date of the most recent commit touching this file    |

### RiskEntry

Per-file risk classification joining churn and complexity data.

| Field                  | Type   | Description                                        |
|------------------------|--------|----------------------------------------------------|
| path                   | string | File path relative to repository root              |
| churn                  | int    | Number of commits that modified this file           |
| complexity_cyclomatic  | int    | Cyclomatic complexity from complexity analysis      |
| language               | string | Programming language of the file                    |
| loc                    | int    | Lines of code                                       |
| risk_level             | string | Computed risk level: "critical", "high", "medium", "low" |

### StaleFile

Potentially unused file identified by the dead code heuristic.

| Field            | Type      | Description                                        |
|------------------|-----------|----------------------------------------------------|
| path             | string    | File path relative to repository root              |
| last_modified    | time.Time | Date of the most recent commit touching this file  |
| months_inactive  | int       | Number of months since last commit (approx 30-day) |

### Thresholds

Percentile thresholds for churn and complexity distributions.

| Field           | Type    | Description                           |
|-----------------|---------|---------------------------------------|
| churn_p50       | float64 | 50th percentile of churn scores       |
| churn_p75       | float64 | 75th percentile of churn scores       |
| churn_p90       | float64 | 90th percentile of churn scores       |
| complexity_p50  | float64 | 50th percentile of cyclomatic values  |
| complexity_p75  | float64 | 75th percentile of cyclomatic values  |
| complexity_p90  | float64 | 90th percentile of cyclomatic values  |

### ChurnMatrixResult

Complete churn analysis output for a repository. Top-level response entity.

| Field               | Type                 | Description                                              |
|---------------------|----------------------|----------------------------------------------------------|
| repository          | Repository           | Repository metadata (owner, name, full_name)             |
| churn               | []FileChurn          | Per-file churn scores for all files in commit window     |
| risk_matrix         | []RiskEntry          | Risk classification for files with both churn + complexity |
| risk_zone           | []RiskEntry          | Critical-level files sorted by churn×complexity desc     |
| thresholds          | Thresholds           | Percentile thresholds for frontend reference lines       |
| stale_files         | []StaleFile          | Files flagged as potentially unused                      |
| total_commits       | int                  | Total commits analyzed in the commit window              |
| commit_window_start | time.Time            | Oldest commit date in the analysis window                |
| commit_window_end   | time.Time            | Newest commit date in the analysis window                |
| total_files_churned | int                  | Number of unique files with churn > 0                    |
| critical_count      | int                  | Number of files classified as critical risk              |
| stale_count         | int                  | Number of files flagged as potentially unused            |
| analyzed_at         | time.Time            | Timestamp when analysis completed                        |
| cached              | bool                 | Whether this result was served from cache                |
| cached_at           | *time.Time           | When the result was cached (nil if not cached)           |

### ChurnSummary

Abbreviated churn data embedded in the main analysis response.

| Field            | Type   | Description                                           |
|------------------|--------|-------------------------------------------------------|
| status           | string | "complete", "pending", or "running"                   |
| total_files      | int    | Number of files with churn data                       |
| critical_count   | int    | Number of critical-risk files                         |
| stale_count      | int    | Number of potentially unused files                    |
| churn_matrix_url | string | URL to the full churn-matrix endpoint                 |

## Relationships

```
ChurnMatrixResult
├── Repository          (1:1, shared model from internal/models)
├── []FileChurn         (1:N, all files modified in commit window)
├── []RiskEntry         (1:N, files with both churn + complexity data)
│   └── risk_zone       (subset of RiskEntry where risk_level == "critical")
├── Thresholds          (1:1, computed percentiles)
└── []StaleFile         (1:N, files matching all 3 stale conditions)
```

## State Transitions

### Churn Job Lifecycle

```
queued → running → completed
                 → failed

Special: running → requeued (if complexity data not yet available, 30s delay)
```

## Cache Keys

| Key Pattern                              | TTL  | Description                    |
|------------------------------------------|------|--------------------------------|
| `{owner}/{repo}:{sha}:churn`            | 24h  | Cached ChurnMatrixResult       |
| `active:{owner}/{repo}:{sha}:churn`     | 1h   | Active churn job ID            |

## Validation Rules

- `churn` MUST be >= 0 for all files
- `risk_level` MUST be one of: "critical", "high", "medium", "low"
- `months_inactive` MUST be >= 6 for stale files (by definition)
- `total_commits` MUST be <= 5,000 (commit window cap)
- `commit_window_start` MUST be no earlier than 2 years before analysis time
