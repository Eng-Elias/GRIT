import { AlertTriangle, ShieldOff, Clock, WifiOff, ServerOff, Ban } from 'lucide-react';
import { ApiError } from '../../api/client';

interface ErrorBannerProps {
  error: Error | null;
  className?: string;
}

function errorInfo(err: Error) {
  if (err instanceof ApiError) {
    switch (err.status) {
      case 403:
        return { icon: ShieldOff, title: 'Private Repository', message: err.message, color: 'text-yellow-400' };
      case 404:
        return { icon: Ban, title: 'Not Found', message: 'Repository not found. Check the owner and repo name.', color: 'text-zinc-400' };
      case 409:
        return { icon: Clock, title: 'Analysis In Progress', message: 'Data is still being generated. It will appear automatically when ready.', color: 'text-blue-400' };
      case 429:
        return { icon: Clock, title: 'Rate Limited', message: 'Too many requests. Please wait a moment and try again.', color: 'text-orange-400' };
      case 503:
        return { icon: ServerOff, title: 'AI Unavailable', message: 'AI features are not available. The Gemini API key may not be configured.', color: 'text-zinc-500' };
      default:
        return { icon: AlertTriangle, title: 'Error', message: err.message, color: 'text-red-400' };
    }
  }
  if (err.message?.includes('fetch') || err.message?.includes('network')) {
    return { icon: WifiOff, title: 'Connection Error', message: 'Unable to reach the server. Check your connection.', color: 'text-red-400' };
  }
  return { icon: AlertTriangle, title: 'Something went wrong', message: err.message, color: 'text-red-400' };
}

export default function ErrorBanner({ error, className }: ErrorBannerProps) {
  if (!error) return null;

  const { icon: Icon, title, message, color } = errorInfo(error);

  return (
    <div className={`flex items-start gap-3 rounded-lg border border-zinc-800 bg-zinc-900/80 px-4 py-3 ${className ?? ''}`}>
      <Icon className={`mt-0.5 h-5 w-5 shrink-0 ${color}`} />
      <div className="min-w-0">
        <p className={`text-sm font-medium ${color}`}>{title}</p>
        <p className="text-sm text-zinc-400">{message}</p>
      </div>
    </div>
  );
}
