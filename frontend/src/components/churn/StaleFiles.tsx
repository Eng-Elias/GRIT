import { useState } from 'react';
import { ChevronDown, ChevronRight, FileWarning } from 'lucide-react';
import type { StaleFile } from '../../types/churn';

interface StaleFilesProps {
  files: StaleFile[];
}

export default function StaleFiles({ files }: StaleFilesProps) {
  const [open, setOpen] = useState(false);

  if (files.length === 0) return null;

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-2 text-sm font-medium text-zinc-400 hover:text-zinc-300 transition-colors w-full"
      >
        {open ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
        <FileWarning className="h-4 w-4 text-yellow-500" />
        Stale Files ({files.length})
      </button>
      {open && (
        <div className="mt-3 space-y-1">
          {files.map(f => (
            <div key={f.path} className="flex items-center justify-between text-sm py-1 border-b border-zinc-800/50">
              <span className="text-zinc-400 truncate max-w-xs" title={f.path}>{f.path}</span>
              <span className="text-zinc-600 text-xs shrink-0 ml-2">{f.months_inactive} months inactive</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
