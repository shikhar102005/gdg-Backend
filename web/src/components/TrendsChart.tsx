import React, { useState } from 'react';
import { TrendPoint } from '../types';

interface TrendsChartProps {
  points: TrendPoint[];
  metric: string;
  unit: string;
}

export const TrendsChart: React.FC<TrendsChartProps> = ({ points, metric, unit }) => {
  const [hoverIndex, setHoverIndex] = useState<number | null>(null);

  const safePoints = (points || []).filter((p) => p && typeof p.avg === 'number' && !isNaN(p.avg));

  if (safePoints.length === 0) {
    return (
      <div style={{ height: '220px', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#5f6368', fontSize: '0.875rem', fontWeight: 700 }}>
        No telemetry readings available in this window. (Run simulator to generate live data)
      </div>
    );
  }

  const values = safePoints.map((p) => p.avg);
  const rawMin = Math.min(...values);
  const rawMax = Math.max(...values);
  const minVal = Math.floor(rawMin * 0.95);
  const maxVal = Math.ceil(rawMax * 1.05) || (minVal + 10);
  const valRange = maxVal - minVal || 1;

  const svgWidth = 800;
  const svgHeight = 220;
  const padLeft = 45;
  const padRight = 20;
  const padTop = 20;
  const padBottom = 30;

  const plotWidth = svgWidth - padLeft - padRight;
  const plotHeight = svgHeight - padTop - padBottom;

  const coords = safePoints.map((p, idx) => {
    const x = padLeft + (idx / (safePoints.length - 1 || 1)) * plotWidth;
    const y = padTop + plotHeight - ((p.avg - minVal) / valRange) * plotHeight;
    return { x, y, point: p };
  });

  const linePath = coords.reduce((acc, c, i) => `${acc} ${i === 0 ? 'M' : 'L'} ${c.x.toFixed(1)} ${c.y.toFixed(1)}`, '');
  const areaPath = coords.length > 0
    ? `${linePath} L ${coords[coords.length - 1].x.toFixed(1)} ${(padTop + plotHeight).toFixed(1)} L ${coords[0].x.toFixed(1)} ${(padTop + plotHeight).toFixed(1)} Z`
    : '';

  // Average line
  const avgTotal = values.reduce((a, b) => a + b, 0) / (values.length || 1);
  const avgY = padTop + plotHeight - ((avgTotal - minVal) / valRange) * plotHeight;

  return (
    <div style={{ position: 'relative', width: '100%', overflowX: 'auto' }}>
      <svg viewBox={`0 0 ${svgWidth} ${svgHeight}`} style={{ width: '100%', height: 'auto', display: 'block' }}>
        <defs>
          <linearGradient id="retroChartGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#d2e3fc" stopOpacity="0.8" />
            <stop offset="100%" stopColor="#e8f0fe" stopOpacity="0.1" />
          </linearGradient>
        </defs>

        {/* Grid lines */}
        {[0, 0.25, 0.5, 0.75, 1].map((pct, idx) => {
          const y = padTop + plotHeight * pct;
          const labelVal = (maxVal - pct * valRange).toFixed(1);
          return (
            <g key={idx}>
              <line x1={padLeft} y1={y} x2={svgWidth - padRight} y2={y} stroke="#202124" strokeWidth="1" strokeDasharray="3 3" opacity="0.15" />
              <text x={padLeft - 8} y={y + 4} textAnchor="end" fontSize="11" fontWeight="700" fill="#5f6368" fontFamily="sans-serif">
                {labelVal}
              </text>
            </g>
          );
        })}

        {/* Average reference line in Google Yellow */}
        <line x1={padLeft} y1={avgY} x2={svgWidth - padRight} y2={avgY} stroke="#fbbc04" strokeDasharray="4 2" strokeWidth="2.5" />

        {/* Filled Area */}
        {areaPath && <path d={areaPath} fill="url(#retroChartGradient)" />}

        {/* Trend Polyline in Google Blue with thick stroke */}
        {linePath && <path d={linePath} fill="none" stroke="#4285f4" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" />}

        {/* Coordinates Points & Interactive Hover Area */}
        {coords.map((c, i) => (
          <g key={i}>
            <circle
              cx={c.x}
              cy={c.y}
              r={hoverIndex === i ? 6 : 3.5}
              fill="#ffffff"
              stroke="#1a73e8"
              strokeWidth="2.5"
            />
            {/* Target overlay */}
            <rect
              x={c.x - 15}
              y={padTop}
              width={30}
              height={plotHeight}
              fill="transparent"
              style={{ cursor: 'pointer' }}
              onMouseEnter={() => setHoverIndex(i)}
              onMouseLeave={() => setHoverIndex(null)}
            />
          </g>
        ))}
      </svg>

      {/* Tooltip Overlay */}
      {hoverIndex !== null && coords[hoverIndex] && (
        <div style={{
          position: 'absolute',
          left: `${(coords[hoverIndex].x / svgWidth) * 100}%`,
          top: `${(coords[hoverIndex].y / svgHeight) * 100}%`,
          transform: 'translate(-50%, -120%)',
          background: '#ffffff',
          border: '2px solid #202124',
          borderRadius: '8px',
          padding: '0.35rem 0.65rem',
          fontSize: '0.75rem',
          fontWeight: 700,
          boxShadow: '2px 2px 0px #202124',
          pointerEvents: 'none',
          whiteSpace: 'nowrap',
          zIndex: 10,
        }}>
          <div>Avg: <strong>{coords[hoverIndex].point.avg.toFixed(2)} {unit}</strong></div>
          <div style={{ color: '#5f6368', fontSize: '0.7rem' }}>
            Min: {coords[hoverIndex].point.min.toFixed(1)} &bull; Max: {coords[hoverIndex].point.max.toFixed(1)}
          </div>
        </div>
      )}
    </div>
  );
};
export default TrendsChart;
