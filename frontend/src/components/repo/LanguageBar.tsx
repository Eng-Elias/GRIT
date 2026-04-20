import type { LanguageBreakdown } from '../../types/analysis';

interface LanguageBarProps {
  languages: LanguageBreakdown[];
}

const COLORS: Record<string, string> = {
  JavaScript: '#f1e05a',
  TypeScript: '#3178c6',
  Python: '#3572a5',
  Go: '#00add8',
  Java: '#b07219',
  Ruby: '#701516',
  Rust: '#dea584',
  C: '#555555',
  'C++': '#f34b7d',
  'C#': '#178600',
  PHP: '#4f5d95',
  Swift: '#f05138',
  Kotlin: '#a97bff',
  Dart: '#00b4ab',
  Shell: '#89e051',
  HTML: '#e34c26',
  CSS: '#563d7c',
  Markdown: '#083fa1',
  YAML: '#cb171e',
  JSON: '#292929',
};

function getColor(lang: string) {
  return COLORS[lang] ?? '#6b7280';
}

export default function LanguageBar({ languages }: LanguageBarProps) {
  if (!languages.length) return null;

  const total = languages.reduce((sum, l) => sum + l.code_lines, 0);
  if (total === 0) return null;

  return (
    <div className="mb-6">
      <div className="flex h-2.5 overflow-hidden rounded-full bg-zinc-800">
        {languages.map(l => {
          const pct = (l.code_lines / total) * 100;
          if (pct < 0.5) return null;
          return (
            <div
              key={l.language}
              className="transition-all duration-300"
              style={{ width: `${pct}%`, backgroundColor: getColor(l.language) }}
              title={`${l.language}: ${pct.toFixed(1)}%`}
            />
          );
        })}
      </div>
      <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1">
        {languages.filter(l => (l.code_lines / total) * 100 >= 1).map(l => (
          <span key={l.language} className="flex items-center gap-1.5 text-xs text-zinc-400">
            <span className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: getColor(l.language) }} />
            {l.language} <span className="text-zinc-600">{((l.code_lines / total) * 100).toFixed(1)}%</span>
          </span>
        ))}
      </div>
    </div>
  );
}
