import React, { useState } from 'react';
import { PlusCircle, Sliders, X } from 'lucide-react';
import { AlertRule, RuleClass, ScopeType, Severity, WindowAggregation } from '../types';

interface RuleFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (rule: Partial<AlertRule>) => Promise<void>;
}

export const RuleFormModal: React.FC<RuleFormModalProps> = ({ isOpen, onClose, onSubmit }) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [metric, setMetric] = useState('temperature');
  const [ruleClass, setRuleClass] = useState<RuleClass>('single_reading');
  const [operator, setOperator] = useState('>');
  const [threshold, setThreshold] = useState<number>(30.0);
  const [severity, setSeverity] = useState<Severity>('warning');
  const [scopeType, setScopeType] = useState<ScopeType>('global');
  const [scopeValue, setScopeValue] = useState('');
  const [windowDuration, setWindowDuration] = useState<number>(600); // 10m
  const [windowAgg, setWindowAgg] = useState<WindowAggregation>('avg');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      await onSubmit({
        name,
        description,
        metric,
        rule_class: ruleClass,
        operator,
        threshold: Number(threshold),
        severity,
        scope_type: scopeType,
        scope_value: scopeValue,
        window_duration_seconds: Number(windowDuration),
        window_aggregation: windowAgg,
        enabled: true,
      });
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to create rule');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="retro-modal-overlay" onClick={onClose}>
      <div className="retro-modal-window" style={{ maxWidth: '560px' }} onClick={(e) => e.stopPropagation()}>
        {/* Retro Header with 3 Google-colored dots */}
        <div className="retro-modal-header">
          <div className="retro-modal-dots">
            <span className="window-dot red" onClick={onClose} style={{ cursor: 'pointer' }}></span>
            <span className="window-dot yellow"></span>
            <span className="window-dot green"></span>
          </div>
          <div className="retro-modal-title">Create Incident Detection Rule</div>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="retro-modal-body" style={{ maxHeight: '70vh', overflowY: 'auto' }}>
            {error && (
              <div style={{ background: '#fad2cf', border: '2px solid #202124', borderRadius: '8px', padding: '0.5rem', fontWeight: 700, fontSize: '0.8rem', marginBottom: '1rem', color: '#c5221f' }}>
                {error}
              </div>
            )}

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 800, marginBottom: '0.25rem' }}>
                Rule Name:
              </label>
              <input
                type="text"
                className="retro-pill-input"
                placeholder="e.g. Server Room Thermal Threshold Exceeded"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem', marginBottom: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 800, marginBottom: '0.25rem' }}>
                  Target Metric:
                </label>
                <select
                  className="retro-pill-input"
                  value={metric}
                  onChange={(e) => setMetric(e.target.value)}
                >
                  <option value="temperature">Temperature (°C)</option>
                  <option value="co2">CO2 (ppm)</option>
                  <option value="humidity">Humidity (%)</option>
                  <option value="occupancy">Occupancy (count)</option>
                  <option value="noise">Noise (dB)</option>
                  <option value="pm25">PM2.5 (µg/m³)</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 800, marginBottom: '0.25rem' }}>
                  Rule Class:
                </label>
                <select
                  className="retro-pill-input"
                  value={ruleClass}
                  onChange={(e) => setRuleClass(e.target.value as RuleClass)}
                >
                  <option value="single_reading">Single Reading (Instant)</option>
                  <option value="time_window">Time Window (Aggregate)</option>
                </select>
              </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem', marginBottom: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 800, marginBottom: '0.25rem' }}>
                  Comparison Operator:
                </label>
                <select
                  className="retro-pill-input"
                  value={operator}
                  onChange={(e) => setOperator(e.target.value)}
                >
                  <option value=">">Greater than (&gt;)</option>
                  <option value=">=">Greater than or equal (&gt;=)</option>
                  <option value="<">Less than (&lt;)</option>
                  <option value="<=">Less than or equal (&lt;=)</option>
                  <option value="==">Equal (==)</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 800, marginBottom: '0.25rem' }}>
                  Threshold Value:
                </label>
                <input
                  type="number"
                  step="any"
                  className="retro-pill-input"
                  value={threshold}
                  onChange={(e) => setThreshold(parseFloat(e.target.value) || 0)}
                  required
                />
              </div>
            </div>

            {ruleClass === 'time_window' && (
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem', marginBottom: '1rem', background: '#f8fafd', border: '1.5px solid #202124', borderRadius: '10px', padding: '0.75rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 800, marginBottom: '0.25rem' }}>
                    Window Aggregation:
                  </label>
                  <select
                    className="retro-pill-input"
                    value={windowAgg}
                    onChange={(e) => setWindowAgg(e.target.value as WindowAggregation)}
                  >
                    <option value="avg">Rolling Average (AVG)</option>
                    <option value="min">Rolling Minimum (MIN)</option>
                    <option value="max">Rolling Maximum (MAX)</option>
                    <option value="count">Sample Count (COUNT)</option>
                  </select>
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 800, marginBottom: '0.25rem' }}>
                    Lookback Seconds:
                  </label>
                  <input
                    type="number"
                    className="retro-pill-input"
                    value={windowDuration}
                    onChange={(e) => setWindowDuration(parseInt(e.target.value) || 60)}
                    placeholder="e.g. 600 for 10 min"
                  />
                </div>
              </div>
            )}

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 800, marginBottom: '0.25rem' }}>
                Severity Level:
              </label>
              <select
                className="retro-pill-input"
                value={severity}
                onChange={(e) => setSeverity(e.target.value as Severity)}
              >
                <option value="critical">CRITICAL (Red alert)</option>
                <option value="warning">WARNING (Yellow alert)</option>
                <option value="info">INFO (Blue notification)</option>
              </select>
            </div>
          </div>

          <div className="retro-modal-footer">
            <button
              type="submit"
              disabled={loading}
              className="retro-modal-ok-btn"
            >
              {loading ? 'Creating...' : 'Save Rule'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
export default RuleFormModal;
