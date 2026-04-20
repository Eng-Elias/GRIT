import { useState, useMemo } from 'react';
import { ArrowUpDown, Circle } from 'lucide-react';
import clsx from 'clsx';
import type { Author } from '../../types/contributors';

interface AuthorTableProps {
  authors: Author[];
}

type SortKey = 'name' | 'total_lines_owned' | 'ownership_percent' | 'files_touched';

export default function AuthorTable({ authors }: AuthorTableProps) {
  const [sortKey, setSortKey] = useState<SortKey>('ownership_percent');
  const [sortAsc, setSortAsc] = useState(false);

  const sorted = useMemo(() => {
    return [...authors].sort((a, b) => {
      const av = a[sortKey], bv = b[sortKey];
      if (typeof av === 'string' && typeof bv === 'string') return sortAsc ? av.localeCompare(bv) : bv.localeCompare(av);
      return sortAsc ? (av as number) - (bv as number) : (bv as number) - (av as number);
    });
  }, [authors, sortKey, sortAsc]);

  function toggleSort(key: SortKey) {
    if (sortKey === key) setSortAsc(!sortAsc);
    else { setSortKey(key); setSortAsc(false); }
  }

  const cols: { key: SortKey; label: string; right?: boolean }[] = [
    { key: 'name', label: 'Contributor' },
    { key: 'total_lines_owned', label: 'Lines Owned', right: true },
    { key: 'ownership_percent', label: 'Ownership %', right: true },
    { key: 'files_touched', label: 'Files', right: true },
  ];

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4">
      <h3 className="text-sm font-medium text-zinc-400 mb-3">Contributors ({authors.length})</h3>
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
              <th className="px-3 py-2 text-right font-medium text-zinc-500">Status</th>
            </tr>
          </thead>
          <tbody>
            {sorted.map(a => (
              <tr key={a.email} className="border-b border-zinc-800/50 hover:bg-zinc-800/30 transition-colors">
                <td className="px-3 py-2">
                  <div>
                    <p className="text-zinc-300 font-medium">{a.name}</p>
                    <p className="text-xs text-zinc-600">{a.email}</p>
                  </div>
                </td>
                <td className="px-3 py-2 text-right text-zinc-400">{a.total_lines_owned.toLocaleString()}</td>
                <td className="px-3 py-2 text-right text-zinc-400">{a.ownership_percent.toFixed(1)}%</td>
                <td className="px-3 py-2 text-right text-zinc-400">{a.files_touched}</td>
                <td className="px-3 py-2 text-right">
                  <span className="inline-flex items-center gap-1 text-xs">
                    <Circle className={clsx('h-2 w-2 fill-current', a.is_active ? 'text-green-400' : 'text-zinc-600')} />
                    {a.is_active ? 'Active' : 'Inactive'}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
