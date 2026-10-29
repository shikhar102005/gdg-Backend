import React, { useState } from 'react';
import { Activity, AlertTriangle, Check, CheckCircle2, Cpu, Flame, Layers, Lock, RotateCcw, Search, Zap } from 'lucide-react';
import { Sensor, TrendPoint } from '../types';
import { TrendsChart } from './TrendsChart';

interface SensorControlPanelProps {
  sensors: Sensor[];
  onTriggerSpike: (metric?: string, sensorId?: string, value?: number) => void;
  onRefresh: () => void;
  isOrganizer?: boolean;
  onRequireLogin?: () => void;
}

export const SensorControlPanel: React.FC<SensorControlPanelProps> = ({
  sensors,
  onTriggerSpike,
  onRefresh,
  isOrganizer = false,
  onRequireLogin,
}) => {
  const [selectedBuilding, setSelectedBuilding] = useState<string>('All');
  const [selectedType, setSelectedType] = useState<string>('All');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [activeSensorId, setActiveSensorId] = useState<string>(sensors[0]?.id || '');
  const [spikeSuccess, setSpikeSuccess] = useState<boolean>(false);

  const safeSensors = sensors || [];
  const dynamicBuildings = Array.from(new Set(safeSensors.map((s) => s.building))).filter(Boolean).sort();
  const buildings = ['All', ...dynamicBuildings];

  const dynamicTypes = Array.from(new Set(safeSensors.map((s) => s.sensor_type))).filter(Boolean).sort();
  const sensorTypes = ['All', ...dynamicTypes];

  const filteredSensors = safeSensors.filter((s) => {
    const matchesBuilding = selectedBuilding === 'All' || s.building.toLowerCase() === selectedBuilding.toLowerCase();
    const matchesType = selectedType === 'All' || s.sensor_type.toLowerCase() === selectedType.toLowerCase();
    const matchesSearch = !searchQuery || s.name.toLowerCase().includes(searchQuery.toLowerCase()) || s.id.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesBuilding && matchesType && matchesSearch;
  });

  const activeSensor = safeSensors.find((s) => s.id === activeSensorId) || filteredSensors[0] || null;

  const handleReset = () => {
    setSelectedBuilding('All');
    setSelectedType('All');
    setSearchQuery('');
    if (safeSensors.length > 0) setActiveSensorId(safeSensors[0].id);
  };

  const handleConfirmSpike = async () => {
    if (!activeSensor) return;
    let spikeVal = 39.5;
    switch (activeSensor.sensor_type) {
      case 'temperature': spikeVal = 39.5; break;
      case 'co2': spikeVal = 1250; break;
      case 'humidity': spikeVal = 88; break;
      case 'occupancy': spikeVal = 550; break;
      case 'noise': spikeVal = 92; break;
      case 'pm25': spikeVal = 70; break;
      default: spikeVal = 50.0;
    }
    onTriggerSpike(activeSensor.sensor_type, activeSensor.id, spikeVal);
    setSpikeSuccess(true);
    setTimeout(() => setSpikeSuccess(false), 3000);
  };

  return (
    <div className="retro-big-card">
      {/* Top Filter Pills Row 1: Buildings with Exact Counts Adding Up to Total */}
      <div className="retro-pill-bar" style={{ marginBottom: '0.65rem' }}>
        {buildings.map((b) => {
          const count = b === 'All'
            ? safeSensors.length
            : safeSensors.filter((s) => s.building.toLowerCase() === b.toLowerCase()).length;
          return (
            <button
              key={b}
              onClick={() => setSelectedBuilding(b)}
              className={`filter-pill ${selectedBuilding.toLowerCase() === b.toLowerCase() ? 'active' : ''}`}
            >
              {b === 'All' ? `ALL BUILDINGS (${count})` : `${b.toUpperCase()} (${count})`}
            </button>
          );
        })}
      </div>

      {/* Top Filter Pills Row 2: Sensor Types with Exact Counts Adding Up to Total */}
      <div className="retro-pill-bar" style={{ marginBottom: '1.25rem' }}>
        {sensorTypes.map((t) => {
          const count = t === 'All'
            ? safeSensors.length
            : safeSensors.filter((s) => s.sensor_type.toLowerCase() === t.toLowerCase()).length;
          const displayLabel = t === 'All'
            ? `ALL METRICS (${count})`
            : `${t.toUpperCase()} ('26 BATCH) (${count})`;
          return (
            <button
              key={t}
              onClick={() => setSelectedType(t)}
              className={`filter-pill ${selectedType.toLowerCase() === t.toLowerCase() ? 'active' : ''}`}
            >
              {displayLabel}
            </button>
          );
        })}
      </div>

      {/* Middle Search & Inputs Bar */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '0.75rem', marginBottom: '1.25rem' }}>
        <div>
          <input
            type="text"
            className="retro-pill-input"
            placeholder="Domain / Location..."
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
            <option value="All">All Sensor Metrics ({safeSensors.length})</option>
            {dynamicTypes.map((t) => {
              const count = safeSensors.filter((s) => s.sensor_type.toLowerCase() === t.toLowerCase()).length;
              return (
                <option key={t} value={t}>
                  {t.toUpperCase()} ({count})
                </option>
              );
            })}
          </select>
        </div>
        <div>
          <select
            className="retro-pill-input"
            value={selectedBuilding}
            onChange={(e) => setSelectedBuilding(e.target.value)}
          >
            <option value="All">All Campus Slots / Zones ({safeSensors.length})</option>
            {dynamicBuildings.map((b) => {
              const count = safeSensors.filter((s) => s.building.toLowerCase() === b.toLowerCase()).length;
              return (
                <option key={b} value={b}>
                  {b} ({count})
                </option>
              );
            })}
          </select>
        </div>
      </div>

      {/* Split Columns Grid */}
      <div className="retro-split-grid">
        {/* Left Column: Select Sensors */}
        <div className="retro-column-box">
          <div className="retro-column-title">
            Select Sensors ({filteredSensors.length} of {safeSensors.length} Available)
          </div>

          <div style={{ flex: 1, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '0.65rem' }}>
            {filteredSensors.length === 0 ? (
              <div style={{ textAlign: 'center', color: '#5f6368', padding: '2rem', fontSize: '0.85rem' }}>
                No campus sensors found matching these filters.
              </div>
            ) : (
              filteredSensors.map((s) => {
                const isSelected = activeSensor?.id === s.id;
                return (
                  <div
                    key={s.id}
                    onClick={() => setActiveSensorId(s.id)}
                    style={{
                      background: isSelected ? '#fde293' : '#ffffff',
                      border: '2px solid #202124',
                      borderRadius: '12px',
                      padding: '0.75rem 1rem',
                      cursor: 'pointer',
                      boxShadow: isSelected ? '3px 3px 0px #202124' : '2px 2px 0px #202124',
                      transform: isSelected ? 'translate(-1px, -1px)' : 'none',
                      transition: 'all 0.1s ease',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                    }}
                  >
                    <div>
                      <div style={{ fontWeight: 800, fontSize: '0.88rem', color: '#202124' }}>
                        {s.name}
                      </div>
                      <div style={{ fontSize: '0.75rem', color: '#5f6368', fontWeight: 600 }}>
                        {s.building} &bull; {s.location}
                      </div>
                    </div>
                    <div style={{ textAlign: 'right' }}>
                      <span className={`retro-badge ${s.status === 'active' ? 'active' : 'warning'}`}>
                        {s.sensor_type}
                      </span>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>

        {/* Right Column: Sensor Priority & Telemetry */}
        <div className="retro-column-box">
          <div className="retro-column-title">
            Sensor Telemetry & Priority Monitor
          </div>

          {activeSensor ? (
            <div style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
              <div>
                <div style={{
                  background: '#f8fafd',
                  border: '2px solid #202124',
                  borderRadius: '12px',
                  padding: '0.85rem',
                  boxShadow: '2px 2px 0px #202124',
                  marginBottom: '1rem',
                }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.35rem' }}>
                    <span style={{ fontWeight: 800, fontSize: '1rem', color: '#202124' }}>{activeSensor.name}</span>
                    <span className="retro-badge active">ONLINE</span>
                  </div>
                  <div style={{ fontSize: '0.78rem', color: '#3c4043', fontWeight: 600 }}>
                    <strong>ID:</strong> <code style={{ fontSize: '0.75rem' }}>{activeSensor.id}</code> &bull; <strong>Metric:</strong> {activeSensor.sensor_type} ({activeSensor.unit})
                  </div>
                  <div style={{ fontSize: '0.78rem', color: '#3c4043', fontWeight: 600 }}>
                    <strong>Campus:</strong> {activeSensor.campus} &bull; <strong>Zone:</strong> {activeSensor.building}
                  </div>
                </div>

                <div style={{
                  background: '#ffffff',
                  border: '2px solid #202124',
                  borderRadius: '12px',
                  padding: '1rem',
                  boxShadow: '2px 2px 0px #202124',
                  marginBottom: '1rem',
                }}>
                  <div style={{ fontSize: '0.85rem', fontWeight: 800, marginBottom: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <Activity size={16} color="#1a73e8" /> Active Nominal Baseline
                  </div>
                  <div style={{ fontSize: '0.82rem', color: '#5f6368', lineHeight: 1.6 }}>
                    This sensor observation is checked against single-reading thresholds and 15-second sliding windows. Ingestion telemetry is decoupled via an asynchronous buffered channel.
                  </div>
                </div>

                {spikeSuccess && (
                  <div style={{
                    background: '#ceead6',
                    border: '2px solid #202124',
                    borderRadius: '10px',
                    padding: '0.5rem 1rem',
                    fontWeight: 700,
                    fontSize: '0.82rem',
                    marginBottom: '1rem',
                    boxShadow: '2px 2px 0px #202124',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    color: '#0d652d',
                  }}>
                    <CheckCircle2 size={16} /> Test threshold violation spike injected! Checking rule alerts...
                  </div>
                )}
              </div>

              {/* Note matching red asterisk footnote */}
              <div style={{
                textAlign: 'center',
                color: isOrganizer ? '#d93025' : '#1a73e8',
                fontSize: '0.75rem',
                fontWeight: 700,
                marginTop: '1rem',
                paddingTop: '0.5rem',
                borderTop: '1px dashed #dadce0',
              }}>
                {isOrganizer
                  ? '*Click any sensor to monitor telemetry or inject a threshold spike test.'
                  : '*Viewing in read-only mode. Organizer log in required to inject test spikes.'}
              </div>
            </div>
          ) : (
            <div style={{ textAlign: 'center', color: '#5f6368', padding: '3rem 1rem' }}>
              Select a sensor from the left column.
            </div>
          )}
        </div>
      </div>

      {/* Bottom Action Buttons (Reset & Confirm buttons in Google Colors) */}
      <div style={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'center', gap: '1rem', marginTop: '1.25rem' }}>
        <button
          onClick={handleReset}
          className="retro-pill-btn coral"
          style={{ padding: '0.55rem 1.5rem', fontSize: '0.9rem' }}
        >
          Reset <RotateCcw size={14} />
        </button>
        <button
          onClick={() => isOrganizer ? handleConfirmSpike() : onRequireLogin?.()}
          className={`retro-pill-btn ${isOrganizer ? 'green' : 'disabled'}`}
          style={{ padding: '0.55rem 1.75rem', fontSize: '0.9rem' }}
          title={isOrganizer ? "Inject test threshold spike" : "Organizer login required"}
        >
          {!isOrganizer && <Lock size={14} />} Confirm / Test Spike <Check size={16} />
        </button>
      </div>
    </div>
  );
};
export default SensorControlPanel;
