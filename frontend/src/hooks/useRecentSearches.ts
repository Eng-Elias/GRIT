import { useState, useCallback } from 'react';

const STORAGE_KEY = 'grit:recent-searches';
const MAX_ITEMS = 5;

export interface RecentSearch {
  owner: string;
  repo: string;
  timestamp: number;
}

function load(): RecentSearch[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

function save(items: RecentSearch[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
}

export function useRecentSearches() {
  const [searches, setSearches] = useState<RecentSearch[]>(load);

  const add = useCallback((owner: string, repo: string) => {
    setSearches(prev => {
      const filtered = prev.filter(s => !(s.owner === owner && s.repo === repo));
      const next = [{ owner, repo, timestamp: Date.now() }, ...filtered].slice(0, MAX_ITEMS);
      save(next);
      return next;
    });
  }, []);

  const clear = useCallback(() => {
    localStorage.removeItem(STORAGE_KEY);
    setSearches([]);
  }, []);

  return { searches, add, clear };
}
