import { useQuery } from '@tanstack/react-query';
import { fetchTemporal } from '../api/endpoints';
import { ApiError } from '../api/client';

export function useTemporal(owner: string, repo: string) {
  return useQuery({
    queryKey: ['temporal', owner, repo],
    queryFn: () => fetchTemporal(owner, repo),
    enabled: !!owner && !!repo,
    retry: (failureCount, error) => {
      if (error instanceof ApiError && [404, 409].includes(error.status)) return false;
      return failureCount < 2;
    },
  });
}
