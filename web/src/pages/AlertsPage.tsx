import React, { useState } from 'react';
import { AlertTriangle, CheckCircle, Clock, FileText, Filter, Lock, RefreshCw, RotateCcw, ShieldAlert } from 'lucide-react';
import { Alert, Sensor } from '../types';
import { IncidentReportModal } from '../components/IncidentReportModal';

interface AlertsPageProps {
  alerts: Alert[];
  sensors: Sensor[];
  onRefresh: () => void;
  onOpenAlertAction: (alert: Alert, action: 'acknowledge' | 'resolve' | 'reopen') => void;
  isOrganizer?: boolean;
  onRequireLogin?: () => void;
}

export const AlertsPage: React.FC<AlertsPageProps> = ({
  alerts,
  sensors,
  onRefresh,
  onOpenAlertAction,
  isOrganizer = false,
  onRequireLogin,
}) => {
  const [selectedStatus, setSelectedStatus] = useState<string>('ALL');
  const [selectedSeverity, setSelectedSeverity] = useState<string>('ALL');
  const [selectedBuilding, setSelectedBuilding] = useState<string>('');
  const [showReportModal, setShowReportModal] = useState<boolean>(false);

  const safeAlerts = alerts || [];
  const safeSensors = sensors || [];
  const buildings = Array.from(new Set(safeSensors.map((s) => s.building))).filter(Boolean);

  const filteredAlerts = safeAlerts.filter((al) => {
    if (selectedStatus !== 'ALL' && al.status !== selectedStatus) return false;
    if (selectedSeverity !== 'ALL' && al.severity.toUpperCase() !== selectedSeverity) return false;
    if (selectedBuilding && al.building !== selectedBuilding) return false;
    return true;
  });

  const openCount = safeAlerts.filter((a) => a.status === 'OPEN').length;
  const ackCount = safeAlerts.filter((a) => a.status === 'ACKNOWLEDGED').length;
  const resolvedCount = safeAlerts.filter((a) => a.status === 'RESOLVED').length;

  return (
    <div>
      <div className="retro-big-card">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '0.75rem' }}>
          <h2 style={{ fontWeight: 800, fontSize: '1.2rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#202124' }}>
            <ShieldAlert size={22} color="#ea4335" />
            Campus Incident Command & Incident Life-Cycle
          </h2>
          <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
            <button
              onClick={() => setShowReportModal(true)}
              className="retro-pill-btn blue"
              style={{ fontSize: '0.82rem', padding: '0.35rem 1rem' }}
              title="Generate, export, or print compliance incident report"
            >
              <FileText size={13} /> Generate Incident Report
            </button>
            <button
              onClick={onRefresh}
              className="retro-pill-btn yellow"
              style={{ fontSize: '0.82rem', padding: '0.35rem 1rem' }}
            >
              <RefreshCw size={13} /> Refresh Incidents
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
              <span>Read-Only Mode: You are viewing real-time incidents. Organizer login is required to acknowledge, resolve, or reopen incident tickets.</span>
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

        {/* Status Filter Pills in Google Colors */}
        <div className="retro-pill-bar" style={{ marginBottom: '1rem' }}>
          {[
            { id: 'ALL', label: `ALL INCIDENTS (${safeAlerts.length})` },
            { id: 'OPEN', label: `OPEN (${openCount})` },
            { id: 'ACKNOWLEDGED', label: `ACKNOWLEDGED (${ackCount})` },
            { id: 'RESOLVED', label: `RESOLVED (${resolvedCount})` },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setSelectedStatus(tab.id)}
              className={`filter-pill ${selectedStatus === tab.id ? 'active' : ''}`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* Severity & Building Filter Selects */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '0.75rem', marginBottom: '1.25rem' }}>
          <div>
            <select
              className="retro-pill-input"
              value={selectedSeverity}
              onChange={(e) => setSelectedSeverity(e.target.value)}
            >
              <option value="ALL">All Severities</option>
              <option value="CRITICAL">Critical Only</option>
              <option value="WARNING">Warning Only</option>
              <option value="INFO">Info Only</option>
            </select>
          </div>
          <div>
            <select
              className="retro-pill-input"
              value={selectedBuilding}
              onChange={(e) => setSelectedBuilding(e.target.value)}
            >
              <option value="">All Campus Buildings</option>
              {buildings.map((b) => (
                <option key={b} value={b}>{b}</option>
              ))}
            </select>
          </div>
        </div>

        {/* Incidents Table */}
        <div className="retro-inner-panel">
          {filteredAlerts.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '3rem 1rem', color: '#5f6368', fontSize: '0.9rem', fontWeight: 600 }}>
              <CheckCircle size={36} color="#34a853" style={{ margin: '0 auto 0.6rem auto', display: 'block' }} />
              No incidents match the active filters.
            </div>
          ) : (
            <div className="retro-table-container">
              <table className="retro-table">
                <thead>
                  <tr>
                    <th>Severity</th>
                    <th>Status</th>
                    <th>Alert Rule</th>
                    <th>Sensor & Building</th>
                    <th>Trigger Value</th>
                    <th>Triggered At</th>
                    <th>Lifecycle Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredAlerts.map((al) => (
                    <tr key={al.id}>
                      <td>
                        <span className={`retro-badge ${al.severity === 'critical' ? 'critical' : 'warning'}`}>
                          {al.severity}
                        </span>
                      </td>
                      <td>
                        <span className={`retro-badge ${al.status === 'OPEN' ? 'warning' : al.status === 'ACKNOWLEDGED' ? 'info' : 'active'}`}>
                          {al.status}
                        </span>
                      </td>
                      <td>
                        <div style={{ fontWeight: 800 }}>{al.rule_name || al.rule_id}</div>
                        {al.resolution_notes && (
                          <div style={{ fontSize: '0.75rem', color: '#34a853', marginTop: '0.2rem' }}>
                            <strong>Root cause:</strong> {al.resolution_notes}
                          </div>
                        )}
                      </td>
                      <td>
                        <div style={{ fontWeight: 700 }}>{al.sensor_name || al.sensor_id}</div>
                        <div style={{ fontSize: '0.75rem', color: '#5f6368' }}>{al.building || al.campus || 'Campus Zone'}</div>
                      </td>
                      <td>
                        <span className="font-mono" style={{ fontWeight: 800, color: '#ea4335' }}>
                          {al.trigger_value}
                        </span>
                      </td>
                      <td style={{ fontSize: '0.78rem', color: '#5f6368', whiteSpace: 'nowrap' }}>
                        {new Date(al.created_at).toLocaleTimeString()}
                      </td>
                      <td>
                        <div style={{ display: 'flex', gap: '0.35rem', flexWrap: 'wrap' }}>
                          {al.status === 'OPEN' && (
                            <button
                              onClick={() => isOrganizer ? onOpenAlertAction(al, 'acknowledge') : onRequireLogin?.()}
                              className={`retro-pill-btn ${isOrganizer ? 'yellow' : 'disabled'}`}
                              style={{ padding: '0.25rem 0.65rem', fontSize: '0.72rem' }}
                              title={isOrganizer ? "Acknowledge incident" : "Organizer login required"}
                            >
                              {!isOrganizer && <Lock size={10} />} Acknowledge
                            </button>
                          )}
                          {(al.status === 'OPEN' || al.status === 'ACKNOWLEDGED') && (
                            <button
                              onClick={() => isOrganizer ? onOpenAlertAction(al, 'resolve') : onRequireLogin?.()}
                              className={`retro-pill-btn ${isOrganizer ? 'green' : 'disabled'}`}
                              style={{ padding: '0.25rem 0.65rem', fontSize: '0.72rem' }}
                              title={isOrganizer ? "Resolve incident" : "Organizer login required"}
                            >
                              {!isOrganizer && <Lock size={10} />} Resolve
                            </button>
                          )}
                          {al.status === 'RESOLVED' && (
                            <button
                              onClick={() => isOrganizer ? onOpenAlertAction(al, 'reopen') : onRequireLogin?.()}
                              className={`retro-pill-btn ${isOrganizer ? 'coral' : 'disabled'}`}
                              style={{ padding: '0.25rem 0.65rem', fontSize: '0.72rem' }}
                              title={isOrganizer ? "Reopen incident" : "Organizer login required"}
                            >
                              {!isOrganizer && <Lock size={10} />} Reopen
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

      <IncidentReportModal
        isOpen={showReportModal}
        onClose={() => setShowReportModal(false)}
        alerts={safeAlerts}
        sensors={safeSensors}
      />
    </div>
  );
};
export default AlertsPage;
