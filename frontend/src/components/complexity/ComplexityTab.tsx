import { useComplexity } from '../../hooks/useComplexity';
import ComplexitySummary from './ComplexitySummary';
import DistributionBar from './DistributionBar';
import HotFilesTable from './HotFilesTable';
import Skeleton from '../shared/Skeleton';
import ErrorBanner from '../shared/ErrorBanner';

interface ComplexityTabProps {
  owner: string;
  repo: string;
}

export default function ComplexityTab({ owner, repo }: ComplexityTabProps) {
  const { data, error, isLoading } = useComplexity(owner, repo);

  if (isLoading) return <Skeleton rows={6} height="h-8" />;
  if (error) return <ErrorBanner error={error} />;
  if (!data) return null;

  return (
    <div>
      <ComplexitySummary data={data} />
      <DistributionBar distribution={data.distribution} />
      <HotFilesTable files={data.hot_files} />
    </div>
  );
}
