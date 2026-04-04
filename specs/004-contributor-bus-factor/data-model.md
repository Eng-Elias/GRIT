# Data Model: Contributor Attribution & Bus Factor Analysis

**Date**: 2026-04-04  
**Feature**: [spec.md](spec.md) | [plan.md](plan.md)

## Entities

### Author

Represents a deduplicated contributor identified by normalized email.

| Field | Type | Description |
|-------|------|-------------|
| `email` | `string` | Canonical email (lowercased). Primary key for dedup. |
| `name` | `string` | Display name from most recent commit for this email. |
| `total_lines_owned` | `int` | Total lines attributed to this author across all files. |
| `ownership_percent` | `float64` | `total_lines_owned / total_repo_lines * 100`. |
| `files_touched` | `int` | Count of distinct files where this author owns ≥1 line. |
| `first_commit_date` | `time.Time` | Earliest blame date across all attributed lines. |
| `last_commit_date` | `time.Time` | Latest blame date across all attributed lines. |
| `primary_languages` | `[]string` | Top 3 languages by LOC owned (e.g., `["Go", "Python", "JavaScript"]`). |
| `is_active` | `bool` | `true` if `last_commit_date` is within the past 6 months. |

**Validation rules**:
- `email` must be non-empty after normalization.
- `ownership_percent` values across all authors must sum to 100% (±0.1% rounding).
- `primary_languages` contains at most 3 entries.

### FileContributors

Per-file ownership breakdown.

| Field | Type | Description |
|-------|------|-------------|
| `path` | `string` | Relative file path from repository root. |
| `total_lines` | `int` | Total blameworthy lines in this file. |
| `top_authors` | `[]FileAuthor` | Top 3 authors by line ownership for this file. |

### FileAuthor

A single author's contribution to a specific file.

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Author display name. |
| `email` | `string` | Canonical email (lowercased). |
| `lines_owned` | `int` | Lines owned in this specific file. |
| `ownership_percent` | `float64` | `lines_owned / file_total_lines * 100`. |

### ContributorResult

Complete blame analysis output for a repository. Stored in Redis.

| Field | Type | Description |
|-------|------|-------------|
| `repository` | `Repository` | Existing repo metadata (owner, name, SHA, etc.). |
| `authors` | `[]Author` | All deduplicated authors, sorted by `total_lines_owned` desc. |
| `bus_factor` | `int` | Minimum authors for 80% ownership. |
| `key_people` | `[]Author` | Authors comprising the 80% threshold. |
| `file_contributors` | `[]FileContributors` | Per-file top-3 breakdown. |
| `total_lines_analyzed` | `int` | Sum of all blameworthy lines across all files. |
| `total_files_analyzed` | `int` | Count of source files that were blamed. |
| `partial` | `bool` | `true` if timeout occurred before all files were processed. |
| `analyzed_at` | `time.Time` | Timestamp when analysis completed. |
| `cached` | `bool` | Set to `true` when served from cache. |
| `cached_at` | `*time.Time` | Timestamp when cache hit occurred. |

### ContributorSummary

Abbreviated contributor data embedded in main `GET /api/:owner/:repo` response.

| Field | Type | Description |
|-------|------|-------------|
| `status` | `string` | `"pending"` or `"complete"`. |
| `bus_factor` | `int` | Repository bus factor. |
| `top_contributors` | `[]AuthorBrief` | Top 3 authors (name, email, lines_owned). |
| `total_authors` | `int` | Total distinct authors. |
| `active_authors` | `int` | Authors with `is_active == true`. |
| `contributors_url` | `string` | Link to full `/api/:owner/:repo/contributors`. |

### AuthorBrief

Minimal author info for summary embedding.

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Author display name. |
| `email` | `string` | Canonical email. |
| `lines_owned` | `int` | Total lines owned. |

## Relationships

```
ContributorResult
├── repository: Repository (existing, reused)
├── authors: []Author
├── key_people: []Author (subset of authors)
└── file_contributors: []FileContributors
    └── top_authors: []FileAuthor

ContributorSummary (embedded in AnalysisResult response)
└── top_contributors: []AuthorBrief
```

## State Transitions

### Blame Job Lifecycle

```
[core completes] → PublishBlame() → queued → running → completed / failed
                                                ↓ (timeout)
                                         completed (partial: true)
```

### Cache States

| State | Endpoint Response | HTTP Status |
|-------|-------------------|-------------|
| No analysis exists | 404 Not Found | 404 |
| Blame job queued/running | Job status JSON | 202 |
| Blame complete (full) | ContributorResult (`partial: false`) | 200, `X-Cache: HIT` |
| Blame complete (partial) | ContributorResult (`partial: true`) | 200, `X-Cache: HIT` |

## Redis Keys

| Key Pattern | TTL | Content |
|-------------|-----|---------|
| `{owner}/{repo}:{sha}:contributors` | 48h | JSON-encoded `ContributorResult` |
| `active:{owner}/{repo}:{sha}:blame` | 10min | Job ID string |
| `job:{job_id}` | 1h | JSON-encoded `AnalysisJob` |
