import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import React from 'react';
import { Activity } from 'lucide-react';
import { MetricCard } from '../components/MetricCard';

describe('MetricCard Component', () => {
  it('renders title, value, and subtext properly', () => {
    render(
      <MetricCard
        title="Active Alerts"
        value={12}
        subtext="4 critical"
        icon={Activity}
        badge={{ text: 'CRITICAL', variant: 'critical' }}
      />
    );

    expect(screen.getByText('Active Alerts')).toBeTruthy();
    expect(screen.getByText('12')).toBeTruthy();
    expect(screen.getByText('4 critical')).toBeTruthy();
    expect(screen.getByText('CRITICAL')).toBeTruthy();
  });
});
