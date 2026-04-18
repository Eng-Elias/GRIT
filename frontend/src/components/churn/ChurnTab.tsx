import Skeleton from '../shared/Skeleton';

interface ChurnTabProps {
  owner: string;
  repo: string;
}

export default function ChurnTab({ owner, repo }: ChurnTabProps) {
  void owner; void repo;
  return <Skeleton rows={5} height="h-8" />;
}
