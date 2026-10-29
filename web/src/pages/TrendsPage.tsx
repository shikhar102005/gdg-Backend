import React, { useEffect, useState } from 'react';
import { Activity, BarChart2, Filter, RefreshCw } from 'lucide-react';
import { Sensor, TrendPoint } from '../types';
import { TrendsChart } from '../components/TrendsChart';
import { api } from '../api/client';

interface TrendsPageProps {
  sensors: Sensor[];
}

export const TrendsPage: React.FC<TrendsPageProps> = ({ sensors }) => {
  const [metric, setMetric] = useState('temperature');
  const [duration, setDuration] = useState('1h');
  const [selectedBuilding, setSelectedBuilding] = useState('');
  const [selectedSensor, setSelectedSensor] = useState('');
  const [points, setPoints] = useState<TrendPoint[]>([]);
  const [loading, setLoading] = useState(false);

  const safeSensors = sensors || [];
  const buildings = Array.from(new Set(safeSensors.map((s) => s.building))).filter(Boolean);

  const units: Record<string, string> = {
    temperature: '°C',
    humidity: '%',
    co2: 'ppm',
    occupancy: 'people',
    noise: 'dB',
    pm25: 'µg/m³',
  };

  const fetchTrends = async () => {
    setLoading(true);
    try {
      const res = await api.getMetricTrends(metric, selectedBuilding, selectedSensor, duration);
      setPoints(res?.points || []);
    } catch {
      setPoints([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTrends();
  }, [metric, duration, selectedBuilding, selectedSensor]);

  const safePoints = points || [];
  const values = safePoints.map((p) => p.avg).filter((v) => typeof v === 'number' && !isNaN(v));
  const avg = values.length ? values.reduce((a, b) => a + b, 0) / values.length : 0;
  const min = values.length ? Math.min(...values) : 0;
  const max = values.length ? Math.max(...values) : 0;
  const totalReadings = safePoints.reduce((acc, p) => acc + (p.count || 0), 0);

  return (
    <div>
      <div className="retro-big-card">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '0.75rem' }}>
          <h2 style={{ fontWeight: 800, fontSize: '1.2rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#202124' }}>
            <Activity size={22} color="#1a73e8" />
            Environmental Telemetry & Time-Window Trends
          </h2>
          <button
            onClick={fetchTrends}
            disabled={loading}
            className="retro-pill-btn yellow"
            style={{ fontSize: '0.82rem', padding: '0.35rem 1rem' }}
          >
            <RefreshCw size={13} className={loading ? 'animate-spin' : ''} />
            {loading ? 'Fetching...' : 'Refresh Trends'}
          </button>
        </div>

        {/* Filter Pills Bar in Google Yellow */}
        <div className="retro-pill-bar" style={{ marginBottom: '1rem' }}>
          {['temperature', 'humidity', 'co2', 'occupancy', 'noise', 'pm25'].map((m) => (
            <button
              key={m}
              onClick={() => setMetric(m)}
              className={`filter-pill ${metric === m ? 'active' : ''}`}
            >
              {m.toUpperCase()} ({units[m] || ''})
            </button>
          ))}
        </div>

        {/* Window & Zone Inputs */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '0.75rem', marginBottom: '1.5rem' }}>
          <div>
            <select
              className="retro-pill-input"
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
            >
              <option value="15m">Window: 15 Minutes</option>
              <option value="1h">Window: 1 Hour</option>
              <option value="6h">Window: 6 Hours</option>
              <option value="24h">Window: 24 Hours</option>
            </select>
          </div>
          <div>
            <select
              className="retro-pill-input"
              value={selectedBuilding}
              onChange={(e) => setSelectedBuilding(e.target.value)}
            >
              <option value="">All Campus Zones</option>
              {buildings.map((b) => (
                <option key={b} value={b}>{b}</option>
              ))}
            </select>
          </div>
          <div>
            <select
              className="retro-pill-input"
              value={selectedSensor}
              onChange={(e) => setSelectedSensor(e.target.value)}
            >
              <option value="">All Compatible Sensors</option>
              {safeSensors.filter((s) => s.sensor_type === metric).map((s) => (
                <option key={s.id} value={s.id}>{s.name} ({s.id})</option>
              ))}
            </select>
          </div>
        </div>

        {/* Inner Chart Card */}
        <div className="retro-inner-panel" style={{ marginBottom: '1.5rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1rem' }}>
            <span style={{ fontWeight: 800, fontSize: '0.95rem', color: '#202124' }}>
              Trend Vector &bull; {metric.toUpperCase()} ({units[metric]})
            </span>
            <span className="retro-badge info">
              {safePoints.length} BUCKETS RECORDED
            </span>
          </div>

          <TrendsChart
            points={safePoints}
            metric={metric}
            unit={units[metric] || ''}
          />
        </div>

        {/* Aggregates Grid */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))', gap: '1rem' }}>
          <div style={{ background: '#ffffff', border: '2px solid #202124', borderRadius: '12px', padding: '0.85rem', boxShadow: '2px 2px 0px #202124' }}>
            <div style={{ fontSize: '0.72rem', fontWeight: 800, color: '#5f6368' }}>WINDOW AVERAGE</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 800, color: '#1a73e8' }}>{avg.toFixed(2)} {units[metric]}</div>
          </div>
          <div style={{ background: '#ffffff', border: '2px solid #202124', borderRadius: '12px', padding: '0.85rem', boxShadow: '2px 2px 0px #202124' }}>
            <div style={{ fontSize: '0.72rem', fontWeight: 800, color: '#5f6368' }}>MINIMUM OBSERVED</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 800, color: '#34a853' }}>{min.toFixed(2)} {units[metric]}</div>
          </div>
          <div style={{ background: '#ffffff', border: '2px solid #202124', borderRadius: '12px', padding: '0.85rem', boxShadow: '2px 2px 0px #202124' }}>
            <div style={{ fontSize: '0.72rem', fontWeight: 800, color: '#5f6368' }}>MAXIMUM OBSERVED</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 800, color: '#ea4335' }}>{max.toFixed(2)} {units[metric]}</div>
          </div>
          <div style={{ background: '#ffffff', border: '2px solid #202124', borderRadius: '12px', padding: '0.85rem', boxShadow: '2px 2px 0px #202124' }}>
            <div style={{ fontSize: '0.72rem', fontWeight: 800, color: '#5f6368' }}>AGGREGATE READINGS</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 800, color: '#202124' }}>{totalReadings.toLocaleString()}</div>
          </div>
        </div>
      </div>
    </div>
  );
};
export default TrendsPage;
