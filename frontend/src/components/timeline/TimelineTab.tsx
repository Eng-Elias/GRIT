import { useTemporal } from '../../hooks/useTemporal';
import LOCChart from './LOCChart';
import VelocityChart from './VelocityChart';
import RefactorPeriods from './RefactorPeriods';
import Skeleton from '../shared/Skeleton';
import ErrorBanner from '../shared/ErrorBanner';

interface TimelineTabProps {
  owner: string;
  repo: string;
}

export default function TimelineTab({ owner, repo }: TimelineTabProps) {
  const { data, error, isLoading } = useTemporal(owner, repo);

  if (isLoading) return <Skeleton rows={6} height="h-8" />;
  if (error) return <ErrorBanner error={error} />;
  if (!data) return null;

  return (
    <div>
      <LOCChart snapshots={data.loc_over_time} />
      <VelocityChart weeks={data.velocity.weekly_activity} />
      <RefactorPeriods periods={data.refactor_periods} />
    </div>
  );
}
