import React, { useRef } from 'react';
import { Activity, AlertTriangle, CheckCircle, Clock, Cpu, Flame, Layers, ShieldAlert, Zap } from 'lucide-react';
import { Alert, DashboardSummary, Sensor, SimulatorStatus } from '../types';
import { MetricCard } from '../components/MetricCard';
import { HeroSection } from '../components/HeroSection';
import { SensorControlPanel } from '../components/SensorControlPanel';

interface OverviewPageProps {
  summary: DashboardSummary | null;
  sensors: Sensor[];
  simulator: SimulatorStatus | null;
  onOpenAlertAction: (alert: Alert, action: 'acknowledge' | 'resolve' | 'reopen') => void;
  onSelectTab: (tab: string) => void;
  onTriggerSpike: (metric?: string, sensorId?: string, value?: number) => void;
  onRefresh: () => void;
}

export const OverviewPage: React.FC<OverviewPageProps> = ({
  summary,
  sensors,
  simulator,
  onOpenAlertAction,
  onSelectTab,
  onTriggerSpike,
  onRefresh,
}) => {
  const panelRef = useRef<HTMLDivElement>(null);

  const handleScrollToPanel = () => {
    panelRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const safeSensors = sensors || [];
  const safeRecentAlerts = summary?.recent_alerts || [];

  return (
    <div>
      <HeroSection
        onStartClick={handleScrollToPanel}
        onSelectTab={onSelectTab}
        onTriggerSpike={onTriggerSpike}
        totalSensors={summary?.total_sensors || safeSensors.length || 58}
        totalReadings={summary?.total_readings || 26000}
        activeAlertsCount={summary?.active_alerts_count ?? (safeRecentAlerts.filter(a => a.status === 'OPEN' || a.status === 'ACKNOWLEDGED').length)}
      />

      {/* Simulator Active Notification Banner */}
      {simulator?.running && (
        <div style={{
          background: '#ceead6',
          border: '2.5px solid #202124',
          borderRadius: '16px',
          padding: '0.85rem 1.25rem',
          marginBottom: '1.75rem',
          boxShadow: '3px 3px 0px #202124',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexWrap: 'wrap',
          gap: '0.75rem',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', fontSize: '0.88rem', fontWeight: 700, color: '#0d652d' }}>
            <Zap size={18} color="#0d652d" />
            <span>Telemetry Simulator Active &bull; Ingesting at <strong>{simulator.rate_per_second} readings/sec</strong> ({simulator.total_emitted?.toLocaleString()} total emitted)</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
            <button
              onClick={() => onTriggerSpike()}
              className="retro-pill-btn coral"
              style={{ padding: '0.35rem 0.9rem', fontSize: '0.8rem' }}
              title="Inject 39.5°C threshold violation spike"
            >
              <Flame size={14} /> Inject Test Spike
            </button>
            <button
              onClick={() => onSelectTab('simulator')}
              className="retro-pill-btn yellow"
              style={{ padding: '0.35rem 0.9rem', fontSize: '0.8rem' }}
            >
              Simulator Settings
            </button>
          </div>
        </div>
      )}

      {/* Metrics Row (4 Original Metric Cards) */}
      <div className="retro-metric-grid">
        <MetricCard
          title="Active Alerts"
          value={summary?.active_alerts_count ?? 0}
          icon={AlertTriangle}
          subtext={`${summary?.critical_alerts ?? 0} critical, ${summary?.warning_alerts ?? 0} warnings`}
          badge={
            (summary?.active_alerts_count ?? 0) > 0
              ? { text: 'ATTENTION', variant: 'critical' }
              : { text: 'ALL CLEAR', variant: 'active' }
          }
        />
        <MetricCard
          title="Campus Sensors"
          value={summary?.total_sensors ?? safeSensors.length}
          icon={Cpu}
          subtext={`${summary?.active_sensors ?? safeSensors.filter((s) => s.status === 'active').length} online`}
          badge={{ text: 'FLEET ONLINE', variant: 'active' }}
        />
        <MetricCard
          title="Readings Ingested"
          value={summary?.total_readings?.toLocaleString() ?? 0}
          icon={Activity}
          subtext={`${summary?.readings_per_minute?.toFixed(0) ?? 0} / min stream`}
          badge={{ text: '10K BURST CAPABLE', variant: 'info' }}
        />
        <MetricCard
          title="Telemetry Engine"
          value={summary?.last_ingest_at ? 'STREAMING' : 'READY'}
          icon={Clock}
          subtext={
            summary?.last_ingest_at
              ? `Last: ${new Date(summary.last_ingest_at).toLocaleTimeString()}`
              : 'Awaiting events'
          }
          badge={{ text: 'OPERATIONAL', variant: 'active' }}
        />
      </div>

      {/* Target anchor for Hero 'Start ↓' button */}
      <div ref={panelRef}>
        {/* Interactive 2-Column Sensor Control Panel (Original Layout) */}
        <SensorControlPanel
          sensors={safeSensors}
          onTriggerSpike={onTriggerSpike}
          onRefresh={onRefresh}
        />
      </div>

      {/* Recent Incidents Card */}
      <div className="retro-inner-panel" style={{ marginTop: '1.5rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.5rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontWeight: 800, fontSize: '1.05rem', color: '#202124' }}>
            <ShieldAlert size={20} color="#ea4335" />
            <span>Recent Incidents & Incident Actions</span>
          </div>
          <button
            onClick={() => onSelectTab('alerts')}
            className="retro-pill-btn blue"
            style={{ fontSize: '0.8rem', padding: '0.35rem 1rem' }}
          >
            View All Incident Feeds &rarr;
          </button>
        </div>

        {safeRecentAlerts.length === 0 ? (
          <div style={{ padding: '2.5rem 1rem', textAlign: 'center', color: '#5f6368', fontSize: '0.9rem', fontWeight: 600 }}>
            <CheckCircle size={36} color="#34a853" style={{ margin: '0 auto 0.6rem auto', display: 'block' }} />
            All campus environmental sensors are operating nominally within established thresholds.
          </div>
        ) : (
          <div className="retro-table-container">
            <table className="retro-table">
              <thead>
                <tr>
                  <th>Severity</th>
                  <th>Incident / Sensor</th>
                  <th>Observed Value</th>
                  <th>Status</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                {safeRecentAlerts.map((al) => (
                  <tr key={al.id}>
                    <td>
                      <span className={`retro-badge ${al.severity === 'critical' ? 'critical' : 'warning'}`}>
                        {al.severity}
                      </span>
                    </td>
                    <td>
                      <div style={{ fontWeight: 800 }}>{al.rule_name || al.rule_id}</div>
                      <div style={{ fontSize: '0.75rem', color: '#5f6368' }}>
                        {al.sensor_name || al.sensor_id} ({al.building || 'Campus'})
                      </div>
                    </td>
                    <td>
                      <span className="font-mono" style={{ fontWeight: 800, color: '#ea4335' }}>
                        {al.trigger_value}
                      </span>
                    </td>
                    <td>
                      <span className={`retro-badge ${al.status === 'OPEN' ? 'warning' : 'active'}`}>
                        {al.status}
                      </span>
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: '0.4rem' }}>
                        {al.status === 'OPEN' && (
                          <button
                            onClick={() => onOpenAlertAction(al, 'acknowledge')}
                            className="retro-pill-btn yellow"
                            style={{ padding: '0.2rem 0.65rem', fontSize: '0.75rem' }}
                          >
                            Acknowledge
                          </button>
                        )}
                        {(al.status === 'OPEN' || al.status === 'ACKNOWLEDGED') && (
                          <button
                            onClick={() => onOpenAlertAction(al, 'resolve')}
                            className="retro-pill-btn green"
                            style={{ padding: '0.2rem 0.65rem', fontSize: '0.75rem' }}
                          >
                            Resolve
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};
export default OverviewPage;
