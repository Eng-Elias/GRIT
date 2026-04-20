import { useState, useCallback, useRef } from 'react';

export interface SSEState {
  chunks: string[];
  fullText: string;
  error: string | null;
  isStreaming: boolean;
}

const initialState: SSEState = {
  chunks: [],
  fullText: '',
  error: null,
  isStreaming: false,
};

export function useSSE() {
  const [state, setState] = useState<SSEState>(initialState);
  const abortRef = useRef<AbortController | null>(null);

  const start = useCallback(async (url: string, body?: object) => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    setState({ chunks: [], fullText: '', error: null, isStreaming: true });

    try {
      const res = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });

      if (!res.ok) {
        let msg = `Request failed (${res.status})`;
        try {
          const err = await res.json();
          msg = err.message ?? msg;
        } catch { /* not JSON */ }
        setState(s => ({ ...s, error: msg, isStreaming: false }));
        return;
      }

      const reader = res.body?.getReader();
      if (!reader) {
        setState(s => ({ ...s, error: 'No response body', isStreaming: false }));
        return;
      }

      const decoder = new TextDecoder();
      let buffer = '';
      let accumulated = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';

        for (const line of lines) {
          if (line.startsWith('event: done')) {
            setState(s => ({ ...s, isStreaming: false }));
            reader.cancel();
            return;
          }
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            if (data.startsWith('[ERROR]')) {
              setState(s => ({ ...s, error: data.slice(8), isStreaming: false }));
              reader.cancel();
              return;
            }
            accumulated += data;
            const current = accumulated;
            setState(s => ({
              ...s,
              chunks: [...s.chunks, data],
              fullText: current,
            }));
          }
        }
      }

      setState(s => ({ ...s, isStreaming: false }));
    } catch (err) {
      if ((err as Error).name === 'AbortError') return;
      setState(s => ({
        ...s,
        error: (err as Error).message ?? 'Stream failed',
        isStreaming: false,
      }));
    }
  }, []);

  const stop = useCallback(() => {
    abortRef.current?.abort();
    setState(s => ({ ...s, isStreaming: false }));
  }, []);

  const reset = useCallback(() => {
    abortRef.current?.abort();
    setState(initialState);
  }, []);

  return { ...state, start, stop, reset };
}
