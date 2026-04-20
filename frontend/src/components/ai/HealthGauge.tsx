import { useQuery } from '@tanstack/react-query';
import { fetchAIHealth } from '../../api/endpoints';
import { ApiError } from '../../api/client';
import Skeleton from '../shared/Skeleton';
import ErrorBanner from '../shared/ErrorBanner';

interface HealthGaugeProps {
  owner: string;
  repo: string;
}

function scoreColor(score: number) {
  if (score >= 80) return 'text-green-400';
  if (score >= 60) return 'text-yellow-400';
  if (score >= 40) return 'text-orange-400';
  return 'text-red-400';
}

function barColor(score: number) {
  if (score >= 80) return 'bg-green-500';
  if (score >= 60) return 'bg-yellow-500';
  if (score >= 40) return 'bg-orange-500';
  return 'bg-red-500';
}

export default function HealthGauge({ owner, repo }: HealthGaugeProps) {
  const { data, error, isLoading } = useQuery({
    queryKey: ['ai-health', owner, repo],
    queryFn: () => fetchAIHealth(owner, repo),
    enabled: !!owner && !!repo,
    retry: (failureCount, err) => {
      if (err instanceof ApiError && [404, 409, 503].includes(err.status)) return false;
      return failureCount < 1;
    },
  });

  if (isLoading) return <Skeleton rows={4} height="h-6" />;
  if (error) return <ErrorBanner error={error} />;
  if (!data) return null;

  const categories = Object.entries(data.categories) as [string, { score: number; notes: string }][];

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4 mb-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-zinc-400">AI Health Score</h3>
        <span className={`text-3xl font-bold ${scoreColor(data.overall_score)}`}>
          {data.overall_score}
        </span>
      </div>

      <div className="space-y-3">
        {categories.map(([key, cat]) => {
          const label = key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
          return (
            <div key={key}>
              <div className="flex items-center justify-between text-sm mb-1">
                <span className="text-zinc-400">{label}</span>
                <span className={`font-medium ${scoreColor(cat.score)}`}>{cat.score}/100</span>
              </div>
              <div className="h-1.5 rounded-full bg-zinc-800 overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all duration-500 ${barColor(cat.score)}`}
                  style={{ width: `${cat.score}%` }}
                />
              </div>
              {cat.notes && <p className="text-xs text-zinc-600 mt-0.5">{cat.notes}</p>}
            </div>
          );
        })}
      </div>

      {data.top_improvements.length > 0 && (
        <div className="mt-4 pt-3 border-t border-zinc-800">
          <p className="text-xs text-zinc-500 uppercase tracking-wider mb-2">Top Improvements</p>
          <ul className="space-y-1">
            {data.top_improvements.map((item, i) => (
              <li key={i} className="text-sm text-zinc-400 flex items-start gap-2">
                <span className="text-violet-400 shrink-0">•</span>
                {item}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
