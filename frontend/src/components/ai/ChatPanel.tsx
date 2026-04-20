import { useState, useRef, useEffect, type FormEvent } from 'react';
import { Send, Loader2, Trash2 } from 'lucide-react';
import { useSSE } from '../../hooks/useSSE';
import type { ChatMessage } from '../../types/ai';

interface ChatPanelProps {
  owner: string;
  repo: string;
}

export default function ChatPanel({ owner, repo }: ChatPanelProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const { fullText, isStreaming, error, start, reset } = useSSE();
  const bottomRef = useRef<HTMLDivElement>(null);
  const streamingAssistantRef = useRef('');

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, fullText]);

  useEffect(() => {
    streamingAssistantRef.current = fullText;
  }, [fullText]);

  useEffect(() => {
    if (!isStreaming && streamingAssistantRef.current) {
      setMessages(prev => [...prev, { role: 'assistant', content: streamingAssistantRef.current }]);
      streamingAssistantRef.current = '';
      reset();
    }
  }, [isStreaming, reset]);

  function handleSend(e: FormEvent) {
    e.preventDefault();
    const trimmed = input.trim();
    if (!trimmed || isStreaming) return;

    const userMsg: ChatMessage = { role: 'user', content: trimmed };
    const nextMessages = [...messages, userMsg];
    setMessages(nextMessages);
    setInput('');

    start(`/api/${owner}/${repo}/ai/chat`, { messages: nextMessages });
  }

  function handleClear() {
    setMessages([]);
    reset();
  }

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 flex flex-col" style={{ height: '480px' }}>
      <div className="flex items-center justify-between border-b border-zinc-800 px-4 py-2">
        <h3 className="text-sm font-medium text-zinc-400">AI Chat</h3>
        <button
          onClick={handleClear}
          className="text-xs text-zinc-600 hover:text-zinc-400 transition-colors flex items-center gap-1"
        >
          <Trash2 className="h-3 w-3" /> Clear
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 && !isStreaming && (
          <p className="text-sm text-zinc-600 italic text-center pt-8">
            Ask anything about this repository…
          </p>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            <div className={`max-w-[80%] rounded-lg px-3 py-2 text-sm ${
              msg.role === 'user'
                ? 'bg-violet-600/20 text-violet-200 border border-violet-700/40'
                : 'bg-zinc-800 text-zinc-300 border border-zinc-700/40'
            }`}>
              <p className="whitespace-pre-wrap">{msg.content}</p>
            </div>
          </div>
        ))}
        {isStreaming && fullText && (
          <div className="flex justify-start">
            <div className="max-w-[80%] rounded-lg px-3 py-2 text-sm bg-zinc-800 text-zinc-300 border border-zinc-700/40">
              <p className="whitespace-pre-wrap">{fullText}</p>
              <Loader2 className="h-3 w-3 text-violet-400 animate-spin mt-1" />
            </div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      {error && (
        <div className="px-4 py-2 text-xs text-red-400 border-t border-zinc-800">{error}</div>
      )}

      <form onSubmit={handleSend} className="flex items-center gap-2 border-t border-zinc-800 px-4 py-3">
        <input
          type="text"
          value={input}
          onChange={e => setInput(e.target.value)}
          placeholder="Ask about the codebase…"
          className="flex-1 bg-transparent text-sm text-white placeholder-zinc-600 outline-none"
          disabled={isStreaming}
        />
        <button
          type="submit"
          disabled={isStreaming || !input.trim()}
          className="rounded-md bg-violet-600 p-2 text-white hover:bg-violet-500 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        >
          {isStreaming ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
        </button>
      </form>
    </div>
  );
}
