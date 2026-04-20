import { useChurn } from '../../hooks/useChurn';
import ScatterPlot from './ScatterPlot';
import RiskZoneList from './RiskZoneList';
import StaleFiles from './StaleFiles';
import Skeleton from '../shared/Skeleton';
import ErrorBanner from '../shared/ErrorBanner';

interface ChurnTabProps {
  owner: string;
  repo: string;
}

export default function ChurnTab({ owner, repo }: ChurnTabProps) {
  const { data, error, isLoading } = useChurn(owner, repo);

  if (isLoading) return <Skeleton rows={6} height="h-8" />;
  if (error) return <ErrorBanner error={error} />;
  if (!data) return null;

  return (
    <div>
      <ScatterPlot riskMatrix={data.risk_matrix} thresholds={data.thresholds} />
      <RiskZoneList entries={data.risk_zone} />
      <StaleFiles files={data.stale_files} />
    </div>
  );
}
