import { Loader2, CheckCircle2, XCircle, Clock } from 'lucide-react';
import type { AnalysisJob } from '../../types/status';
import type { SubJobStatus } from '../../types/status';

interface AnalysisStatusProps {
  job: AnalysisJob;
}

function statusIcon(s: SubJobStatus) {
  switch (s) {
    case 'completed': return <CheckCircle2 className="h-4 w-4 text-green-400" />;
    case 'running': return <Loader2 className="h-4 w-4 text-violet-400 animate-spin" />;
    case 'failed': return <XCircle className="h-4 w-4 text-red-400" />;
    default: return <Clock className="h-4 w-4 text-zinc-600" />;
  }
}

const LABELS: Record<string, string> = {
  clone: 'Clone',
  file_walk: 'File Walk',
  metadata_fetch: 'Metadata',
  commit_activity_fetch: 'Commits',
};

export default function AnalysisStatus({ job }: AnalysisStatusProps) {
  if (job.status === 'completed') return null;

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/80 px-4 py-3 mb-6">
      <div className="flex items-center gap-2 mb-3">
        {job.status === 'failed' ? (
          <XCircle className="h-5 w-5 text-red-400" />
        ) : (
          <Loader2 className="h-5 w-5 text-violet-400 animate-spin" />
        )}
        <span className="text-sm font-medium text-white">
          {job.status === 'failed' ? 'Analysis Failed' : 'Analysis In Progress'}
        </span>
      </div>
      {job.error && (
        <p className="text-sm text-red-400 mb-3">{job.error}</p>
      )}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {Object.entries(job.progress).map(([key, status]) => (
          <div key={key} className="flex items-center gap-2 text-sm text-zinc-400">
            {statusIcon(status)}
            <span>{LABELS[key] ?? key}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
