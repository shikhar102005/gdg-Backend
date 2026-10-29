import React, { useState } from 'react';
import { Alert } from '../types';

interface AlertActionModalProps {
  alert: Alert | null;
  action: 'acknowledge' | 'resolve' | 'reopen' | null;
  onClose: () => void;
  onSubmit: (alertId: string, action: string, notes?: string) => Promise<void>;
}

export const AlertActionModal: React.FC<AlertActionModalProps> = ({
  alert,
  action,
  onClose,
  onSubmit,
}) => {
  const [notes, setNotes] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (!alert || !action) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await onSubmit(alert.id, action, notes);
      onClose();
    } catch (err: any) {
      setError(err.message || 'Action failed');
    } finally {
      setLoading(false);
    }
  };

  const actionDescriptions = {
    acknowledge: {
      mode: 'Incident Status – ACKNOWLEDGED',
      desc: 'Acknowledging records that the response team is actively addressing this condition. Deduplication remains active.',
    },
    resolve: {
      mode: 'Incident Status – RESOLVED',
      desc: 'Resolving records root cause analysis and clears the incident. A recurrence will trigger a new lifecycle incident.',
    },
    reopen: {
      mode: 'Incident Status – REOPEN',
      desc: 'Reopens this alert incident for investigation if telemetry violation persists.',
    },
  };

  const currentDesc = actionDescriptions[action];

  return (
    <div className="retro-modal-overlay" onClick={onClose}>
      <div className="retro-modal-window" onClick={(e) => e.stopPropagation()}>
        {/* Retro Window Header with macOS-style control dots with Google Colors */}
        <div className="retro-modal-header">
          <div className="retro-modal-dots">
            <span className="window-dot red" onClick={onClose} style={{ cursor: 'pointer' }}></span>
            <span className="window-dot yellow"></span>
            <span className="window-dot green"></span>
          </div>
          <div className="retro-modal-title">Alert</div>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="retro-modal-body">
            {error && (
              <div style={{
                background: '#fad2cf',
                border: '2px solid #202124',
                padding: '0.5rem',
                borderRadius: '8px',
                fontSize: '0.8rem',
                marginBottom: '1rem',
                color: '#c5221f',
                fontWeight: 700,
              }}>
                {error}
              </div>
            )}

            <div style={{ textAlign: 'center', marginBottom: '1.25rem' }}>
              <div style={{ fontWeight: 800, fontSize: '1rem', marginBottom: '0.35rem', color: '#202124' }}>
                {currentDesc.mode}
              </div>
              <div style={{ fontSize: '0.82rem', color: '#5f6368', lineHeight: 1.5 }}>
                {currentDesc.desc}
              </div>
            </div>

            <div style={{
              background: '#f8fafd',
              border: '2px solid #202124',
              borderRadius: '10px',
              padding: '0.75rem',
              marginBottom: '1.25rem',
              fontSize: '0.82rem',
            }}>
              <div><strong>Alert Rule:</strong> {alert.rule_name || alert.rule_id}</div>
              <div><strong>Sensor:</strong> {alert.sensor_name || alert.sensor_id} ({alert.building || alert.campus || 'Campus'})</div>
              <div><strong>Trigger Value:</strong> <span className="font-mono" style={{ fontWeight: 800, color: '#ea4335' }}>{alert.trigger_value}</span></div>
            </div>

            {action === 'resolve' && (
              <div>
                <label style={{ display: 'block', fontWeight: 800, fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                  Root Cause Analysis / Resolution Notes *
                </label>
                <textarea
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  placeholder="e.g. Ventilation dampers reopened; temperature normalized."
                  className="retro-pill-input"
                  style={{ borderRadius: '12px', minHeight: '80px', resize: 'vertical' }}
                  required
                />
              </div>
            )}
          </div>

          <div className="retro-modal-footer">
            <button
              type="button"
              onClick={onClose}
              className="retro-pill-btn"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="retro-modal-ok-btn"
            >
              {loading ? 'Processing...' : 'OK'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
export default AlertActionModal;
