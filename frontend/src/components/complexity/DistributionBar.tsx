import type { ComplexityDistribution } from '../../types/complexity';

interface DistributionBarProps {
  distribution: ComplexityDistribution;
}

const SEGMENTS = [
  { key: 'low' as const, label: 'Low (1-5)', color: 'bg-green-500' },
  { key: 'medium' as const, label: 'Medium (6-10)', color: 'bg-yellow-500' },
  { key: 'high' as const, label: 'High (11-20)', color: 'bg-orange-500' },
  { key: 'critical' as const, label: 'Critical (21+)', color: 'bg-red-500' },
];

export default function DistributionBar({ distribution }: DistributionBarProps) {
  const total = distribution.low + distribution.medium + distribution.high + distribution.critical;
  if (total === 0) return null;

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4 mb-6">
      <h3 className="text-sm font-medium text-zinc-400 mb-3">Complexity Distribution</h3>
      <div className="flex h-4 overflow-hidden rounded-full bg-zinc-800">
        {SEGMENTS.map(({ key, color }) => {
          const pct = (distribution[key] / total) * 100;
          if (pct === 0) return null;
          return (
            <div
              key={key}
              className={`${color} transition-all duration-300`}
              style={{ width: `${pct}%` }}
              title={`${key}: ${distribution[key]} files (${pct.toFixed(1)}%)`}
            />
          );
        })}
      </div>
      <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1">
        {SEGMENTS.map(({ key, label, color }) => (
          <span key={key} className="flex items-center gap-1.5 text-xs text-zinc-400">
            <span className={`h-2.5 w-2.5 rounded-full ${color}`} />
            {label}: {distribution[key]}
          </span>
        ))}
      </div>
    </div>
  );
}
