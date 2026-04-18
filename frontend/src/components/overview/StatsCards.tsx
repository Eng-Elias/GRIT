import { FileCode2, Files, GitCommitHorizontal, Users } from 'lucide-react';
import type { AnalysisResult } from '../../types/analysis';

interface StatsCardsProps {
  data: AnalysisResult;
}

function fmt(n: number) {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
  return String(n);
}

export default function StatsCards({ data }: StatsCardsProps) {
  const cards = [
    { label: 'Total LOC', value: fmt(data.total_lines), icon: FileCode2, color: 'text-blue-400' },
    { label: 'Files', value: fmt(data.total_files), icon: Files, color: 'text-green-400' },
    { label: 'Commits', value: fmt(data.commit_activity.total_commits), icon: GitCommitHorizontal, color: 'text-orange-400' },
    { label: 'Contributors', value: data.contributor_summary ? fmt(data.contributor_summary.total_authors) : '—', icon: Users, color: 'text-violet-400' },
  ];

  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      {cards.map(({ label, value, icon: Icon, color }) => (
        <div key={label} className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4">
          <div className="flex items-center gap-2 mb-1">
            <Icon className={`h-4 w-4 ${color}`} />
            <span className="text-xs text-zinc-500 uppercase tracking-wider">{label}</span>
          </div>
          <p className="text-2xl font-bold text-white">{value}</p>
        </div>
      ))}
    </div>
  );
}
