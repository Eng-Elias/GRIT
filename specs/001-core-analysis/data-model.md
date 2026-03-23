# Data Model: Core Analysis Engine

**Branch**: `001-core-analysis`
**Date**: 2026-03-23

## Entities

### Repository

Represents a GitHub repository targeted for analysis.

| Field | Type | Description |
|-------|------|-------------|
| Owner | string | GitHub user or organization (e.g., `torvalds`) |
| Name | string | Repository name (e.g., `linux`) |
| FullName | string | `{Owner}/{Name}` |
| DefaultBranch | string | e.g., `main`, `master` |
| LatestSHA | string | HEAD commit SHA of default branch |
| SizeKB | int64 | Repository size in KB (from GitHub API) |

**Identity**: `{Owner}/{Name}` is globally unique.

---

### GitHubMetadata

Snapshot of GitHub repository metadata at analysis time.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Repository name |
| Description | string | Repository description |
| HomepageURL | string | Project homepage URL |
| Stars | int | Stargazer count |
| Forks | int | Fork count |
| Watchers | int | Watcher count |
| OpenIssues | int | Open issue count |
| SizeKB | int64 | Repo size in KB |
| DefaultBranch | string | Default branch name |
| PrimaryLanguage | string | GitHub-detected primary language |
| LicenseName | string | License display name (e.g., "MIT License") |
| LicenseSPDX | string | SPDX identifier (e.g., "MIT") |
| Topics | []string | Repository topics |
| CreatedAt | time.Time | Repository creation date |
| UpdatedAt | time.Time | Last update date |
| PushedAt | time.Time | Last push date |
| HasWiki | bool | Wiki enabled |
| HasProjects | bool | Projects enabled |
| HasDiscussions | bool | Discussions enabled |
| CommunityHealth | int | Community health percentage (0–100) |

---

### CommitActivity

Weekly commit counts for the past 52 weeks.

| Field | Type | Description |
|-------|------|-------------|
| WeeklyCounts | []int | 52 entries, oldest first. Each = commits in that week |
| TotalCommits | int | Total commit count on default branch |

---

### FileStats

Per-file line count breakdown.

| Field | Type | Description |
|-------|------|-------------|
| Path | string | File path relative to repo root |
| Language | string | Detected language name |
| TotalLines | int | Total line count |
| CodeLines | int | Non-comment, non-blank lines |
| CommentLines | int | Lines that are comments |
| BlankLines | int | Whitespace-only lines |
| ByteSize | int64 | File size in bytes |

**Identity**: `Path` is unique within an analysis result.

---

### LanguageBreakdown

Aggregated statistics per detected language.

| Field | Type | Description |
|-------|------|-------------|
| Language | string | Language name |
| FileCount | int | Number of files in this language |
| TotalLines | int | Sum of total lines across files |
| CodeLines | int | Sum of code lines |
| CommentLines | int | Sum of comment lines |
| BlankLines | int | Sum of blank lines |
| Percentage | float64 | Percentage of total lines (0.0–100.0) |

---

### AnalysisResult

The complete output of core analysis for a repository at a specific commit.

| Field | Type | Description |
|-------|------|-------------|
| Repository | Repository | Repository identity |
| Metadata | GitHubMetadata | GitHub metadata snapshot |
| CommitActivity | CommitActivity | 52-week commit data |
| Files | []FileStats | Per-file breakdown |
| Languages | []LanguageBreakdown | Per-language aggregation, sorted by TotalLines desc |
| TotalFiles | int | Total file count (excluding ignored) |
| TotalLines | int | Grand total lines |
| TotalCodeLines | int | Grand total code lines |
| TotalCommentLines | int | Grand total comment lines |
| TotalBlankLines | int | Grand total blank lines |
| AnalyzedAt | time.Time | When the analysis was performed |
| Cached | bool | Whether this result was served from cache |
| CachedAt | *time.Time | When the result was cached (null if not cached) |

**Identity**: `{Owner}/{Name}:{LatestSHA}:core` (also the Redis key).

---

### AnalysisJob

A background job representing an in-progress analysis.

| Field | Type | Description |
|-------|------|-------------|
| JobID | string | UUID v4 |
| Owner | string | Repository owner |
| Repo | string | Repository name |
| SHA | string | Commit SHA being analyzed |
| Status | JobStatus | Current status |
| Progress | JobProgress | Sub-job progress |
| CreatedAt | time.Time | Job creation time |
| CompletedAt | *time.Time | Job completion time (null if not done) |
| Error | string | Error message if failed |

---

### JobStatus (enum)

| Value | Description |
|-------|-------------|
| `queued` | Job published to NATS, not yet picked up |
| `running` | Worker is processing |
| `completed` | Analysis finished successfully |
| `failed` | Analysis failed with error |

---

### JobProgress

Tracks sub-job completion.

| Field | Type | Description |
|-------|------|-------------|
| Clone | SubJobStatus | Repository clone status |
| FileWalk | SubJobStatus | File tree walk + line counting status |
| MetadataFetch | SubJobStatus | GitHub metadata fetch status |
| CommitActivityFetch | SubJobStatus | Commit activity fetch status |

---

### SubJobStatus (enum)

| Value | Description |
|-------|-------------|
| `pending` | Not started |
| `running` | In progress |
| `completed` | Finished successfully |
| `failed` | Failed with error |

## Relationships

```text
AnalysisResult
├── 1:1 Repository
├── 1:1 GitHubMetadata
├── 1:1 CommitActivity
├── 1:N FileStats
└── 1:N LanguageBreakdown (derived from FileStats aggregation)

AnalysisJob
├── references Repository (by Owner/Repo/SHA)
├── 1:1 JobProgress
│   └── 4x SubJobStatus
└── on completion → produces AnalysisResult (cached in Redis)
```

## Redis Key Schema

| Key Pattern | Value | TTL |
|-------------|-------|-----|
| `{owner}/{repo}:{sha}:core` | JSON `AnalysisResult` | 24h |
| `job:{job_id}` | JSON `AnalysisJob` | 1h |
| `active:{owner}/{repo}:{sha}` | `{job_id}` (string) | 10m |

- **`active:*`** keys enable job deduplication: before publishing a new
  job, check if an active key exists for the same repo+SHA.
- **`job:*`** keys store job state for the polling endpoint.
- **`{owner}/{repo}:{sha}:core`** stores the final analysis result.
