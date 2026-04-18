import type { Repository } from './analysis';

export interface Author {
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

export interface FileAuthor {
  name: string;
  email: string;
  lines_owned: number;
  ownership_percent: number;
}

export interface FileContributors {
  path: string;
  total_lines: number;
  top_authors: FileAuthor[];
}

export interface ContributorResult {
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
