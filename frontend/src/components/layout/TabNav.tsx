import { useSearchParams } from 'react-router-dom';
import clsx from 'clsx';
import { BarChart3, Bug, Users, Clock, Sparkles, LayoutDashboard } from 'lucide-react';

export type TabId = 'overview' | 'complexity' | 'churn' | 'contributors' | 'timeline' | 'ai';

interface Tab {
  id: TabId;
  label: string;
  icon: React.ElementType;
}

const TABS: Tab[] = [
  { id: 'overview', label: 'Overview', icon: LayoutDashboard },
  { id: 'complexity', label: 'Complexity', icon: BarChart3 },
  { id: 'churn', label: 'Churn', icon: Bug },
  { id: 'contributors', label: 'Contributors', icon: Users },
  { id: 'timeline', label: 'Timeline', icon: Clock },
  { id: 'ai', label: 'AI', icon: Sparkles },
];

interface TabNavProps {
  active: TabId;
}

export default function TabNav({ active }: TabNavProps) {
  const [, setSearchParams] = useSearchParams();

  return (
    <nav className="flex gap-1 overflow-x-auto border-b border-zinc-800 mb-6 pb-px scrollbar-none">
      {TABS.map(({ id, label, icon: Icon }) => (
        <button
          key={id}
          onClick={() => setSearchParams({ tab: id })}
          className={clsx(
            'flex items-center gap-1.5 whitespace-nowrap rounded-t-md px-3 py-2 text-sm font-medium transition-colors',
            active === id
              ? 'border-b-2 border-violet-500 text-white'
              : 'text-zinc-500 hover:text-zinc-300',
          )}
        >
          <Icon className="h-4 w-4" />
          {label}
        </button>
      ))}
    </nav>
  );
}
