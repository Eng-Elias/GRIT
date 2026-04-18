import Skeleton from '../shared/Skeleton';

interface ContributorsTabProps {
  owner: string;
  repo: string;
}

export default function ContributorsTab({ owner, repo }: ContributorsTabProps) {
  void owner; void repo;
  return <Skeleton rows={5} height="h-8" />;
}
