import type { Repository } from './analysis';

export interface MonthlySnapshot {
  date: string;
  total_loc: number;
  by_language: Record<string, number>;
}

export interface WeeklyActivity {
  week: string;
  week_start: string;
  additions: number;
  deletions: number;
  commits: number;
}

export interface AuthorWeeklyActivity {
  email: string;
  name: string;
  total_additions: number;
  total_deletions: number;
  weeks: WeeklyActivity[];
}

export interface CommitCadence {
  window_start: string;
  window_end: string;
  commits_per_day: number;
}

export interface PRMergeTime {
  median_hours: number;
  p75_hours: number;
  p95_hours: number;
  sample_size: number;
}

export interface VelocityMetrics {
  weekly_activity: WeeklyActivity[];
  author_activity: AuthorWeeklyActivity[];
  commit_cadence: CommitCadence[];
  pr_merge_time: PRMergeTime | null;
}

export interface RefactorPeriod {
  start: string;
  end: string;
  net_loc_change: number;
  weeks: number;
}

export interface TemporalResult {
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
