import React, { useState } from 'react';
import { Copy, Download, FileText, Printer, ShieldAlert, X } from 'lucide-react';
import { Alert, Sensor } from '../types';

interface IncidentReportModalProps {
  isOpen: boolean;
  onClose: () => void;
  alerts: Alert[];
  sensors: Sensor[];
}

export const IncidentReportModal: React.FC<IncidentReportModalProps> = ({
  isOpen,
  onClose,
  alerts,
  sensors,
}) => {
  const [copied, setCopied] = useState(false);

  if (!isOpen) return null;

  const now = new Date();
  const safeAlerts = alerts || [];
  const openAlerts = safeAlerts.filter((a) => a.status === 'OPEN');
  const ackAlerts = safeAlerts.filter((a) => a.status === 'ACKNOWLEDGED');
  const resolvedAlerts = safeAlerts.filter((a) => a.status === 'RESOLVED');

  const reportId = `IR-${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}-${String(Math.floor(1000 + Math.random() * 9000))}`;

  const markdownReport = `# CampusPulse Incident & Telemetry Compliance Report
**Report ID:** ${reportId}
**Generated At:** ${now.toISOString()}
**Environment:** Campus Sensor Fleet & Automated Alert Engine

---

## 1. Executive Summary
- **Total Incidents Recorded:** ${safeAlerts.length}
- **Open Active Incidents:** ${openAlerts.length}
- **Under Active Investigation (Acknowledged):** ${ackAlerts.length}
- **Successfully Remediated (Resolved):** ${resolvedAlerts.length}

---

## 2. Detailed Incident Log
${safeAlerts.length === 0 ? '_No incidents recorded in the active period._' : safeAlerts.map((a, i) => `
### Incident #${i + 1}: ${a.rule_name || a.rule_id}
- **Status:** ${a.status}
- **Severity:** ${a.severity.toUpperCase()}
- **Sensor:** ${a.sensor_name || a.sensor_id}
- **Campus / Zone:** ${a.campus || 'Main Campus'} / ${a.building || 'Campus Zone'}
- **Trigger Value:** ${a.trigger_value}
- **Triggered At:** ${new Date(a.created_at).toLocaleString()}
- **Acknowledged By:** ${a.acknowledged_by || 'Pending'}
- **Resolved By:** ${a.resolved_by || 'Pending'}
- **Root Cause / Remediation Notes:** ${a.resolution_notes || 'Investigation in progress'}
`).join('\n')}

---

## 3. Operational Certification
This automated report is compiled directly from the campus telemetry sensor ingestion stream, single-reading threshold evaluators, and time-window sliding aggregations.
`;

  const handleCopyMarkdown = () => {
    navigator.clipboard.writeText(markdownReport);
    setCopied(true);
    setTimeout(() => setCopied(false), 2500);
  };

  const handleDownloadJSON = () => {
    const data = {
      report_id: reportId,
      generated_at: now.toISOString(),
      summary: {
        total: safeAlerts.length,
        open: openAlerts.length,
        acknowledged: ackAlerts.length,
        resolved: resolvedAlerts.length,
      },
      incidents: safeAlerts,
    };
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `incident-report-${reportId}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handlePrint = () => {
    window.print();
  };

  return (
    <div className="retro-modal-overlay" onClick={onClose}>
      <div
        className="retro-modal-window"
        style={{ maxWidth: '780px', width: '92%' }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Retro Header with colored control dots */}
        <div className="retro-modal-header">
          <div className="retro-modal-dots">
            <span className="window-dot red" onClick={onClose} style={{ cursor: 'pointer' }}></span>
            <span className="window-dot yellow"></span>
            <span className="window-dot green"></span>
          </div>
          <div className="retro-modal-title">Official Incident Report &bull; {reportId}</div>
        </div>

        <div className="retro-modal-body" style={{ maxHeight: '70vh', overflowY: 'auto' }}>
          {/* Header Banner */}
          <div style={{
            background: '#e8f0fe',
            border: '2px solid #202124',
            borderRadius: '12px',
            padding: '1rem',
            marginBottom: '1.25rem',
            boxShadow: '3px 3px 0px #202124',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem' }}>
              <div>
                <h3 style={{ fontSize: '1.1rem', fontWeight: 800, color: '#1a73e8', display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                  <ShieldAlert size={20} color="#ea4335" />
                  Campus Incident Audit & Compliance Report
                </h3>
                <div style={{ fontSize: '0.78rem', color: '#5f6368', fontWeight: 600, marginTop: '0.2rem' }}>
                  Generated on {now.toLocaleDateString()} at {now.toLocaleTimeString()} &bull; ID: <code>{reportId}</code>
                </div>
              </div>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <span className="retro-badge critical">{openAlerts.length} OPEN</span>
                <span className="retro-badge warning">{ackAlerts.length} ACK</span>
                <span className="retro-badge active">{resolvedAlerts.length} RESOLVED</span>
              </div>
            </div>
          </div>

          {/* Quick Metrics */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))', gap: '0.75rem', marginBottom: '1.25rem' }}>
            <div style={{ background: '#f8fafd', border: '1.5px solid #202124', borderRadius: '10px', padding: '0.65rem' }}>
              <div style={{ fontSize: '0.7rem', fontWeight: 800, color: '#5f6368' }}>TOTAL INCIDENTS</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 800, color: '#202124' }}>{safeAlerts.length}</div>
            </div>
            <div style={{ background: '#f8fafd', border: '1.5px solid #202124', borderRadius: '10px', padding: '0.65rem' }}>
              <div style={{ fontSize: '0.7rem', fontWeight: 800, color: '#ea4335' }}>CRITICAL SEVERITY</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 800, color: '#ea4335' }}>
                {safeAlerts.filter((a) => a.severity === 'critical').length}
              </div>
            </div>
            <div style={{ background: '#f8fafd', border: '1.5px solid #202124', borderRadius: '10px', padding: '0.65rem' }}>
              <div style={{ fontSize: '0.7rem', fontWeight: 800, color: '#f29900' }}>WARNING SEVERITY</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 800, color: '#f29900' }}>
                {safeAlerts.filter((a) => a.severity === 'warning').length}
              </div>
            </div>
            <div style={{ background: '#f8fafd', border: '1.5px solid #202124', borderRadius: '10px', padding: '0.65rem' }}>
              <div style={{ fontSize: '0.7rem', fontWeight: 800, color: '#34a853' }}>REMEDIATION RATE</div>
              <div style={{ fontSize: '1.3rem', fontWeight: 800, color: '#34a853' }}>
                {safeAlerts.length > 0 ? `${Math.round((resolvedAlerts.length / safeAlerts.length) * 100)}%` : '100%'}
              </div>
            </div>
          </div>

          {/* Incident Table / List */}
          <div style={{ marginBottom: '1.25rem' }}>
            <h4 style={{ fontSize: '0.9rem', fontWeight: 800, marginBottom: '0.6rem', color: '#202124' }}>
              Incident Inventory & Root Cause Summaries
            </h4>
            {safeAlerts.length === 0 ? (
              <div style={{ padding: '2rem', textAlign: 'center', background: '#f8fafd', border: '1.5px solid #202124', borderRadius: '10px', color: '#5f6368', fontWeight: 700 }}>
                No active or historical incidents recorded.
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.65rem' }}>
                {safeAlerts.map((a, idx) => (
                  <div
                    key={a.id}
                    style={{
                      background: '#ffffff',
                      border: '2px solid #202124',
                      borderRadius: '10px',
                      padding: '0.75rem',
                      boxShadow: '2px 2px 0px #202124',
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '0.4rem', marginBottom: '0.35rem' }}>
                      <div style={{ fontWeight: 800, fontSize: '0.92rem' }}>
                        #{idx + 1} &bull; {a.rule_name || a.rule_id}
                      </div>
                      <div style={{ display: 'flex', gap: '0.35rem' }}>
                        <span className={`retro-badge ${a.severity === 'critical' ? 'critical' : 'warning'}`}>
                          {a.severity}
                        </span>
                        <span className={`retro-badge ${a.status === 'OPEN' ? 'warning' : a.status === 'ACKNOWLEDGED' ? 'info' : 'active'}`}>
                          {a.status}
                        </span>
                      </div>
                    </div>
                    <div style={{ fontSize: '0.78rem', color: '#5f6368', marginBottom: '0.35rem' }}>
                      <strong>Sensor:</strong> {a.sensor_name || a.sensor_id} &bull; <strong>Zone:</strong> {a.building || 'Campus'} &bull; <strong>Trigger:</strong> <span className="font-mono" style={{ fontWeight: 800, color: '#ea4335' }}>{a.trigger_value}</span> &bull; <strong>Time:</strong> {new Date(a.created_at).toLocaleString()}
                    </div>
                    {a.resolution_notes && (
                      <div style={{
                        background: '#e6f4ea',
                        border: '1.5px solid #34a853',
                        borderRadius: '6px',
                        padding: '0.4rem 0.6rem',
                        fontSize: '0.75rem',
                        color: '#0d652d',
                        fontWeight: 600,
                      }}>
                        <strong>Root Cause Resolution:</strong> {a.resolution_notes}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Modal Actions */}
        <div className="retro-modal-footer" style={{ justifyContent: 'space-between', flexWrap: 'wrap', gap: '0.5rem' }}>
          <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
            <button
              type="button"
              onClick={handleCopyMarkdown}
              className="retro-pill-btn yellow"
              style={{ fontSize: '0.8rem', padding: '0.35rem 0.85rem' }}
            >
              <Copy size={13} /> {copied ? 'Copied Markdown!' : 'Copy Markdown'}
            </button>
            <button
              type="button"
              onClick={handleDownloadJSON}
              className="retro-pill-btn blue"
              style={{ fontSize: '0.8rem', padding: '0.35rem 0.85rem' }}
            >
              <Download size={13} /> Download JSON
            </button>
            <button
              type="button"
              onClick={handlePrint}
              className="retro-pill-btn green"
              style={{ fontSize: '0.8rem', padding: '0.35rem 0.85rem' }}
            >
              <Printer size={13} /> Print Report
            </button>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="retro-pill-btn"
            style={{ fontSize: '0.8rem', padding: '0.35rem 1rem' }}
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};
export default IncidentReportModal;
