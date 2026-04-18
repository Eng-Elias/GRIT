export type SubJobStatus = 'pending' | 'running' | 'completed' | 'failed';

export interface JobProgress {
  clone: SubJobStatus;
  file_walk: SubJobStatus;
  metadata_fetch: SubJobStatus;
  commit_activity_fetch: SubJobStatus;
}

export interface AnalysisJob {
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
