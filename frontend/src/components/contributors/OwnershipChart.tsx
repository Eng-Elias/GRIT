import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer } from 'recharts';
import type { Author } from '../../types/contributors';

interface OwnershipChartProps {
  authors: Author[];
}

const COLORS = ['#a78bfa', '#7c3aed', '#4f46e5', '#3b82f6', '#06b6d4', '#14b8a6', '#22c55e', '#eab308'];

export default function OwnershipChart({ authors }: OwnershipChartProps) {
  const top = authors.slice(0, 7);
  const othersPercent = authors.slice(7).reduce((sum, a) => sum + a.ownership_percent, 0);

  const data = [
    ...top.map(a => ({ name: a.name, value: Math.round(a.ownership_percent * 10) / 10 })),
    ...(othersPercent > 0 ? [{ name: 'Others', value: Math.round(othersPercent * 10) / 10 }] : []),
  ];

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4">
      <h3 className="text-sm font-medium text-zinc-400 mb-3">Code Ownership</h3>
      <ResponsiveContainer width="100%" height={260}>
        <PieChart>
          <Pie
            data={data}
            dataKey="value"
            nameKey="name"
            cx="50%"
            cy="50%"
            innerRadius={55}
            outerRadius={90}
            paddingAngle={2}
            stroke="#18181b"
            strokeWidth={2}
          >
            {data.map((_, i) => (
              <Cell key={i} fill={COLORS[i % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{ backgroundColor: '#27272a', border: '1px solid #3f3f46', borderRadius: '8px', fontSize: '12px' }}
            itemStyle={{ color: '#e4e4e7' }}
            formatter={(value) => `${value}%`}
          />
        </PieChart>
      </ResponsiveContainer>
      <div className="flex flex-wrap gap-x-4 gap-y-1 mt-2 justify-center">
        {data.map((d, i) => (
          <span key={d.name} className="flex items-center gap-1.5 text-xs text-zinc-400">
            <span className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: COLORS[i % COLORS.length] }} />
            {d.name} <span className="text-zinc-600">{d.value}%</span>
          </span>
        ))}
      </div>
    </div>
  );
}
