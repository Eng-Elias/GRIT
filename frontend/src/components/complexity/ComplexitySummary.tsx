import type { ComplexityResult } from '../../types/complexity';

interface ComplexitySummaryProps {
  data: ComplexityResult;
}

export default function ComplexitySummary({ data }: ComplexitySummaryProps) {
  const cards = [
    { label: 'Mean', value: data.mean_complexity.toFixed(1), color: 'text-blue-400' },
    { label: 'Median', value: data.median_complexity.toFixed(1), color: 'text-cyan-400' },
    { label: 'P90', value: data.p90_complexity.toFixed(1), color: 'text-orange-400' },
    { label: 'Functions', value: data.total_function_count.toLocaleString(), color: 'text-violet-400' },
    { label: 'Files Analyzed', value: data.total_files_analyzed.toLocaleString(), color: 'text-green-400' },
    { label: 'Hot Files', value: data.hot_files.length.toString(), color: 'text-red-400' },
  ];

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-6">
      {cards.map(({ label, value, color }) => (
        <div key={label} className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-3 text-center">
          <p className="text-xs text-zinc-500 mb-1">{label}</p>
          <p className={`text-xl font-bold ${color}`}>{value}</p>
        </div>
      ))}
    </div>
  );
}
