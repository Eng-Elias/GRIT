import { useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchStatus } from '../api/endpoints';
import { ApiError } from '../api/client';

export function useStatus(owner: string, repo: string, enabled: boolean) {
  const queryClient = useQueryClient();

  return useQuery({
    queryKey: ['status', owner, repo],
    queryFn: () => fetchStatus(owner, repo),
    enabled: enabled && !!owner && !!repo,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      if (status === 'completed' || status === 'failed') return false;
      return 3000;
    },
    retry: (failureCount, error) => {
      if (error instanceof ApiError && error.status === 404) return false;
      return failureCount < 2;
    },
    meta: {
      onSettled: (_data: unknown) => {
        const data = _data as { status?: string } | undefined;
        if (data?.status === 'completed') {
          queryClient.invalidateQueries({ queryKey: ['analysis', owner, repo] });
        }
      },
    },
  });
}
