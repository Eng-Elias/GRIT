import { useQuery } from '@tanstack/react-query';
import { fetchChurnMatrix } from '../api/endpoints';
import { ApiError } from '../api/client';

export function useChurn(owner: string, repo: string) {
  return useQuery({
    queryKey: ['churn', owner, repo],
    queryFn: () => fetchChurnMatrix(owner, repo),
    enabled: !!owner && !!repo,
    retry: (failureCount, error) => {
      if (error instanceof ApiError && [404, 409].includes(error.status)) return false;
      return failureCount < 2;
    },
  });
}
