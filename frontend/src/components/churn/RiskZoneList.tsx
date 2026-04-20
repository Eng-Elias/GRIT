import RiskBadge from '../shared/RiskBadge';
import type { RiskEntry } from '../../types/churn';

interface RiskZoneListProps {
  entries: RiskEntry[];
}

export default function RiskZoneList({ entries }: RiskZoneListProps) {
  if (entries.length === 0) return null;

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4 mb-6">
      <h3 className="text-sm font-medium text-zinc-400 mb-3">Risk Zone ({entries.length} files)</h3>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800">
              <th className="px-3 py-2 text-left font-medium text-zinc-500">File</th>
              <th className="px-3 py-2 text-right font-medium text-zinc-500">Churn</th>
              <th className="px-3 py-2 text-right font-medium text-zinc-500">Complexity</th>
              <th className="px-3 py-2 text-right font-medium text-zinc-500">LOC</th>
              <th className="px-3 py-2 text-right font-medium text-zinc-500">Language</th>
              <th className="px-3 py-2 text-right font-medium text-zinc-500">Risk</th>
            </tr>
          </thead>
          <tbody>
            {entries.map(e => (
              <tr key={e.path} className="border-b border-zinc-800/50 hover:bg-zinc-800/30 transition-colors">
                <td className="px-3 py-2 text-zinc-300 max-w-xs truncate" title={e.path}>{e.path}</td>
                <td className="px-3 py-2 text-right text-zinc-400">{e.churn}</td>
                <td className="px-3 py-2 text-right text-zinc-400">{e.complexity_cyclomatic}</td>
                <td className="px-3 py-2 text-right text-zinc-400">{e.loc}</td>
                <td className="px-3 py-2 text-right text-zinc-400">{e.language}</td>
                <td className="px-3 py-2 text-right"><RiskBadge level={e.risk_level} /></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
