import { lazy, Suspense } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { useAnalysis } from '../hooks/useAnalysis';
import { useStatus } from '../hooks/useStatus';
import { ApiError } from '../api/client';
import RepoHeader from '../components/repo/RepoHeader';
import LanguageBar from '../components/repo/LanguageBar';
import AnalysisStatus from '../components/repo/AnalysisStatus';
import TabNav, { type TabId } from '../components/layout/TabNav';
import OverviewTab from '../components/overview/OverviewTab';
import Skeleton from '../components/shared/Skeleton';
import ErrorBanner from '../components/shared/ErrorBanner';

const ComplexityTab = lazy(() => import('../components/complexity/ComplexityTab'));
const ChurnTab = lazy(() => import('../components/churn/ChurnTab'));
const ContributorsTab = lazy(() => import('../components/contributors/ContributorsTab'));
const TimelineTab = lazy(() => import('../components/timeline/TimelineTab'));
const AITab = lazy(() => import('../components/ai/AITab'));

export default function RepoPage() {
  const { owner = '', repo = '' } = useParams<{ owner: string; repo: string }>();
  const [searchParams] = useSearchParams();
  const tab = (searchParams.get('tab') ?? 'overview') as TabId;

  const analysis = useAnalysis(owner, repo);
  const needsPolling = analysis.error instanceof ApiError && analysis.error.status === 202;
  const status = useStatus(owner, repo, needsPolling || (!analysis.data && !analysis.error));

  if (analysis.error && !(analysis.error instanceof ApiError && analysis.error.status === 202)) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-8">
        <ErrorBanner error={analysis.error} />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      {status.data && <AnalysisStatus job={status.data} />}

      {!analysis.data && !analysis.error && (
        <div className="space-y-4">
          <Skeleton height="h-8" rows={1} />
          <Skeleton height="h-4" rows={3} />
        </div>
      )}

      {analysis.data && (
        <>
          <RepoHeader data={analysis.data} />
          <LanguageBar languages={analysis.data.languages} />
          <TabNav active={tab} />

          <Suspense fallback={<Skeleton rows={6} height="h-6" />}>
            {tab === 'overview' && <OverviewTab data={analysis.data} />}
            {tab === 'complexity' && <ComplexityTab owner={owner} repo={repo} />}
            {tab === 'churn' && <ChurnTab owner={owner} repo={repo} />}
            {tab === 'contributors' && <ContributorsTab owner={owner} repo={repo} />}
            {tab === 'timeline' && <TimelineTab owner={owner} repo={repo} />}
            {tab === 'ai' && <AITab owner={owner} repo={repo} />}
          </Suspense>
        </>
      )}
    </div>
  );
}
