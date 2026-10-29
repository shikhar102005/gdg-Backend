import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import React from 'react';
import { TrendsChart } from '../components/TrendsChart';
import { TrendPoint } from '../types';

describe('TrendsChart Component', () => {
  it('renders empty message when no points provided', () => {
    render(<TrendsChart points={[]} metric="temperature" unit="°C" />);
    expect(screen.getByText(/No telemetry readings available/i)).toBeTruthy();
  });

  it('renders SVG polyline and coordinates when points provided', () => {
    const samplePoints: TrendPoint[] = [
      { timestamp: '2026-10-26T04:00:00Z', avg: 22.5, min: 21.0, max: 24.0, count: 10 },
      { timestamp: '2026-10-26T04:05:00Z', avg: 23.0, min: 22.0, max: 25.0, count: 12 },
      { timestamp: '2026-10-26T04:10:00Z', avg: 24.1, min: 23.0, max: 26.0, count: 15 },
    ];

    const { container } = render(<TrendsChart points={samplePoints} metric="temperature" unit="°C" />);
    const svg = container.querySelector('svg');
    expect(svg).toBeTruthy();

    const circles = container.querySelectorAll('circle');
    expect(circles.length).toBeGreaterThan(0);
  });
});
