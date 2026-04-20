import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import type { MonthlySnapshot } from '../../types/temporal';

interface LOCChartProps {
  snapshots: MonthlySnapshot[];
}

export default function LOCChart({ snapshots }: LOCChartProps) {
  if (snapshots.length === 0) return null;

  const data = snapshots.map(s => ({
    date: new Date(s.date).toLocaleDateString('en-US', { year: '2-digit', month: 'short' }),
    loc: s.total_loc,
  }));

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4 mb-6">
      <h3 className="text-sm font-medium text-zinc-400 mb-3">Lines of Code Over Time</h3>
      <ResponsiveContainer width="100%" height={280}>
        <AreaChart data={data}>
          <defs>
            <linearGradient id="locGrad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#a78bfa" stopOpacity={0.3} />
              <stop offset="95%" stopColor="#a78bfa" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
          <XAxis dataKey="date" tick={{ fill: '#71717a', fontSize: 11 }} tickLine={false} axisLine={{ stroke: '#3f3f46' }} />
          <YAxis tick={{ fill: '#71717a', fontSize: 11 }} tickLine={false} axisLine={{ stroke: '#3f3f46' }} tickFormatter={v => v >= 1000 ? `${(v / 1000).toFixed(0)}k` : v} />
          <Tooltip
            contentStyle={{ backgroundColor: '#27272a', border: '1px solid #3f3f46', borderRadius: '8px', fontSize: '12px' }}
            itemStyle={{ color: '#e4e4e7' }}
            labelStyle={{ color: '#a1a1aa' }}
            formatter={(value) => [Number(value).toLocaleString(), 'LOC']}
          />
          <Area type="monotone" dataKey="loc" stroke="#a78bfa" strokeWidth={2} fill="url(#locGrad)" />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
