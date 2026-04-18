import { useQuery } from '@tanstack/react-query';
import { fetchAnalysis } from '../api/endpoints';
import { ApiError } from '../api/client';

export function useAnalysis(owner: string, repo: string) {
  return useQuery({
    queryKey: ['analysis', owner, repo],
    queryFn: () => fetchAnalysis(owner, repo),
    enabled: !!owner && !!repo,
    retry: (failureCount, error) => {
      if (error instanceof ApiError && [404, 403].includes(error.status)) return false;
      return failureCount < 2;
    },
  });
}
