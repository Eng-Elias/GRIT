import Skeleton from '../shared/Skeleton';

interface TimelineTabProps {
  owner: string;
  repo: string;
}

export default function TimelineTab({ owner, repo }: TimelineTabProps) {
  void owner; void repo;
  return <Skeleton rows={5} height="h-8" />;
}
