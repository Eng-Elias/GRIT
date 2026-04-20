export interface Repository {
  owner: string;
  name: string;
  full_name: string;
  default_branch: string;
  latest_sha: string;
  size_kb?: number;
}

export interface GitHubMetadata {
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

export interface CommitActivity {
  weekly_counts: number[];
  total_commits: number;
}

export interface FileStats {
  path: string;
  language: string;
  total_lines: number;
  code_lines: number;
  comment_lines: number;
  blank_lines: number;
  byte_size: number;
}

export interface LanguageBreakdown {
  language: string;
  file_count: number;
  total_lines: number;
  code_lines: number;
  comment_lines: number;
  blank_lines: number;
  percentage: number;
}

export interface ComplexitySummaryEmbed {
  status: string;
  mean_complexity: number;
  total_function_count: number;
  hot_file_count: number;
  complexity_url: string;
}

export interface ChurnSummaryEmbed {
  status: string;
  total_files: number;
  critical_count: number;
  stale_count: number;
  churn_matrix_url: string;
}

export interface ContributorSummaryEmbed {
  status: string;
  bus_factor: number;
  top_contributors: { name: string; email: string; lines_owned: number }[];
  total_authors: number;
  active_authors: number;
  contributors_url: string;
}

export interface TemporalSummaryEmbed {
  status: string;
  current_loc: number;
  loc_trend_6m_percent: number;
  avg_weekly_commits: number;
  temporal_url: string;
}

export interface AISummaryEmbed {
  status: string;
  ai_summary_url: string;
}

export interface AnalysisResult {
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
  complexity_summary?: ComplexitySummaryEmbed;
  churn_summary?: ChurnSummaryEmbed;
  contributor_summary?: ContributorSummaryEmbed;
  temporal_summary?: TemporalSummaryEmbed;
  ai_summary?: AISummaryEmbed;
}
