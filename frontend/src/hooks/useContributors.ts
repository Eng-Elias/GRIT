import { useQuery } from '@tanstack/react-query';
import { fetchContributors } from '../api/endpoints';
import { ApiError } from '../api/client';

export function useContributors(owner: string, repo: string) {
  return useQuery({
    queryKey: ['contributors', owner, repo],
    queryFn: () => fetchContributors(owner, repo),
    enabled: !!owner && !!repo,
    retry: (failureCount, error) => {
      if (error instanceof ApiError && [404, 409].includes(error.status)) return false;
      return failureCount < 2;
    },
  });
}
