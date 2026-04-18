import type { CommitActivity } from '../../types/analysis';

interface CommitHeatmapProps {
  activity: CommitActivity;
}

function intensity(count: number, max: number): string {
  if (count === 0) return 'bg-zinc-800';
  const ratio = count / max;
  if (ratio < 0.25) return 'bg-violet-900/60';
  if (ratio < 0.5) return 'bg-violet-800/70';
  if (ratio < 0.75) return 'bg-violet-700';
  return 'bg-violet-500';
}

export default function CommitHeatmap({ activity }: CommitHeatmapProps) {
  const counts = activity.weekly_counts ?? [];
  const weeks = counts.slice(-52);
  const max = Math.max(...weeks, 1);

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4 mb-6">
      <h3 className="text-sm font-medium text-zinc-400 mb-3">52-Week Commit Activity</h3>
      <div className="flex gap-1 overflow-x-auto pb-1">
        {weeks.map((count, i) => (
          <div
            key={i}
            className={`h-8 w-2.5 rounded-sm shrink-0 transition-colors ${intensity(count, max)}`}
            title={`Week ${i + 1}: ${count} commits`}
          />
        ))}
      </div>
      <div className="flex items-center gap-2 mt-2">
        <span className="text-xs text-zinc-600">Less</span>
        <div className="flex gap-0.5">
          {['bg-zinc-800', 'bg-violet-900/60', 'bg-violet-800/70', 'bg-violet-700', 'bg-violet-500'].map(c => (
            <div key={c} className={`h-3 w-3 rounded-sm ${c}`} />
          ))}
        </div>
        <span className="text-xs text-zinc-600">More</span>
      </div>
    </div>
  );
}
