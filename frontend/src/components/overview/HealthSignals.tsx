import { CheckCircle2, XCircle } from 'lucide-react';
import type { GitHubMetadata } from '../../types/analysis';

interface HealthSignalsProps {
  metadata: GitHubMetadata;
}

export default function HealthSignals({ metadata }: HealthSignalsProps) {
  const signals = [
    { label: 'License', present: !!metadata.license_spdx },
    { label: 'Wiki', present: metadata.has_wiki },
    { label: 'Projects', present: metadata.has_projects },
    { label: 'Discussions', present: metadata.has_discussions },
  ];

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4">
      <h3 className="text-sm font-medium text-zinc-400 mb-3">Health Signals</h3>
      <div className="grid grid-cols-2 gap-3">
        {signals.map(({ label, present }) => (
          <div key={label} className="flex items-center gap-2 text-sm">
            {present
              ? <CheckCircle2 className="h-4 w-4 text-green-400" />
              : <XCircle className="h-4 w-4 text-zinc-600" />}
            <span className={present ? 'text-zinc-300' : 'text-zinc-600'}>{label}</span>
          </div>
        ))}
      </div>
      {metadata.community_health > 0 && (
        <div className="mt-3 pt-3 border-t border-zinc-800">
          <div className="flex items-center justify-between text-sm">
            <span className="text-zinc-500">Community Health</span>
            <span className="text-white font-medium">{metadata.community_health}%</span>
          </div>
          <div className="mt-1 h-1.5 rounded-full bg-zinc-800 overflow-hidden">
            <div
              className="h-full rounded-full bg-green-500 transition-all duration-500"
              style={{ width: `${metadata.community_health}%` }}
            />
          </div>
        </div>
      )}
    </div>
  );
}
