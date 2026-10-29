import React from 'react';
import { Activity, ArrowRight, Cpu, Layers } from 'lucide-react';

interface HeroSectionProps {
  onStartClick?: () => void;
  onSelectTab?: (tab: string) => void;
  onTriggerSpike?: (metric?: string, sensorId?: string, value?: number) => void;
  totalSensors: number;
  totalReadings: number;
  activeAlertsCount?: number;
}

export const HeroSection: React.FC<HeroSectionProps> = ({
  onSelectTab,
  totalSensors,
  totalReadings,
  activeAlertsCount = 0,
}) => {
  return (
    <section className="command-hero-card">
      <div className="command-hero-grid">
        {/* Left Column: Mission & Navigation Actions */}
        <div className="command-hero-main">
          <div className="command-tag">
            <span className="command-tag-dot" />
            CAMPUS TELEMETRY & INCIDENT COMMAND
          </div>
          <h1 className="command-hero-title">
            Environmental Health & Facility Operations
          </h1>
          <p className="command-hero-desc">
            Continuous real-time telemetry across 5 campus academic buildings with sub-second batch ingestion, deterministic deduplication, and automated threshold alerts.
          </p>

          <div className="command-action-row">
            <button
              onClick={() => onSelectTab?.('sensors')}
              className="command-action-btn primary"
            >
              <Cpu size={15} />
              <span>Inspect Sensors Fleet</span>
              <ArrowRight size={14} />
            </button>
            <button
              onClick={() => onSelectTab?.('trends')}
              className="command-action-btn secondary"
            >
              <Activity size={15} />
              <span>Multi-Metric Trends</span>
            </button>
            <button
              onClick={() => onSelectTab?.('simulator')}
              className="command-action-btn tertiary"
            >
              <Layers size={15} />
              <span>Simulator Engine</span>
            </button>
          </div>
        </div>

        {/* Right Column: Operational KPI Tiles */}
        <div className="command-kpi-grid">
          <div className="command-kpi-tile">
            <div className="kpi-header">
              <span className="kpi-label">MONITORED FLEET</span>
              <span className="kpi-status-tag green">ACTIVE</span>
            </div>
            <div className="kpi-val-row">
              <span className="kpi-value">{totalSensors}</span>
              <span className="kpi-unit">Nodes</span>
            </div>
            <div className="kpi-sub">Across 5 Campus Buildings</div>
          </div>

          <div className="command-kpi-tile">
            <div className="kpi-header">
              <span className="kpi-label">INGESTION SCALE</span>
              <span className="kpi-status-tag blue">VERIFIED</span>
            </div>
            <div className="kpi-val-row">
              <span className="kpi-value">10,000</span>
              <span className="kpi-unit">/sec</span>
            </div>
            <div className="kpi-sub">{totalReadings.toLocaleString()} Durable Readings</div>
          </div>

          <div className={`command-kpi-tile ${activeAlertsCount > 0 ? 'alert-warning' : 'alert-normal'}`}>
            <div className="kpi-header">
              <span className="kpi-label">INCIDENT POSTURE</span>
              <span className={`kpi-status-tag ${activeAlertsCount > 0 ? 'red' : 'green'}`}>
                {activeAlertsCount > 0 ? 'ACTION REQUIRED' : 'NOMINAL'}
              </span>
            </div>
            <div className="kpi-val-row">
              <span className="kpi-value">{activeAlertsCount}</span>
              <span className="kpi-unit">{activeAlertsCount === 1 ? 'Open Incident' : 'Open Incidents'}</span>
            </div>
            {activeAlertsCount > 0 ? (
              <button
                onClick={() => onSelectTab?.('alerts')}
                className="kpi-link-btn"
              >
                <span>Triage Incidents</span>
                <ArrowRight size={13} />
              </button>
            ) : (
              <div className="kpi-sub green-text">All 58 Nodes Within Safe Limits</div>
            )}
          </div>
        </div>
      </div>
    </section>
  );
};

export default HeroSection;
