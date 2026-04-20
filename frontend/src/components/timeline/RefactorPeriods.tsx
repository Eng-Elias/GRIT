import { Wrench } from 'lucide-react';
import type { RefactorPeriod } from '../../types/temporal';

interface RefactorPeriodsProps {
  periods: RefactorPeriod[];
}

export default function RefactorPeriods({ periods }: RefactorPeriodsProps) {
  if (periods.length === 0) return null;

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4">
      <h3 className="text-sm font-medium text-zinc-400 mb-3 flex items-center gap-1.5">
        <Wrench className="h-4 w-4 text-cyan-400" /> Refactor Periods ({periods.length})
      </h3>
      <div className="space-y-2">
        {periods.map((p, i) => {
          const start = new Date(p.start).toLocaleDateString('en-US', { year: 'numeric', month: 'short' });
          const end = new Date(p.end).toLocaleDateString('en-US', { year: 'numeric', month: 'short' });
          return (
            <div key={i} className="flex items-center justify-between text-sm py-2 border-b border-zinc-800/50 last:border-0">
              <div>
                <p className="text-zinc-300">{start} — {end}</p>
                <p className="text-xs text-zinc-600">{p.weeks} weeks</p>
              </div>
              <span className={`font-medium ${p.net_loc_change < 0 ? 'text-green-400' : 'text-orange-400'}`}>
                {p.net_loc_change > 0 ? '+' : ''}{p.net_loc_change.toLocaleString()} LOC
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
