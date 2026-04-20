import { useRef, useEffect, useState } from 'react';
import * as d3 from 'd3';
import type { RiskEntry, Thresholds } from '../../types/churn';

interface ScatterPlotProps {
  riskMatrix: RiskEntry[];
  thresholds: Thresholds;
}

const LANG_COLORS: Record<string, string> = {
  JavaScript: '#f1e05a', TypeScript: '#3178c6', Python: '#3572a5', Go: '#00add8',
  Java: '#b07219', Ruby: '#701516', Rust: '#dea584', C: '#555555', 'C++': '#f34b7d',
  'C#': '#178600', PHP: '#4f5d95', Swift: '#f05138', Kotlin: '#a97bff', Shell: '#89e051',
};

const MARGIN = { top: 20, right: 20, bottom: 50, left: 60 };

export default function ScatterPlot({ riskMatrix, thresholds }: ScatterPlotProps) {
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [tooltip, setTooltip] = useState<{ x: number; y: number; entry: RiskEntry } | null>(null);

  useEffect(() => {
    if (!svgRef.current || !containerRef.current || riskMatrix.length === 0) return;

    const containerWidth = containerRef.current.clientWidth;
    const width = containerWidth;
    const height = Math.min(width * 0.6, 500);
    const innerW = width - MARGIN.left - MARGIN.right;
    const innerH = height - MARGIN.top - MARGIN.bottom;

    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();
    svg.attr('width', width).attr('height', height);

    const g = svg.append('g').attr('transform', `translate(${MARGIN.left},${MARGIN.top})`);

    const xMax = d3.max(riskMatrix, d => d.churn) ?? 10;
    const yMax = d3.max(riskMatrix, d => d.complexity_cyclomatic) ?? 10;

    const x = d3.scaleLinear().domain([0, xMax * 1.1]).range([0, innerW]);
    const y = d3.scaleLinear().domain([0, yMax * 1.1]).range([innerH, 0]);

    // Risk quadrant shading
    const churnThresh = thresholds.churn_p75;
    const compThresh = thresholds.complexity_p75;
    if (churnThresh > 0 && compThresh > 0) {
      g.append('rect')
        .attr('x', x(churnThresh))
        .attr('y', 0)
        .attr('width', Math.max(0, innerW - x(churnThresh)))
        .attr('height', Math.max(0, y(compThresh)))
        .attr('fill', '#ef4444')
        .attr('opacity', 0.08);
    }

    // P75 reference lines
    if (churnThresh > 0) {
      g.append('line')
        .attr('x1', x(churnThresh)).attr('x2', x(churnThresh))
        .attr('y1', 0).attr('y2', innerH)
        .attr('stroke', '#ef4444').attr('stroke-dasharray', '4,4').attr('opacity', 0.4);
    }
    if (compThresh > 0) {
      g.append('line')
        .attr('x1', 0).attr('x2', innerW)
        .attr('y1', y(compThresh)).attr('y2', y(compThresh))
        .attr('stroke', '#ef4444').attr('stroke-dasharray', '4,4').attr('opacity', 0.4);
    }

    // Axes
    g.append('g')
      .attr('transform', `translate(0,${innerH})`)
      .call(d3.axisBottom(x).ticks(6))
      .call(g => g.selectAll('text').attr('fill', '#71717a').attr('font-size', '11px'))
      .call(g => g.selectAll('line, path').attr('stroke', '#3f3f46'));

    g.append('g')
      .call(d3.axisLeft(y).ticks(6))
      .call(g => g.selectAll('text').attr('fill', '#71717a').attr('font-size', '11px'))
      .call(g => g.selectAll('line, path').attr('stroke', '#3f3f46'));

    // Axis labels
    svg.append('text')
      .attr('x', width / 2).attr('y', height - 8)
      .attr('text-anchor', 'middle').attr('fill', '#a1a1aa').attr('font-size', '12px')
      .text('Churn (commit touches)');

    svg.append('text')
      .attr('transform', `rotate(-90)`)
      .attr('x', -height / 2).attr('y', 16)
      .attr('text-anchor', 'middle').attr('fill', '#a1a1aa').attr('font-size', '12px')
      .text('Cyclomatic Complexity');

    // Dots
    g.selectAll('circle')
      .data(riskMatrix)
      .join('circle')
      .attr('cx', d => x(d.churn))
      .attr('cy', d => y(d.complexity_cyclomatic))
      .attr('r', d => Math.min(3 + Math.sqrt(d.loc) * 0.15, 12))
      .attr('fill', d => LANG_COLORS[d.language] ?? '#6b7280')
      .attr('opacity', 0.75)
      .attr('stroke', '#18181b')
      .attr('stroke-width', 1)
      .style('cursor', 'pointer')
      .on('mouseenter', function (event, d) {
        d3.select(this).attr('opacity', 1).attr('stroke', '#fff').attr('stroke-width', 2);
        const rect = svgRef.current!.getBoundingClientRect();
        setTooltip({ x: event.clientX - rect.left, y: event.clientY - rect.top, entry: d });
      })
      .on('mouseleave', function () {
        d3.select(this).attr('opacity', 0.75).attr('stroke', '#18181b').attr('stroke-width', 1);
        setTooltip(null);
      });
  }, [riskMatrix, thresholds]);

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4 mb-6">
      <h3 className="text-sm font-medium text-zinc-400 mb-3">Churn × Complexity Scatter Plot</h3>
      <div ref={containerRef} className="relative w-full">
        <svg ref={svgRef} className="w-full" />
        {tooltip && (
          <div
            className="absolute z-10 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-xs shadow-lg pointer-events-none"
            style={{ left: tooltip.x + 12, top: tooltip.y - 10 }}
          >
            <p className="font-medium text-white truncate max-w-[200px]">{tooltip.entry.path}</p>
            <p className="text-zinc-400">Churn: {tooltip.entry.churn} · Complexity: {tooltip.entry.complexity_cyclomatic}</p>
            <p className="text-zinc-400">LOC: {tooltip.entry.loc} · {tooltip.entry.language}</p>
            <p className={`font-medium capitalize ${
              tooltip.entry.risk_level === 'critical' ? 'text-red-400' :
              tooltip.entry.risk_level === 'high' ? 'text-orange-400' :
              tooltip.entry.risk_level === 'medium' ? 'text-yellow-400' : 'text-green-400'
            }`}>
              {tooltip.entry.risk_level}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
