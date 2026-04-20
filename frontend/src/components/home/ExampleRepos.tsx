import { Link } from 'react-router-dom';
import { ArrowRight } from 'lucide-react';

const EXAMPLES = [
  { owner: 'facebook', repo: 'react', desc: 'A JavaScript library for building user interfaces' },
  { owner: 'golang', repo: 'go', desc: 'The Go programming language' },
  { owner: 'torvalds', repo: 'linux', desc: 'Linux kernel source tree' },
];

export default function ExampleRepos() {
  return (
    <div className="w-full max-w-xl">
      <span className="text-xs text-zinc-500 uppercase tracking-wider mb-2 block">Try an example</span>
      <div className="space-y-2">
        {EXAMPLES.map(({ owner, repo, desc }) => (
          <Link
            key={`${owner}/${repo}`}
            to={`/repo/${owner}/${repo}`}
            className="flex items-center justify-between rounded-lg border border-zinc-800 bg-zinc-900/50 px-4 py-3 hover:border-zinc-700 hover:bg-zinc-900 transition-colors group"
          >
            <div className="min-w-0">
              <p className="text-sm font-medium text-white">{owner}/{repo}</p>
              <p className="text-xs text-zinc-500 truncate">{desc}</p>
            </div>
            <ArrowRight className="h-4 w-4 text-zinc-600 group-hover:text-violet-400 transition-colors shrink-0 ml-3" />
          </Link>
        ))}
      </div>
    </div>
  );
}
