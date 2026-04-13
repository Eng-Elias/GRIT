# Data Model: AST-Based Code Complexity Analysis

**Feature**: 002-complexity-analysis  
**Date**: 2026-03-30  
**Source**: [spec.md](spec.md) Key Entities + [research.md](research.md)

## Entities

### FunctionComplexity

Per-function complexity detail extracted from AST parsing.

| Field                | Type     | Description                                      | Constraints            |
|----------------------|----------|--------------------------------------------------|------------------------|
| `name`               | string   | Function or method name                          | Non-empty              |
| `start_line`         | int      | 1-indexed start line in source file              | ≥ 1                    |
| `end_line`           | int      | 1-indexed end line in source file                | ≥ start_line           |
| `cyclomatic`         | int      | Cyclomatic complexity (1 + decision points)      | ≥ 1                    |
| `cognitive`          | int      | Cognitive complexity (nesting-weighted score)     | ≥ 0                    |

**Notes**:
- Cyclomatic minimum is 1 (a function with no decision points has CC=1)
- Cognitive minimum is 0 (a linear function has zero cognitive load)
- Anonymous functions / lambdas use `<anonymous>` as name with file-unique suffix

---

### FileComplexity

Per-file complexity breakdown aggregated from function-level metrics.

| Field                    | Type                 | Description                                          | Constraints            |
|--------------------------|----------------------|------------------------------------------------------|------------------------|
| `path`                   | string               | File path relative to repo root                      | Non-empty              |
| `language`               | string               | Detected programming language                        | From supported set     |
| `cyclomatic`             | int                  | Sum of cyclomatic complexity across all functions     | ≥ 0                    |
| `cognitive`              | int                  | Sum of cognitive complexity across all functions      | ≥ 0                    |
| `function_count`         | int                  | Number of top-level functions and methods             | ≥ 0                    |
| `avg_function_complexity`| float64              | cyclomatic / function_count (0 if no functions)      | ≥ 0                    |
| `max_function_complexity`| int                  | Highest cyclomatic complexity of any single function  | ≥ 0                    |
| `loc`                    | int                  | Lines of code (from core analysis FileStats)         | ≥ 0                    |
| `complexity_density`     | float64              | cyclomatic / loc (0 if loc is 0)                     | ≥ 0                    |
| `functions`              | []FunctionComplexity | Per-function breakdown                               | May be empty           |

**Derived fields**:
- `avg_function_complexity` = `cyclomatic / function_count` (0 when `function_count == 0`)
- `complexity_density` = `cyclomatic / loc` (0 when `loc == 0`)
- `max_function_complexity` = `max(f.cyclomatic for f in functions)` (0 when empty)

---

### ComplexityDistribution

Histogram of files bucketed by average function complexity.

| Field      | Type | Description                                            | Constraints |
|------------|------|--------------------------------------------------------|-------------|
| `low`      | int  | Files with avg function complexity 1–5                 | ≥ 0         |
| `medium`   | int  | Files with avg function complexity 6–10                | ≥ 0         |
| `high`     | int  | Files with avg function complexity 11–20               | ≥ 0         |
| `critical` | int  | Files with avg function complexity 21+                 | ≥ 0         |

**Bucket rules**:
- Files with 0 functions are excluded from the distribution
- Bucket thresholds are inclusive: Low = [1, 5], Medium = [6, 10], High = [11, 20], Critical = [21, ∞)

---

### ComplexityResult

Complete complexity analysis output for a repository at a specific commit.

| Field                | Type                     | Description                                      | Constraints               |
|----------------------|--------------------------|--------------------------------------------------|---------------------------|
| `repository`         | Repository               | Repository identifier (shared with core)         | Required                  |
| `files`              | []FileComplexity         | Per-file complexity breakdown                    | May be empty              |
| `hot_files`          | []FileComplexity         | Top 20 files by complexity_density (descending)  | len ≤ 20                  |
| `total_files_analyzed`| int                     | Number of supported-language files analyzed       | ≥ 0                       |
| `total_function_count`| int                     | Sum of function_count across all files            | ≥ 0                       |
| `mean_complexity`    | float64                  | Mean cyclomatic complexity across all files       | ≥ 0                       |
| `median_complexity`  | float64                  | Median cyclomatic complexity across all files     | ≥ 0                       |
| `p90_complexity`     | float64                  | 90th percentile cyclomatic complexity             | ≥ 0                       |
| `distribution`       | ComplexityDistribution   | Histogram by avg function complexity              | Required                  |
| `analyzed_at`        | time.Time (ISO 8601)     | Timestamp of analysis completion                 | UTC                       |
| `cached`             | bool                     | Whether this result was served from cache        | Required                  |
| `cached_at`          | *time.Time (ISO 8601)    | When the result was cached (null if not cached)  | Nullable                  |

**Aggregate computation**:
- `mean_complexity`: average of `file.cyclomatic` across all files with ≥1 function
- `median_complexity`: median of `file.cyclomatic` across all files with ≥1 function
- `p90_complexity`: 90th percentile of `file.cyclomatic` across all files with ≥1 function
- Files with 0 functions are excluded from aggregate statistics

---

### ComplexitySummary

Abbreviated complexity data embedded in the main `GET /api/:owner/:repo` response.

| Field                | Type    | Description                                       | Constraints            |
|----------------------|---------|---------------------------------------------------|------------------------|
| `status`             | string  | "complete" or "pending"                           | Required               |
| `mean_complexity`    | float64 | Mean cyclomatic complexity (0 if pending)         | ≥ 0                    |
| `total_function_count`| int    | Total functions analyzed (0 if pending)           | ≥ 0                    |
| `hot_file_count`     | int     | Number of hot files identified (0 if pending)     | ≥ 0                    |
| `complexity_url`     | string  | URL to full complexity data                       | Non-empty              |

---

## Relationships

```
ComplexityResult
├── repository: Repository (shared entity from core pillar)
├── files: []FileComplexity
│   └── functions: []FunctionComplexity
├── hot_files: []FileComplexity (subset of files, sorted by density)
└── distribution: ComplexityDistribution

ComplexitySummary (embedded in AnalysisResult from core pillar)
└── references ComplexityResult via complexity_url
```

## Cache Keys

| Key Pattern                              | Value              | TTL  |
|------------------------------------------|--------------------|------|
| `{owner}/{repo}:{sha}:complexity`        | ComplexityResult   | 24h  |
| `job:{job_id}`                           | AnalysisJob        | 1h   |
| `active:{owner}/{repo}:{sha}:complexity` | job_id string      | 10m  |

## State Transitions

Complexity analysis job lifecycle (reuses existing `AnalysisJob` model):

```
queued → running → completed
                 → failed
```

Sub-job progress for complexity (extends `JobProgress` or uses separate tracking):
- `ast_parsing`: pending → running → completed/failed
- `aggregation`: pending → running → completed/failed
