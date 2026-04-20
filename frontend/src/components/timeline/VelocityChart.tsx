import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, ReferenceLine } from 'recharts';
import type { WeeklyActivity } from '../../types/temporal';

interface VelocityChartProps {
  weeks: WeeklyActivity[];
}

export default function VelocityChart({ weeks }: VelocityChartProps) {
  if (weeks.length === 0) return null;

  const data = weeks.slice(-26).map(w => ({
    week: new Date(w.week_start).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    additions: w.additions,
    deletions: -w.deletions,
    commits: w.commits,
  }));

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4 mb-6">
      <h3 className="text-sm font-medium text-zinc-400 mb-3">Weekly Velocity (Last 26 Weeks)</h3>
      <ResponsiveContainer width="100%" height={280}>
        <BarChart data={data} stackOffset="sign">
          <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
          <XAxis dataKey="week" tick={{ fill: '#71717a', fontSize: 11 }} tickLine={false} axisLine={{ stroke: '#3f3f46' }} interval={3} />
          <YAxis tick={{ fill: '#71717a', fontSize: 11 }} tickLine={false} axisLine={{ stroke: '#3f3f46' }} tickFormatter={v => {
            const abs = Math.abs(v);
            return abs >= 1000 ? `${(abs / 1000).toFixed(0)}k` : String(abs);
          }} />
          <Tooltip
            contentStyle={{ backgroundColor: '#27272a', border: '1px solid #3f3f46', borderRadius: '8px', fontSize: '12px' }}
            itemStyle={{ color: '#e4e4e7' }}
            labelStyle={{ color: '#a1a1aa' }}
            formatter={(value, name) => [Math.abs(Number(value)).toLocaleString(), name === 'deletions' ? 'Deletions' : name === 'additions' ? 'Additions' : 'Commits']}
          />
          <ReferenceLine y={0} stroke="#3f3f46" />
          <Bar dataKey="additions" fill="#22c55e" stackId="stack" radius={[2, 2, 0, 0]} />
          <Bar dataKey="deletions" fill="#ef4444" stackId="stack" radius={[0, 0, 2, 2]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
