import { useQuery } from '@tanstack/react-query';
import { fetchComplexity } from '../api/endpoints';
import { ApiError } from '../api/client';

export function useComplexity(owner: string, repo: string) {
  return useQuery({
    queryKey: ['complexity', owner, repo],
    queryFn: () => fetchComplexity(owner, repo),
    enabled: !!owner && !!repo,
    retry: (failureCount, error) => {
      if (error instanceof ApiError && [404, 409].includes(error.status)) return false;
      return failureCount < 2;
    },
  });
}
