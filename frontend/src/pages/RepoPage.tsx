import { useParams } from 'react-router-dom';

export default function RepoPage() {
  const { owner, repo } = useParams<{ owner: string; repo: string }>();

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <h1 className="text-2xl font-bold text-white">{owner}/{repo}</h1>
      <p className="text-zinc-400 mt-2">Placeholder — will be replaced in T031</p>
    </div>
  );
}
