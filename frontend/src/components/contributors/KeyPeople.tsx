import { Crown } from 'lucide-react';
import type { Author } from '../../types/contributors';

interface KeyPeopleProps {
  keyPeople: Author[];
}

export default function KeyPeople({ keyPeople }: KeyPeopleProps) {
  if (keyPeople.length === 0) return null;

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4">
      <h3 className="text-sm font-medium text-zinc-400 mb-3 flex items-center gap-1.5">
        <Crown className="h-4 w-4 text-yellow-500" /> Key People
      </h3>
      <div className="space-y-3">
        {keyPeople.map(a => (
          <div key={a.email} className="flex items-center justify-between">
            <div className="min-w-0">
              <p className="text-sm font-medium text-zinc-300 truncate">{a.name}</p>
              <p className="text-xs text-zinc-600">{a.primary_languages?.join(', ') ?? '—'}</p>
            </div>
            <div className="text-right shrink-0 ml-3">
              <p className="text-sm font-bold text-white">{a.ownership_percent.toFixed(1)}%</p>
              <p className="text-xs text-zinc-600">{a.total_lines_owned.toLocaleString()} lines</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
