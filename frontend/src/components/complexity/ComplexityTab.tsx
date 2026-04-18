import Skeleton from '../shared/Skeleton';

interface ComplexityTabProps {
  owner: string;
  repo: string;
}

export default function ComplexityTab({ owner, repo }: ComplexityTabProps) {
  void owner; void repo;
  return <Skeleton rows={5} height="h-8" />;
}
