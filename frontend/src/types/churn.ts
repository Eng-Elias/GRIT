import type { Repository } from './analysis';

export interface FileChurn {
  path: string;
  churn: number;
  last_modified: string;
}

export interface RiskEntry {
  path: string;
  churn: number;
  complexity_cyclomatic: number;
  language: string;
  loc: number;
  risk_level: string;
}

export interface StaleFile {
  path: string;
  last_modified: string;
  months_inactive: number;
}

export interface Thresholds {
  churn_p50: number;
  churn_p75: number;
  churn_p90: number;
  complexity_p50: number;
  complexity_p75: number;
  complexity_p90: number;
}

export interface ChurnMatrixResult {
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
