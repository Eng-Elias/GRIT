import { useContributors } from '../../hooks/useContributors';
import BusFactorGauge from './BusFactorGauge';
import OwnershipChart from './OwnershipChart';
import KeyPeople from './KeyPeople';
import AuthorTable from './AuthorTable';
import Skeleton from '../shared/Skeleton';
import ErrorBanner from '../shared/ErrorBanner';

interface ContributorsTabProps {
  owner: string;
  repo: string;
}

export default function ContributorsTab({ owner, repo }: ContributorsTabProps) {
  const { data, error, isLoading } = useContributors(owner, repo);

  if (isLoading) return <Skeleton rows={6} height="h-8" />;
  if (error) return <ErrorBanner error={error} />;
  if (!data) return null;

  return (
    <div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
        <BusFactorGauge busFactor={data.bus_factor} />
        <OwnershipChart authors={data.authors} />
        <KeyPeople keyPeople={data.key_people} />
      </div>
      <AuthorTable authors={data.authors} />
    </div>
  );
}
