import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search } from 'lucide-react';

interface SearchBarProps {
  onSearch?: (owner: string, repo: string) => void;
}

const REPO_REGEX = /^[a-zA-Z0-9_.-]+\/[a-zA-Z0-9_.-]+$/;

export default function SearchBar({ onSearch }: SearchBarProps) {
  const [value, setValue] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const trimmed = value.trim();

    if (!trimmed) {
      setError('Enter a repository in owner/repo format');
      return;
    }
    if (!REPO_REGEX.test(trimmed)) {
      setError('Invalid format. Use owner/repo (e.g., facebook/react)');
      return;
    }

    setError('');
    const [owner, repo] = trimmed.split('/');
    onSearch?.(owner, repo);
    navigate(`/repo/${owner}/${repo}`);
  }

  return (
    <form onSubmit={handleSubmit} className="w-full max-w-xl">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-zinc-500" />
        <input
          type="text"
          value={value}
          onChange={e => { setValue(e.target.value); setError(''); }}
          placeholder="owner/repo (e.g., facebook/react)"
          className="w-full rounded-lg border border-zinc-700 bg-zinc-900 py-3 pl-10 pr-4 text-white placeholder-zinc-500 outline-none transition focus:border-violet-500 focus:ring-1 focus:ring-violet-500"
          autoFocus
        />
      </div>
      {error && <p className="mt-2 text-sm text-red-400">{error}</p>}
    </form>
  );
}
