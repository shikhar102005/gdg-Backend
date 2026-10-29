import React, { useState } from 'react';
import { Activity, Flame, Layers, Play, RefreshCw, Square, Wind, Zap } from 'lucide-react';
import { SimulatorStatus } from '../types';
import { api } from '../api/client';

interface SimulatorPageProps {
  simulator: SimulatorStatus | null;
  onRefresh: () => void;
}

export const SimulatorPage: React.FC<SimulatorPageProps> = ({ simulator, onRefresh }) => {
  const [rateInput, setRateInput] = useState(simulator?.rate_per_second || 10);
  const [actionLoading, setActionLoading] = useState(false);
  const [spikeMessage, setSpikeMessage] = useState('');

  const handleStart = async () => {
    setActionLoading(true);
    try {
      await api.startSimulator(rateInput);
      await onRefresh();
    } finally {
      setActionLoading(false);
    }
  };

  const handleStop = async () => {
    setActionLoading(true);
    try {
      await api.stopSimulator();
      await onRefresh();
    } finally {
      setActionLoading(false);
    }
  };

  const handleTriggerSpike = async (metric: string, sensorId: string, val: number, desc: string) => {
    setActionLoading(true);
    try {
      await api.triggerSpike(metric, sensorId, val);
      setSpikeMessage(`Dispatched deliberate threshold spike: ${desc} (${val})! Check Alerts tab.`);
      setTimeout(() => setSpikeMessage(''), 5000);
      await onRefresh();
    } finally {
      setActionLoading(false);
    }
  };

  return (
    <div>
      <div className="retro-big-card">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '0.75rem' }}>
          <div>
            <h2 style={{ fontWeight: 800, fontSize: '1.2rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#202124' }}>
              <Layers size={22} color="#1a73e8" />
              Campus Sensor Fleet Telemetry Simulator
            </h2>
            <p style={{ fontSize: '0.8rem', color: '#5f6368', fontWeight: 600, marginTop: '0.2rem' }}>
              Background sensor telemetry generator & intentional threshold spike injector.
            </p>
          </div>
          <button
            onClick={onRefresh}
            className="retro-pill-btn yellow"
            style={{ fontSize: '0.82rem', padding: '0.35rem 1rem' }}
          >
            <RefreshCw size={13} /> Refresh Status
          </button>
        </div>

        {spikeMessage && (
          <div style={{
            background: '#ceead6',
            border: '2px solid #202124',
            borderRadius: '12px',
            padding: '0.75rem 1rem',
            marginBottom: '1.25rem',
            fontWeight: 800,
            fontSize: '0.88rem',
            boxShadow: '3px 3px 0px #202124',
            color: '#0d652d',
          }}>
            {spikeMessage}
          </div>
        )}

        {/* Simulator Status Panel */}
        <div className="retro-inner-panel" style={{ marginBottom: '1.5rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.5rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              <span className={`retro-badge ${simulator?.running ? 'active' : 'warning'}`}>
                {simulator?.running ? 'RUNNING' : 'STOPPED'}
              </span>
              <span style={{ fontSize: '0.9rem', fontWeight: 700, color: '#202124' }}>
                {simulator?.running
                  ? `Active Rate: ${simulator.rate_per_second} readings/sec`
                  : 'Telemetry generation idle'}
              </span>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              <input
                type="number"
                min="1"
                max="500"
                className="retro-pill-input"
                style={{ width: '120px' }}
                value={rateInput}
                onChange={(e) => setRateInput(parseInt(e.target.value) || 5)}
              />
              {simulator?.running ? (
                <button
                  onClick={handleStop}
                  disabled={actionLoading}
                  className="retro-pill-btn coral"
                >
                  <Square size={14} /> Stop
                </button>
              ) : (
                <button
                  onClick={handleStart}
                  disabled={actionLoading}
                  className="retro-pill-btn green"
                >
                  <Play size={14} /> Start
                </button>
              )}
            </div>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '1rem', marginTop: '1rem' }}>
            <div style={{ background: '#f8fafd', border: '1.5px solid #202124', borderRadius: '10px', padding: '0.75rem' }}>
              <div style={{ fontSize: '0.72rem', fontWeight: 800, color: '#5f6368' }}>TOTAL EMITTED</div>
              <div style={{ fontSize: '1.25rem', fontWeight: 800, color: '#202124' }}>{(simulator?.total_emitted || 0).toLocaleString()}</div>
            </div>
            <div style={{ background: '#f8fafd', border: '1.5px solid #202124', borderRadius: '10px', padding: '0.75rem' }}>
              <div style={{ fontSize: '0.72rem', fontWeight: 800, color: '#5f6368' }}>SIMULATED SENSORS</div>
              <div style={{ fontSize: '1.25rem', fontWeight: 800, color: '#202124' }}>{simulator?.active_sensors || 8} Devices</div>
            </div>
          </div>
        </div>

        {/* Spike Injection Buttons Grid in Google Colors */}
        <div className="retro-inner-panel">
          <h3 style={{ fontWeight: 800, fontSize: '0.95rem', marginBottom: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#202124' }}>
            <Flame size={18} color="#ea4335" />
            Inject Deliberate Incident Spikes (Verify Alert Triggering & Deduplication)
          </h3>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '0.75rem' }}>
            <button
              onClick={() => handleTriggerSpike('temperature', 'mb-srv-temp', 39.5, 'Server Room Thermal Overheat (39.5°C)')}
              className="retro-pill-btn coral"
              style={{ justifyContent: 'flex-start', padding: '0.65rem 1rem' }}
            >
              <Flame size={16} /> Server Room Thermal Overheat (39.5°C)
            </button>
            <button
              onClick={() => handleTriggerSpike('co2', 'lib-fl1-co2', 1250, 'Library High CO2 Spike (1250 ppm)')}
              className="retro-pill-btn yellow"
              style={{ justifyContent: 'flex-start', padding: '0.65rem 1rem' }}
            >
              <Wind size={16} /> Library Poor Ventilation Spike (1250 ppm)
            </button>
            <button
              onClick={() => handleTriggerSpike('occupancy', 'sjt-aud-occ', 550, 'Auditorium Overcrowding (550 People)')}
              className="retro-pill-btn blue"
              style={{ justifyContent: 'flex-start', padding: '0.65rem 1rem' }}
            >
              <Activity size={16} /> Auditorium Max Capacity (550 People)
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
export default SimulatorPage;
