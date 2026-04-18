import type { AnalysisResult } from '../../types/analysis';
import StatsCards from './StatsCards';
import CommitHeatmap from './CommitHeatmap';
import HealthSignals from './HealthSignals';

interface OverviewTabProps {
  data: AnalysisResult;
}

export default function OverviewTab({ data }: OverviewTabProps) {
  return (
    <div>
      <StatsCards data={data} />
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <CommitHeatmap activity={data.commit_activity} />
        </div>
        <div>
          <HealthSignals metadata={data.metadata} />
        </div>
      </div>
    </div>
  );
}
