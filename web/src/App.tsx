import React, { useEffect, useState } from 'react';
import { Alert, AlertRule, DashboardSummary, Sensor, SimulatorStatus, UserProfile } from './types';
import { api } from './api/client';
import { Navbar } from './components/Navbar';
import { AlertActionModal } from './components/AlertActionModal';
import { RuleFormModal } from './components/RuleFormModal';
import { OverviewPage } from './pages/OverviewPage';
import { SensorsPage } from './pages/SensorsPage';
import { TrendsPage } from './pages/TrendsPage';
import { AlertsPage } from './pages/AlertsPage';
import { RulesPage } from './pages/RulesPage';
import { SimulatorPage } from './pages/SimulatorPage';
import { ErrorBoundary } from './components/ErrorBoundary';

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState('overview');
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [sensors, setSensors] = useState<Sensor[]>([]);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [simulator, setSimulator] = useState<SimulatorStatus | null>(null);
  const [user, setUser] = useState<UserProfile | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date>(new Date());

  // Modals state
  const [activeAlert, setActiveAlert] = useState<Alert | null>(null);
  const [alertAction, setAlertAction] = useState<'acknowledge' | 'resolve' | 'reopen' | null>(null);
  const [showCreateRule, setShowCreateRule] = useState(false);

  const fetchDashboardData = async () => {
    try {
      const [sumRes, senRes, altRes, rulRes, simRes] = await Promise.all([
        api.getDashboardSummary().catch(() => null),
        api.getSensors().catch(() => ({ sensors: [], count: 0 })),
        api.getAlerts().catch(() => ({ alerts: [], count: 0 })),
        api.getRules().catch(() => ({ rules: [], count: 0 })),
        api.getSimulatorStatus().catch(() => null),
      ]);

      if (sumRes) setSummary(sumRes);
      if (senRes?.sensors) setSensors(senRes.sensors);
      if (altRes?.alerts) setAlerts(altRes.alerts);
      if (rulRes?.rules) setRules(rulRes.rules);
      if (simRes) setSimulator(simRes);
      setLastUpdated(new Date());
    } catch (err) {
      console.error('Failed to refresh data', err);
    }
  };

  useEffect(() => {
    fetchDashboardData();
    // Poll telemetry data every 4 seconds for a dynamic, live dashboard feel
    const interval = setInterval(fetchDashboardData, 4000);
    return () => clearInterval(interval);
  }, []);

  // Alert actions
  const handleOpenAlertAction = (alert: Alert, action: 'acknowledge' | 'resolve' | 'reopen') => {
    setActiveAlert(alert);
    setAlertAction(action);
  };

  const handleAlertActionSubmit = async (alertId: string, action: string, notes?: string) => {
    // 1. Optimistic update in UI
    const targetStatus = action === 'acknowledge' ? 'ACKNOWLEDGED' : action === 'resolve' ? 'RESOLVED' : 'OPEN';
    setAlerts((prev) =>
      prev.map((a) =>
        a.id === alertId
          ? {
              ...a,
              status: targetStatus as any,
              acknowledged_by: action === 'acknowledge' ? (user?.email || 'organizer') : a.acknowledged_by,
              resolved_by: action === 'resolve' ? (user?.email || 'organizer') : a.resolved_by,
              resolution_notes: notes || a.resolution_notes,
            }
          : a
      )
    );

    // 2. Execute API action
    try {
      if (action === 'acknowledge') {
        await api.acknowledgeAlert(alertId);
      } else if (action === 'resolve') {
        await api.resolveAlert(alertId, notes || 'Resolved via organizer dashboard');
      } else if (action === 'reopen') {
        await api.reopenAlert(alertId);
      }
    } catch (err) {
      console.error('Failed alert action', err);
    }
    // 3. Background sync
    await fetchDashboardData();
  };

  // Rule actions
  const handleToggleRule = async (id: string, enabled: boolean) => {
    // 1. Optimistic toggle in UI
    setRules((prev) => prev.map((r) => (r.id === id ? { ...r, enabled } : r)));

    // 2. Execute API toggle
    try {
      await api.toggleRule(id, enabled);
    } catch (err) {
      console.error('Failed to toggle rule', err);
      // Revert if API failed
      setRules((prev) => prev.map((r) => (r.id === id ? { ...r, enabled: !enabled } : r)));
    }
    // 3. Background sync
    await fetchDashboardData();
  };

  const handleCreateRule = async (rule: Partial<AlertRule>) => {
    await api.createRule(rule);
    await fetchDashboardData();
  };

  const handleTriggerSpike = async (metric?: string, sensorId?: string, value?: number) => {
    const targetMetric = metric || 'temperature';
    const targetSensor = sensorId || 'mb-srv-temp';
    let targetVal = value;
    if (!targetVal) {
      switch (targetMetric) {
        case 'temperature': targetVal = 39.5; break;
        case 'co2': targetVal = 1250; break;
        case 'humidity': targetVal = 88; break;
        case 'occupancy': targetVal = 550; break;
        case 'noise': targetVal = 92; break;
        case 'pm25': targetVal = 70; break;
        default: targetVal = 40.0;
      }
    }
    await api.triggerSpike(targetMetric, targetSensor, targetVal);
    await fetchDashboardData();
  };

  const safeAlerts = alerts || [];
  const activeAlertsCount = safeAlerts.filter((a) => a.status === 'OPEN' || a.status === 'ACKNOWLEDGED').length;

  return (
    <ErrorBoundary fallbackTitle="CampusPulse Application Error">
      <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
        <Navbar
          activeTab={activeTab}
          setActiveTab={setActiveTab}
          activeAlertsCount={activeAlertsCount}
          lastUpdated={lastUpdated}
          onRefresh={fetchDashboardData}
          user={user}
          setUser={setUser}
        />

        <main className="main-container" style={{ flex: 1 }}>
          {activeTab === 'overview' && (
            <OverviewPage
              summary={summary}
              sensors={sensors}
              simulator={simulator}
              onOpenAlertAction={handleOpenAlertAction}
              onSelectTab={setActiveTab}
              onTriggerSpike={handleTriggerSpike}
              onRefresh={fetchDashboardData}
            />
          )}

          {activeTab === 'sensors' && (
            <SensorsPage
              sensors={sensors}
              onRefresh={fetchDashboardData}
            />
          )}

          {activeTab === 'trends' && (
            <TrendsPage
              sensors={sensors}
            />
          )}

          {activeTab === 'alerts' && (
            <AlertsPage
              alerts={alerts}
              sensors={sensors}
              onRefresh={fetchDashboardData}
              onOpenAlertAction={handleOpenAlertAction}
            />
          )}

          {activeTab === 'rules' && (
            <RulesPage
              rules={rules}
              onToggleRule={handleToggleRule}
              onOpenCreateRule={() => setShowCreateRule(true)}
              onRefresh={fetchDashboardData}
            />
          )}

          {activeTab === 'simulator' && (
            <SimulatorPage
              simulator={simulator}
              onRefresh={fetchDashboardData}
            />
          )}
        </main>

        <footer className="retro-footer">
          <div style={{ maxWidth: '1400px', margin: '0 auto', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.75rem' }}>
            <div>
              <strong>CampusPulse Telemetry Hub</strong> &bull; Environmental Sensing & Incident Control
            </div>
            <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
              <span>Go &bull; Supabase PostgreSQL</span>
              <span>React &bull; TypeScript &bull; Vite</span>
              <span className="retro-badge active">10,000 Readings/sec Capable</span>
            </div>
          </div>
        </footer>

        {/* Action Modals */}
        <AlertActionModal
          alert={activeAlert}
          action={alertAction}
          onClose={() => {
            setActiveAlert(null);
            setAlertAction(null);
          }}
          onSubmit={handleAlertActionSubmit}
        />

        <RuleFormModal
          isOpen={showCreateRule}
          onClose={() => setShowCreateRule(false)}
          onSubmit={handleCreateRule}
        />
      </div>
    </ErrorBoundary>
  );
};
export default App;
