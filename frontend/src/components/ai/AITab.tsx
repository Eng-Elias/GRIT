import Skeleton from '../shared/Skeleton';

interface AITabProps {
  owner: string;
  repo: string;
}

export default function AITab({ owner, repo }: AITabProps) {
  void owner; void repo;
  return <Skeleton rows={5} height="h-8" />;
}
