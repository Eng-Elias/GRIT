import { useState, useCallback } from 'react';
import { Copy, Check, Shield } from 'lucide-react';
import type { AnalysisResult } from '../../types/analysis';

interface BadgeGeneratorProps {
  data: AnalysisResult;
}

type BadgeType = 'lines' | 'files' | 'languages' | 'complexity' | 'busFactor';

interface BadgeDef {
  id: BadgeType;
  label: string;
  getValue: (d: AnalysisResult) => string;
  color: string;
}

const BADGES: BadgeDef[] = [
  { id: 'lines', label: 'Lines of Code', getValue: d => d.total_lines.toLocaleString(), color: 'blue' },
  { id: 'files', label: 'Files', getValue: d => d.total_files.toLocaleString(), color: 'green' },
  { id: 'languages', label: 'Languages', getValue: d => String(d.languages.length), color: 'orange' },
  { id: 'complexity', label: 'Avg Complexity', getValue: d => d.complexity_summary?.mean_complexity?.toFixed(1) ?? 'N/A', color: 'yellow' },
  { id: 'busFactor', label: 'Bus Factor', getValue: d => String(d.contributor_summary?.bus_factor ?? 'N/A'), color: 'red' },
];

function shieldsUrl(label: string, value: string, color: string) {
  const l = encodeURIComponent(label);
  const v = encodeURIComponent(value);
  return `https://img.shields.io/badge/${l}-${v}-${color}?style=flat-square`;
}

function markdownBadge(label: string, url: string) {
  return `![${label}](${url})`;
}

export default function BadgeGenerator({ data }: BadgeGeneratorProps) {
  const [copied, setCopied] = useState<string | null>(null);

  const copyToClipboard = useCallback((id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(id);
    setTimeout(() => setCopied(null), 2000);
  }, []);

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4">
      <h3 className="text-sm font-medium text-zinc-400 mb-3 flex items-center gap-1.5">
        <Shield className="h-4 w-4 text-blue-400" /> Badges
      </h3>
      <div className="space-y-3">
        {BADGES.map(({ id, label, getValue, color }) => {
          const value = getValue(data);
          const url = shieldsUrl(label, value, color);
          const md = markdownBadge(label, url);
          return (
            <div key={id} className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-3 min-w-0">
                <img src={url} alt={label} className="h-5" />
                <code className="text-xs text-zinc-500 truncate">{md}</code>
              </div>
              <button
                onClick={() => copyToClipboard(id, md)}
                className="shrink-0 rounded border border-zinc-700 bg-zinc-800 p-1.5 text-zinc-400 hover:text-white hover:border-zinc-600 transition-colors"
                title="Copy markdown"
              >
                {copied === id ? <Check className="h-3.5 w-3.5 text-green-400" /> : <Copy className="h-3.5 w-3.5" />}
              </button>
            </div>
          );
        })}
      </div>
    </div>
  );
}
