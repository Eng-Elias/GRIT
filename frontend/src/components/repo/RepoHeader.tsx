import { Star, GitFork, ExternalLink, Scale, CalendarDays } from 'lucide-react';
import type { AnalysisResult } from '../../types/analysis';

interface RepoHeaderProps {
  data: AnalysisResult;
  children?: React.ReactNode;
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
}

function formatNumber(n: number) {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
  return String(n);
}

export default function RepoHeader({ data, children }: RepoHeaderProps) {
  const { repository, metadata } = data;

  return (
    <div className="mb-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold text-white truncate">
            {repository.owner}/{repository.name}
          </h1>
          {metadata.description && (
            <p className="mt-1 text-sm text-zinc-400 max-w-2xl">{metadata.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2 shrink-0">
          {children}
          <a
            href={`https://github.com/${repository.owner}/${repository.name}`}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-300 hover:border-zinc-600 hover:text-white transition-colors"
          >
            <ExternalLink className="h-3.5 w-3.5" /> GitHub
          </a>
        </div>
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-zinc-500">
        <span className="flex items-center gap-1"><Star className="h-3.5 w-3.5" /> {formatNumber(metadata.stars)}</span>
        <span className="flex items-center gap-1"><GitFork className="h-3.5 w-3.5" /> {formatNumber(metadata.forks)}</span>
        {metadata.primary_language && (
          <span className="rounded-full bg-zinc-800 px-2 py-0.5 text-xs text-zinc-300">
            {metadata.primary_language}
          </span>
        )}
        {metadata.license_spdx && (
          <span className="flex items-center gap-1"><Scale className="h-3.5 w-3.5" /> {metadata.license_spdx}</span>
        )}
        {metadata.pushed_at && (
          <span className="flex items-center gap-1"><CalendarDays className="h-3.5 w-3.5" /> Pushed {formatDate(metadata.pushed_at)}</span>
        )}
      </div>
    </div>
  );
}
