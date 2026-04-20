import { apiFetch } from './client';
import type { AnalysisResult } from '../types/analysis';
import type { ComplexityResult } from '../types/complexity';
import type { ChurnMatrixResult } from '../types/churn';
import type { ContributorResult } from '../types/contributors';
import type { TemporalResult } from '../types/temporal';
import type { AISummary, HealthScore } from '../types/ai';
import type { AnalysisJob } from '../types/status';

const base = (owner: string, repo: string) => `/api/${owner}/${repo}`;

export function fetchAnalysis(owner: string, repo: string) {
  return apiFetch<AnalysisResult>(base(owner, repo));
}

export function fetchStatus(owner: string, repo: string) {
  return apiFetch<AnalysisJob>(`${base(owner, repo)}/status`);
}

export function fetchComplexity(owner: string, repo: string) {
  return apiFetch<ComplexityResult>(`${base(owner, repo)}/complexity`);
}

export function fetchChurnMatrix(owner: string, repo: string) {
  return apiFetch<ChurnMatrixResult>(`${base(owner, repo)}/churn-matrix`);
}

export function fetchContributors(owner: string, repo: string) {
  return apiFetch<ContributorResult>(`${base(owner, repo)}/contributors`);
}

export function fetchTemporal(owner: string, repo: string, period = '3y') {
  return apiFetch<TemporalResult>(`${base(owner, repo)}/temporal?period=${period}`);
}

export function fetchAISummary(owner: string, repo: string) {
  return apiFetch<AISummary>(`${base(owner, repo)}/ai/summary`, { method: 'POST' });
}

export function fetchAIHealth(owner: string, repo: string) {
  return apiFetch<HealthScore>(`${base(owner, repo)}/ai/health`);
}

export function deleteCache(owner: string, repo: string) {
  return apiFetch<void>(`${base(owner, repo)}/cache`, { method: 'DELETE' });
}
