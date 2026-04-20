import { useCallback, useEffect, useState } from 'react';
import { Sparkles, Loader2 } from 'lucide-react';
import { useSSE } from '../../hooks/useSSE';
import type { AISummary } from '../../types/ai';
import ErrorBanner from '../shared/ErrorBanner';

interface AISummaryPanelProps {
  owner: string;
  repo: string;
}

export default function AISummaryPanel({ owner, repo }: AISummaryPanelProps) {
  const { fullText, error, isStreaming, start, reset } = useSSE();
  const [cached, setCached] = useState<AISummary | null>(null);

  const generate = useCallback(() => {
    setCached(null);
    start(`/api/${owner}/${repo}/ai/summary`);
  }, [owner, repo, start]);

  useEffect(() => {
    reset();
    setCached(null);
  }, [owner, repo, reset]);

  const displayText = cached?.description ?? fullText;

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4 mb-6">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-medium text-zinc-400 flex items-center gap-1.5">
          <Sparkles className="h-4 w-4 text-violet-400" /> AI Summary
        </h3>
        <button
          onClick={generate}
          disabled={isStreaming}
          className="inline-flex items-center gap-1.5 rounded-md border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-xs font-medium text-zinc-300 hover:border-violet-500 hover:text-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isStreaming ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Sparkles className="h-3.5 w-3.5" />}
          {isStreaming ? 'Generating…' : 'Generate'}
        </button>
      </div>

      {error && <ErrorBanner error={new Error(error)} className="mb-3" />}

      {displayText ? (
        <div className="prose prose-sm prose-invert max-w-none">
          <p className="text-sm text-zinc-300 whitespace-pre-wrap leading-relaxed">{displayText}</p>
        </div>
      ) : (
        <p className="text-sm text-zinc-600 italic">Click "Generate" to create an AI-powered summary of this repository.</p>
      )}
    </div>
  );
}
