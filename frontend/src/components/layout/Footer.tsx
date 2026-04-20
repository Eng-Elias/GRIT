import { GitBranch } from 'lucide-react';

export default function Footer() {
  return (
    <footer className="border-t border-zinc-800 bg-zinc-950 py-4">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-4">
        <span className="flex items-center gap-1.5 text-xs text-zinc-600">
          <GitBranch className="h-3 w-3" /> GRIT — Git Repo Intelligence Tool
        </span>
        <a
          href="https://github.com/Eng-Elias/GRIT"
          target="_blank"
          rel="noopener noreferrer"
          className="text-xs text-zinc-600 hover:text-zinc-400 transition-colors"
        >
          Source
        </a>
      </div>
    </footer>
  );
}
