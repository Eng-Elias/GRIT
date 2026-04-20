import AISummaryPanel from './AISummaryPanel';
import HealthGauge from './HealthGauge';
import ChatPanel from './ChatPanel';

interface AITabProps {
  owner: string;
  repo: string;
}

export default function AITab({ owner, repo }: AITabProps) {
  return (
    <div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <AISummaryPanel owner={owner} repo={repo} />
        <HealthGauge owner={owner} repo={repo} />
      </div>
      <ChatPanel owner={owner} repo={repo} />
    </div>
  );
}
