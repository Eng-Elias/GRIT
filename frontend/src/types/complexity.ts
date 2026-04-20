import type { Repository } from './analysis';

export interface FunctionComplexity {
  name: string;
  start_line: number;
  end_line: number;
  cyclomatic: number;
  cognitive: number;
}

export interface FileComplexity {
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

export interface ComplexityDistribution {
  low: number;
  medium: number;
  high: number;
  critical: number;
}

export interface ComplexityResult {
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
