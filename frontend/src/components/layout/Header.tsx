import { Link } from 'react-router-dom';
import { GitBranch } from 'lucide-react';

export default function Header() {
  return (
    <header className="sticky top-0 z-50 border-b border-zinc-800 bg-zinc-950/80 backdrop-blur-sm">
      <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4">
        <Link to="/" className="flex items-center gap-2 text-white font-semibold text-lg hover:opacity-90 transition-opacity">
          <GitBranch className="h-5 w-5 text-violet-400" />
          GRIT
        </Link>
        <span className="text-xs text-zinc-600 hidden sm:block">Git Repo Intelligence Tool</span>
      </div>
    </header>
  );
}
