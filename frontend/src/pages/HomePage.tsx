import { GitBranch } from 'lucide-react';
import SearchBar from '../components/home/SearchBar';
import RecentSearches from '../components/home/RecentSearches';
import ExampleRepos from '../components/home/ExampleRepos';
import { useRecentSearches } from '../hooks/useRecentSearches';

export default function HomePage() {
  const { searches, add, clear } = useRecentSearches();

  return (
    <div className="flex flex-1 flex-col items-center justify-center px-4 py-16 gap-10">
      <div className="text-center">
        <GitBranch className="h-12 w-12 text-violet-400 mx-auto mb-4" />
        <h1 className="text-4xl sm:text-5xl font-bold text-white tracking-tight">GRIT</h1>
        <p className="mt-2 text-zinc-400 text-lg max-w-md mx-auto">
          Analyze any GitHub repository — complexity, churn, contributors, timeline &amp; AI insights.
        </p>
      </div>

      <SearchBar onSearch={(owner, repo) => add(owner, repo)} />
      <RecentSearches searches={searches} onClear={clear} />
      <ExampleRepos />
    </div>
  );
}
