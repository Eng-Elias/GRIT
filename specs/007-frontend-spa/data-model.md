# Data Model: GRIT Frontend SPA

All TypeScript interfaces mirror the Go backend JSON responses exactly. Field names use `snake_case` matching the JSON tags. Types are defined in `frontend/src/types/`.

## Core Types (types/analysis.ts)

### Repository
```typescript
interface Repository {
  owner: string;
  name: string;
  full_name: string;
  default_branch: string;
  latest_sha: string;
  size_kb?: number;
}
```

### GitHubMetadata
```typescript
interface GitHubMetadata {
  name: string;
  description: string;
  homepage_url: string;
  stars: number;
  forks: number;
  watchers: number;
  open_issues: number;
  size_kb: number;
  default_branch: string;
  primary_language: string;
  license_name: string;
  license_spdx: string;
  topics: string[];
  created_at: string;
  updated_at: string;
  pushed_at: string;
  has_wiki: boolean;
  has_projects: boolean;
  has_discussions: boolean;
  community_health: number;
}
```

### AnalysisResult
```typescript
interface AnalysisResult {
  repository: Repository;
  metadata: GitHubMetadata;
  commit_activity: CommitActivity;
  files: FileStats[];
  languages: LanguageBreakdown[];
  total_files: number;
  total_lines: number;
  total_code_lines: number;
  total_comment_lines: number;
  total_blank_lines: number;
  analyzed_at: string;
  cached: boolean;
  cached_at: string | null;
  // Embedded summaries
  complexity_summary?: ComplexitySummaryEmbed;
  churn_summary?: ChurnSummaryEmbed;
  contributor_summary?: ContributorSummaryEmbed;
  temporal_summary?: TemporalSummaryEmbed;
  ai_summary?: AISummaryEmbed;
}
```

### FileStats
```typescript
interface FileStats {
  path: string;
  language: string;
  total_lines: number;
  code_lines: number;
  comment_lines: number;
  blank_lines: number;
  byte_size: number;
}
```

### LanguageBreakdown
```typescript
interface LanguageBreakdown {
  language: string;
  file_count: number;
  total_lines: number;
  code_lines: number;
  comment_lines: number;
  blank_lines: number;
  percentage: number;
}
```

### CommitActivity
```typescript
interface CommitActivity {
  weekly_counts: number[];
  total_commits: number;
}
```

## Complexity Types (types/complexity.ts)

### ComplexityResult
```typescript
interface ComplexityResult {
  repository: Repository;
  files: FileComplexity[];
  hot_files: FileComplexity[];
  total_files_analyzed: number;
  total_function_count: number;
  mean_complexity: number;
  median_complexity: number;
  p90_complexity: number;
  distribution: ComplexityDistribution;
  analyzed_at: string;
  cached: boolean;
  cached_at: string | null;
}
```

### FileComplexity
```typescript
interface FileComplexity {
  path: string;
  language: string;
  cyclomatic: number;
  cognitive: number;
  function_count: number;
  avg_function_complexity: number;
  max_function_complexity: number;
  loc: number;
  complexity_density: number;
  functions: FunctionComplexity[];
}
```

### FunctionComplexity
```typescript
interface FunctionComplexity {
  name: string;
  start_line: number;
  end_line: number;
  cyclomatic: number;
  cognitive: number;
}
```

### ComplexityDistribution
```typescript
interface ComplexityDistribution {
  low: number;
  medium: number;
  high: number;
  critical: number;
}
```

## Churn Types (types/churn.ts)

### ChurnMatrixResult
```typescript
interface ChurnMatrixResult {
  repository: Repository;
  churn: FileChurn[];
  risk_matrix: RiskEntry[];
  risk_zone: RiskEntry[];
  thresholds: Thresholds;
  stale_files: StaleFile[];
  total_commits: number;
  commit_window_start: string;
  commit_window_end: string;
  total_files_churned: number;
  critical_count: number;
  stale_count: number;
  analyzed_at: string;
  cached: boolean;
  cached_at: string | null;
}
```

### RiskEntry
```typescript
interface RiskEntry {
  path: string;
  churn: number;
  complexity_cyclomatic: number;
  language: string;
  loc: number;
  risk_level: string;
}
```

### StaleFile
```typescript
interface StaleFile {
  path: string;
  last_modified: string;
  months_inactive: number;
}
```

### Thresholds
```typescript
interface Thresholds {
  churn_p50: number;
  churn_p75: number;
  churn_p90: number;
  complexity_p50: number;
  complexity_p75: number;
  complexity_p90: number;
}
```

