import clsx from 'clsx';

interface RiskBadgeProps {
  level: string;
}

const colors: Record<string, string> = {
  low: 'bg-green-900/50 text-green-400 border-green-800',
  medium: 'bg-yellow-900/50 text-yellow-400 border-yellow-800',
  high: 'bg-orange-900/50 text-orange-400 border-orange-800',
  critical: 'bg-red-900/50 text-red-400 border-red-800',
};

export default function RiskBadge({ level }: RiskBadgeProps) {
  const normalized = level.toLowerCase();
  return (
    <span
      className={clsx(
        'inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium capitalize',
        colors[normalized] ?? 'bg-zinc-800 text-zinc-400 border-zinc-700',
      )}
    >
      {normalized}
    </span>
  );
}
