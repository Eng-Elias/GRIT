import { useState, useMemo } from 'react';
import { ArrowUpDown } from 'lucide-react';
import clsx from 'clsx';
import RiskBadge from '../shared/RiskBadge';
import type { FileComplexity } from '../../types/complexity';

interface HotFilesTableProps {
  files: FileComplexity[];
}

type SortKey = 'path' | 'cyclomatic' | 'cognitive' | 'function_count' | 'loc' | 'complexity_density';

function riskLevel(density: number): string {
  if (density >= 0.5) return 'critical';
  if (density >= 0.3) return 'high';
  if (density >= 0.15) return 'medium';
  return 'low';
}

export default function HotFilesTable({ files }: HotFilesTableProps) {
  const [sortKey, setSortKey] = useState<SortKey>('complexity_density');
  const [sortAsc, setSortAsc] = useState(false);
  const [langFilter, setLangFilter] = useState('');

  const languages = useMemo(() => [...new Set(files.map(f => f.language))].sort(), [files]);

  const sorted = useMemo(() => {
    let filtered = langFilter ? files.filter(f => f.language === langFilter) : files;
    return [...filtered].sort((a, b) => {
      const av = a[sortKey], bv = b[sortKey];
      if (typeof av === 'string' && typeof bv === 'string') return sortAsc ? av.localeCompare(bv) : bv.localeCompare(av);
      return sortAsc ? (av as number) - (bv as number) : (bv as number) - (av as number);
    });
  }, [files, sortKey, sortAsc, langFilter]);

  function toggleSort(key: SortKey) {
    if (sortKey === key) setSortAsc(!sortAsc);
    else { setSortKey(key); setSortAsc(false); }
  }

  const cols: { key: SortKey; label: string; right?: boolean }[] = [
    { key: 'path', label: 'File' },
    { key: 'cyclomatic', label: 'Cyclomatic', right: true },
    { key: 'cognitive', label: 'Cognitive', right: true },
    { key: 'function_count', label: 'Functions', right: true },
    { key: 'loc', label: 'LOC', right: true },
    { key: 'complexity_density', label: 'Density', right: true },
  ];

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4">
      <div className="flex items-center justify-between mb-3 flex-wrap gap-2">
        <h3 className="text-sm font-medium text-zinc-400">Hot Files</h3>
        <select
          value={langFilter}
          onChange={e => setLangFilter(e.target.value)}
          className="rounded border border-zinc-700 bg-zinc-800 px-2 py-1 text-xs text-zinc-300 outline-none"
        >
          <option value="">All Languages</option>
          {languages.map(l => <option key={l} value={l}>{l}</option>)}
        </select>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800">
              {cols.map(({ key, label, right }) => (
                <th
                  key={key}
                  onClick={() => toggleSort(key)}
                  className={clsx(
                    'cursor-pointer px-3 py-2 font-medium text-zinc-500 hover:text-zinc-300 transition-colors whitespace-nowrap',
                    right ? 'text-right' : 'text-left',
                  )}
                >
                  <span className="inline-flex items-center gap-1">
                    {label}
                    {sortKey === key && <ArrowUpDown className="h-3 w-3" />}
                  </span>
                </th>
              ))}
              <th className="px-3 py-2 text-right font-medium text-zinc-500">Risk</th>
            </tr>
          </thead>
          <tbody>
            {sorted.slice(0, 20).map(f => (
              <tr key={f.path} className="border-b border-zinc-800/50 hover:bg-zinc-800/30 transition-colors">
                <td className="px-3 py-2 text-zinc-300 max-w-xs truncate" title={f.path}>{f.path}</td>
                <td className="px-3 py-2 text-right text-zinc-400">{f.cyclomatic}</td>
                <td className="px-3 py-2 text-right text-zinc-400">{f.cognitive}</td>
                <td className="px-3 py-2 text-right text-zinc-400">{f.function_count}</td>
                <td className="px-3 py-2 text-right text-zinc-400">{f.loc}</td>
                <td className="px-3 py-2 text-right text-zinc-400">{f.complexity_density.toFixed(3)}</td>
                <td className="px-3 py-2 text-right"><RiskBadge level={riskLevel(f.complexity_density)} /></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {sorted.length > 20 && (
        <p className="mt-2 text-xs text-zinc-600 text-center">Showing top 20 of {sorted.length} files</p>
      )}
    </div>
  );
}
