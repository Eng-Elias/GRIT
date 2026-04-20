import { Link } from 'react-router-dom';
import { Clock, X } from 'lucide-react';
import type { RecentSearch } from '../../hooks/useRecentSearches';

interface RecentSearchesProps {
  searches: RecentSearch[];
  onClear: () => void;
}

export default function RecentSearches({ searches, onClear }: RecentSearchesProps) {
  if (searches.length === 0) return null;

  return (
    <div className="w-full max-w-xl">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs text-zinc-500 uppercase tracking-wider flex items-center gap-1">
          <Clock className="h-3 w-3" /> Recent
        </span>
        <button onClick={onClear} className="text-xs text-zinc-600 hover:text-zinc-400 transition-colors flex items-center gap-1">
          <X className="h-3 w-3" /> Clear
        </button>
      </div>
      <div className="flex flex-wrap gap-2">
        {searches.map(s => (
          <Link
            key={`${s.owner}/${s.repo}`}
            to={`/repo/${s.owner}/${s.repo}`}
            className="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-300 hover:border-violet-600 hover:text-white transition-colors"
          >
            {s.owner}/{s.repo}
          </Link>
        ))}
      </div>
    </div>
  );
}
