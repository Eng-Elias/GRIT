import type { Repository } from './analysis';

export interface AISummary {
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

export interface CategoryScore {
  score: number;
  notes: string;
}

export interface HealthCategories {
  readme_quality: CategoryScore;
  contributing_guide: CategoryScore;
  code_documentation: CategoryScore;
  test_coverage_signals: CategoryScore;
  project_hygiene: CategoryScore;
}

export interface HealthScore {
  repository: Repository;
  overall_score: number;
  categories: HealthCategories;
  top_improvements: string[];
  generated_at: string;
  model: string;
  cached: boolean;
  cached_at: string | null;
}

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

export interface ChatRequest {
  messages: ChatMessage[];
}