## Contributor Types (types/contributors.ts)

### ContributorResult
```typescript
interface ContributorResult {
  repository: Repository;
  authors: Author[];
  bus_factor: number;
  key_people: Author[];
  file_contributors: FileContributors[];
  total_lines_analyzed: number;
  total_files_analyzed: number;
  partial: boolean;
  analyzed_at: string;
  cached: boolean;
  cached_at: string | null;
}
```

### Author
```typescript
interface Author {
  email: string;
  name: string;
  total_lines_owned: number;
  ownership_percent: number;
  files_touched: number;
  first_commit_date: string;
  last_commit_date: string;
  primary_languages: string[];
  is_active: boolean;
}
```

## Temporal Types (types/temporal.ts)

### TemporalResult
```typescript
interface TemporalResult {
  repository: Repository;
  period: string;
  loc_over_time: MonthlySnapshot[];
  velocity: VelocityMetrics;
  refactor_periods: RefactorPeriod[];
  total_months: number;
  total_weeks: number;
  analyzed_at: string;
  cached: boolean;
  cached_at: string | null;
}
```

### MonthlySnapshot
```typescript
interface MonthlySnapshot {
  date: string;
  total_loc: number;
  by_language: Record<string, number>;
}
```

### WeeklyActivity
```typescript
interface WeeklyActivity {
  week: string;
  week_start: string;
  additions: number;
  deletions: number;
  commits: number;
}
```

### RefactorPeriod
```typescript
interface RefactorPeriod {
  start: string;
  end: string;
  net_loc_change: number;
  weeks: number;
}
```

## AI Types (types/ai.ts)

### AISummary
```typescript
interface AISummary {
  repository: Repository;
  description: string;
  architecture: string;
  tech_stack: string[];
  red_flags: string[];
  entry_points: string[];
  generated_at: string;
  model: string;
  cached: boolean;
  cached_at: string | null;
}
```

### HealthScore
```typescript
interface HealthScore {
  repository: Repository;
  overall_score: number;
  categories: HealthCategories;
  top_improvements: string[];
  generated_at: string;
  model: string;
  cached: boolean;
  cached_at: string | null;
}
```

### HealthCategories
```typescript
interface HealthCategories {
  readme_quality: CategoryScore;
  contributing_guide: CategoryScore;
  code_documentation: CategoryScore;
  test_coverage_signals: CategoryScore;
  project_hygiene: CategoryScore;
}
```

### CategoryScore
```typescript
interface CategoryScore {
  score: number;
  notes: string;
}
```

### ChatMessage
```typescript
interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}
```

### ChatRequest
```typescript
interface ChatRequest {
  messages: ChatMessage[];
}
```

## Job Status Types (types/status.ts)

### AnalysisJob
```typescript
interface AnalysisJob {
  job_id: string;
  owner: string;
  repo: string;
  sha: string;
  status: 'queued' | 'running' | 'completed' | 'failed';
  progress: JobProgress;
  created_at: string;
  completed_at?: string;
  error?: string;
  result_url?: string;
}
```

### JobProgress
```typescript
interface JobProgress {
  clone: SubJobStatus;
  file_walk: SubJobStatus;
  metadata_fetch: SubJobStatus;
  commit_activity_fetch: SubJobStatus;
}
```

```typescript
type SubJobStatus = 'pending' | 'running' | 'completed' | 'failed';
```

## Embedded Summary Types (in types/analysis.ts)

```typescript
interface ComplexitySummaryEmbed {
  status: string;
  mean_complexity: number;
  total_function_count: number;
  hot_file_count: number;
  complexity_url: string;
}

interface ChurnSummaryEmbed {
  status: string;
  total_files: number;
  critical_count: number;
  stale_count: number;
  churn_matrix_url: string;
}

interface ContributorSummaryEmbed {
  status: string;
  bus_factor: number;
  top_contributors: { name: string; email: string; lines_owned: number }[];
  total_authors: number;
  active_authors: number;
  contributors_url: string;
}

interface TemporalSummaryEmbed {
  status: string;
  current_loc: number;
  loc_trend_6m_percent: number;
  avg_weekly_commits: number;
  temporal_url: string;
}

interface AISummaryEmbed {
  status: string;
  ai_summary_url: string;
}
```

## Client-Only Types

### RecentSearch (localStorage)
```typescript
interface RecentSearch {
  owner: string;
  repo: string;
  timestamp: number;
}
```

### SSE Hook State
```typescript
interface SSEState<T> {
  data: T | null;
  chunks: string[];
  error: string | null;
  isStreaming: boolean;
}
```
