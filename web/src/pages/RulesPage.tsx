import React, { useState } from 'react';
import { Check, Clock, Lock, Plus, Power, Sliders, X } from 'lucide-react';
import { AlertRule } from '../types';

interface RulesPageProps {
  rules: AlertRule[];
  onToggleRule: (id: string, enabled: boolean) => Promise<void>;
  onOpenCreateRule: () => void;
  onRefresh: () => void;
  isOrganizer?: boolean;
  onRequireLogin?: () => void;
}

export const RulesPage: React.FC<RulesPageProps> = ({
  rules,
  onToggleRule,
  onOpenCreateRule,
  onRefresh,
  isOrganizer = false,
  onRequireLogin,
}) => {
  const [togglingId, setTogglingId] = useState<string | null>(null);

  const safeRules = rules || [];

  const handleToggle = async (rule: AlertRule) => {
    if (!isOrganizer) {
      onRequireLogin?.();
      return;
    }
    setTogglingId(rule.id);
    try {
      await onToggleRule(rule.id, !rule.enabled);
    } finally {
      setTogglingId(null);
    }
  };

  return (
    <div>
      <div className="retro-big-card">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '0.75rem' }}>
          <div>
            <h2 style={{ fontWeight: 800, fontSize: '1.2rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#202124' }}>
              <Sliders size={22} color="#1a73e8" />
              Alert Rules & Incident Evaluation Criteria
            </h2>
            <p style={{ fontSize: '0.8rem', color: '#5f6368', fontWeight: 600, marginTop: '0.2rem' }}>
              Sub-millisecond single reading evaluation and 15-second rolling time-window aggregations.
            </p>
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button
              onClick={() => isOrganizer ? onOpenCreateRule() : onRequireLogin?.()}
              className={`retro-pill-btn ${isOrganizer ? 'green' : 'disabled'}`}
              style={{ fontSize: '0.82rem', padding: '0.35rem 1rem' }}
              title={isOrganizer ? "Create new alert threshold rule" : "Organizer login required"}
            >
              {!isOrganizer ? <Lock size={13} /> : <Plus size={14} />} Create Rule
            </button>
          </div>
        </div>

        {/* Read-Only Restriction Banner when unauthenticated */}
        {!isOrganizer && (
          <div style={{
            background: '#e8f0fe',
            border: '2px solid #202124',
            borderRadius: '12px',
            padding: '0.65rem 1rem',
            marginBottom: '1rem',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            flexWrap: 'wrap',
            gap: '0.5rem',
            boxShadow: '2px 2px 0px #202124',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.84rem', fontWeight: 700, color: '#1a73e8' }}>
              <Lock size={15} />
              <span>Read-Only Mode: You are viewing active threshold rules. Organizer login is required to create, configure, or toggle rules.</span>
            </div>
            <button
              onClick={onRequireLogin}
              className="retro-pill-btn green"
              style={{ padding: '0.25rem 0.85rem', fontSize: '0.75rem' }}
            >
              Organizer Log In
            </button>
          </div>
        )}

        {/* Rules Table */}
        <div className="retro-inner-panel">
          <div className="retro-table-container">
            <table className="retro-table">
              <thead>
                <tr>
                  <th>Rule Name</th>
                  <th>Evaluation Class</th>
                  <th>Target Scope</th>
                  <th>Condition</th>
                  <th>Severity</th>
                  <th>Status</th>
                  <th>Toggle</th>
                </tr>
              </thead>
              <tbody>
                {safeRules.length === 0 ? (
                  <tr>
                    <td colSpan={7} style={{ textAlign: 'center', padding: '2.5rem', color: '#5f6368' }}>
                      No alert rules configured yet. Click "Create Rule" above to add one.
                    </td>
                  </tr>
                ) : (
                  safeRules.map((r) => (
                    <tr key={r.id}>
                      <td>
                        <div style={{ fontWeight: 800 }}>{r.name}</div>
                        <div style={{ fontSize: '0.72rem', color: '#5f6368' }}>{r.description || 'Automated evaluation rule'}</div>
                      </td>
                      <td>
                        <span className="retro-badge info" style={{ fontSize: '0.72rem' }}>
                          {r.rule_class === 'single_reading' ? 'INSTANT' : `WINDOW (${r.window_aggregation?.toUpperCase()} / ${r.window_duration_seconds}s)`}
                        </span>
                      </td>
                      <td>
                        <span style={{ fontSize: '0.8rem', fontWeight: 700 }}>
                          {r.scope_type.toUpperCase()}{r.scope_value ? `: ${r.scope_value}` : ''}
                        </span>
                      </td>
                      <td>
                        <span className="font-mono" style={{ fontWeight: 800, color: '#202124' }}>
                          {r.metric} {r.operator} {r.threshold}
                        </span>
                      </td>
                      <td>
                        <span className={`retro-badge ${r.severity === 'critical' ? 'critical' : r.severity === 'warning' ? 'warning' : 'info'}`}>
                          {r.severity}
                        </span>
                      </td>
                      <td>
                        <span className={`retro-badge ${r.enabled ? 'active' : 'warning'}`}>
                          {r.enabled ? 'ACTIVE' : 'DISABLED'}
                        </span>
                      </td>
                      <td>
                        <button
                          onClick={() => handleToggle(r)}
                          disabled={togglingId === r.id}
                          className={`retro-pill-btn ${!isOrganizer ? 'disabled' : r.enabled ? 'coral' : 'green'}`}
                          style={{ padding: '0.2rem 0.65rem', fontSize: '0.72rem' }}
                          title={!isOrganizer ? "Organizer login required" : r.enabled ? "Disable rule" : "Enable rule"}
                        >
                          {!isOrganizer && <Lock size={10} />}
                          {r.enabled ? 'Disable' : 'Enable'}
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
    </div>
  );
};
export default RulesPage;
