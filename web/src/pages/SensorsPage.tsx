import React, { useState } from 'react';
import { Cpu, History, RefreshCw, Search, X } from 'lucide-react';
import { Reading, Sensor } from '../types';
import { api } from '../api/client';

interface SensorsPageProps {
  sensors: Sensor[];
  onRefresh: () => void;
}

export const SensorsPage: React.FC<SensorsPageProps> = ({ sensors, onRefresh }) => {
  const [selectedBuilding, setSelectedBuilding] = useState('');
  const [selectedType, setSelectedType] = useState('');
  const [selectedStatus, setSelectedStatus] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  // Sensor history inspection
  const [inspectSensor, setInspectSensor] = useState<Sensor | null>(null);
  const [historyReadings, setHistoryReadings] = useState<Reading[]>([]);
  const [loadingHistory, setLoadingHistory] = useState(false);

  const safeSensors = sensors || [];
  const buildings = Array.from(new Set(safeSensors.map((s) => s.building))).filter(Boolean);
  const types = Array.from(new Set(safeSensors.map((s) => s.sensor_type))).filter(Boolean);

  const filteredSensors = safeSensors.filter((s) => {
    if (selectedBuilding && s.building !== selectedBuilding) return false;
    if (selectedType && s.sensor_type !== selectedType) return false;
    if (selectedStatus && s.status !== selectedStatus) return false;
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      const matchID = s.id.toLowerCase().includes(q);
      const matchName = s.name.toLowerCase().includes(q);
      const matchLoc = s.location.toLowerCase().includes(q);
      if (!matchID && !matchName && !matchLoc) return false;
    }
    return true;
  });

  const handleInspect = async (sensor: Sensor) => {
    setInspectSensor(sensor);
    setLoadingHistory(true);
    try {
      const res = await api.getSensorHistory(sensor.id, 50);
      setHistoryReadings(res?.readings || []);
    } catch {
      setHistoryReadings([]);
    } finally {
      setLoadingHistory(false);
    }
  };

  return (
    <div>
      <div className="retro-big-card">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '0.75rem' }}>
          <h2 style={{ fontWeight: 800, fontSize: '1.2rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#202124' }}>
            <Cpu size={22} color="#1a73e8" />
            Campus Sensor Fleet Directory ({filteredSensors.length} sensors)
          </h2>
          <button
            onClick={onRefresh}
            className="retro-pill-btn yellow"
            style={{ fontSize: '0.82rem', padding: '0.35rem 1rem' }}
          >
            <RefreshCw size={13} /> Refresh Fleet
          </button>
        </div>

        {/* Filter Toolbar with Pills */}
        <div className="retro-pill-bar" style={{ marginBottom: '1rem' }}>
          <button
            onClick={() => setSelectedBuilding('')}
            className={`filter-pill ${selectedBuilding === '' ? 'active' : ''}`}
          >
            ALL BUILDINGS
          </button>
          {buildings.map((b) => (
            <button
              key={b}
              onClick={() => setSelectedBuilding(b)}
              className={`filter-pill ${selectedBuilding === b ? 'active' : ''}`}
            >
              {b.toUpperCase()}
            </button>
          ))}
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '0.75rem', marginBottom: '1.25rem' }}>
          <div>
            <input
              type="text"
              className="retro-pill-input"
              placeholder="Search sensor ID, name, location..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
          <div>
            <select
              className="retro-pill-input"
              value={selectedType}
              onChange={(e) => setSelectedType(e.target.value)}
            >
              <option value="">All Sensor Types</option>
              {types.map((t) => (
                <option key={t} value={t}>{t}</option>
              ))}
            </select>
          </div>
          <div>
            <select
              className="retro-pill-input"
              value={selectedStatus}
              onChange={(e) => setSelectedStatus(e.target.value)}
            >
              <option value="">All Operational Statuses</option>
              <option value="active">Active & Online</option>
              <option value="offline">Offline / Standby</option>
            </select>
          </div>
        </div>

        {/* Sensors Table */}
        <div className="retro-inner-panel">
          <div className="retro-table-container">
            <table className="retro-table">
              <thead>
                <tr>
                  <th>Sensor ID</th>
                  <th>Device Name</th>
                  <th>Campus & Zone</th>
                  <th>Metric Type</th>
                  <th>Status</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                {filteredSensors.length === 0 ? (
                  <tr>
                    <td colSpan={6} style={{ textAlign: 'center', padding: '2rem', color: '#5f6368' }}>
                      No sensors match the criteria.
                    </td>
                  </tr>
                ) : (
                  filteredSensors.map((s) => (
                    <tr key={s.id}>
                      <td>
                        <code style={{ fontWeight: 800 }}>{s.id}</code>
                      </td>
                      <td>
                        <div style={{ fontWeight: 800 }}>{s.name}</div>
                        <div style={{ fontSize: '0.75rem', color: '#5f6368' }}>{s.location}</div>
                      </td>
                      <td>
                        <div style={{ fontWeight: 700 }}>{s.building}</div>
                        <div style={{ fontSize: '0.75rem', color: '#5f6368' }}>{s.campus}</div>
                      </td>
                      <td>
                        <span className="retro-badge info">
                          {s.sensor_type} ({s.unit})
                        </span>
                      </td>
                      <td>
                        <span className={`retro-badge ${s.status === 'active' ? 'active' : 'warning'}`}>
                          {s.status}
                        </span>
                      </td>
                      <td>
                        <button
                          onClick={() => handleInspect(s)}
                          className="retro-pill-btn blue"
                          style={{ padding: '0.2rem 0.65rem', fontSize: '0.75rem' }}
                        >
                          <History size={12} /> Inspect History
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* History Inspection Modal */}
      {inspectSensor && (
        <div className="retro-modal-overlay" onClick={() => setInspectSensor(null)}>
          <div className="retro-modal-window" style={{ maxWidth: '650px' }} onClick={(e) => e.stopPropagation()}>
            <div className="retro-modal-header">
              <div className="retro-modal-dots">
                <span className="window-dot red" onClick={() => setInspectSensor(null)} style={{ cursor: 'pointer' }}></span>
                <span className="window-dot yellow"></span>
                <span className="window-dot green"></span>
              </div>
              <div className="retro-modal-title">Telemetry Observations &bull; {inspectSensor.id}</div>
            </div>

            <div className="retro-modal-body" style={{ maxHeight: '420px', overflowY: 'auto' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '1rem', background: '#fff', border: '2px solid #202124', borderRadius: '10px', padding: '0.75rem', boxShadow: '2px 2px 0px #202124' }}>
                <div>
                  <strong>{inspectSensor.name}</strong>
                  <div style={{ fontSize: '0.75rem', color: '#5f6368' }}>{inspectSensor.building} &bull; {inspectSensor.location}</div>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <span className="retro-badge info">{inspectSensor.sensor_type} ({inspectSensor.unit})</span>
                </div>
              </div>

              {loadingHistory ? (
                <div style={{ textAlign: 'center', padding: '2rem' }}>Loading historical events...</div>
              ) : historyReadings.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '2rem', color: '#5f6368' }}>No historical readings recorded yet.</div>
              ) : (
                <div className="retro-table-container">
                  <table className="retro-table">
                    <thead>
                      <tr>
                        <th>Observed At</th>
                        <th>Value</th>
                        <th>Quality</th>
                      </tr>
                    </thead>
                    <tbody>
                      {historyReadings.map((r) => (
                        <tr key={r.id}>
                          <td style={{ fontSize: '0.78rem' }}>{new Date(r.observed_at).toLocaleTimeString()}</td>
                          <td>
                            <strong className="font-mono">{r.value.toFixed(2)}</strong> {inspectSensor.unit}
                          </td>
                          <td>
                            <span className="retro-badge active" style={{ fontSize: '0.68rem', padding: '0.1rem 0.35rem' }}>
                              {r.quality}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>

            <div className="retro-modal-footer">
              <button onClick={() => setInspectSensor(null)} className="retro-modal-ok-btn">
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
export default SensorsPage;
