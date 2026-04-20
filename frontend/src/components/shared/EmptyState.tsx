import type { LucideIcon } from 'lucide-react';
import { Inbox } from 'lucide-react';

interface EmptyStateProps {
  icon?: LucideIcon;
  title: string;
  description?: string;
  className?: string;
}

export default function EmptyState({ icon: Icon = Inbox, title, description, className }: EmptyStateProps) {
  return (
    <div className={`flex flex-col items-center justify-center py-16 text-center ${className ?? ''}`}>
      <Icon className="h-12 w-12 text-zinc-700 mb-4" />
      <h3 className="text-lg font-medium text-zinc-400">{title}</h3>
      {description && <p className="mt-1 text-sm text-zinc-500 max-w-md">{description}</p>}
    </div>
  );
}
