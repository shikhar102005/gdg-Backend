import React, { useState } from 'react';
import { Activity, AlertTriangle, Cpu, Gauge, Layers, RefreshCw, Shield, Sliders, User, X } from 'lucide-react';
import { UserProfile } from '../types';
import { api } from '../api/client';

interface NavbarProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
  activeAlertsCount: number;
  lastUpdated: Date;
  onRefresh: () => void;
  user: UserProfile | null;
  setUser: (u: UserProfile | null) => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  activeTab,
  setActiveTab,
  activeAlertsCount,
  lastUpdated,
  onRefresh,
  user,
  setUser,
}) => {
  const [showLoginModal, setShowLoginModal] = useState(false);
  const [email, setEmail] = useState('organizer@campus.edu');
  const [password, setPassword] = useState('CampusOrganizerSecure2026!');
  const [loginError, setLoginError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setLoginError('');
    try {
      const u = await api.login(email, password);
      setUser(u);
      setShowLoginModal(false);
    } catch (err: any) {
      setLoginError(err.message || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    api.setToken(null);
    setUser(null);
  };

  const [isRefreshing, setIsRefreshing] = useState(false);

  const handleRefreshClick = () => {
    setIsRefreshing(true);
    onRefresh();
    setTimeout(() => setIsRefreshing(false), 800);
  };

  const navItems = [
    { id: 'overview', label: 'Overview', icon: Gauge },
    { id: 'sensors', label: 'Sensors Fleet', icon: Cpu },
    { id: 'trends', label: 'Trends & Charts', icon: Activity },
    { id: 'alerts', label: 'Incidents Feed', icon: AlertTriangle, badge: activeAlertsCount },
    { id: 'rules', label: 'Alert Rules', icon: Sliders },
    { id: 'simulator', label: 'Simulator Control', icon: Layers },
  ];

  return (
    <>
      <header className="retro-header">
        <div className="command-navbar">
          {/* Brand Identity */}
          <div className="command-brand" onClick={() => setActiveTab('overview')} role="button" tabIndex={0}>
            <div className="command-logo-icon">
              <Activity size={20} color="#ffffff" strokeWidth={2.5} />
            </div>
            <div className="command-brand-text">
              <div className="command-title-row">
                <span className="command-brand-title">CampusPulse</span>
                <span className="command-live-indicator" title="Connected to continuous telemetry stream">
                  <span className="command-live-dot" /> LIVE
                </span>
              </div>
              <span className="command-brand-subtitle">IoT Facility Telemetry</span>
            </div>
          </div>

          {/* Unified Primary Navigation Tabs */}
          <nav className="command-nav-tabs">
            {navItems.map((item) => {
              const Icon = item.icon;
              const isActive = activeTab === item.id;
              const isAlertTab = item.id === 'alerts';
              return (
                <button
                  key={item.id}
                  onClick={() => setActiveTab(item.id)}
                  className={`command-nav-item ${isActive ? 'active' : ''} ${isAlertTab && activeAlertsCount > 0 ? 'has-alerts' : ''}`}
                >
                  <Icon size={16} />
                  <span>{item.label}</span>
                  {item.badge !== undefined && item.badge > 0 && (
                    <span className="command-alert-badge">
                      {item.badge}
                    </span>
                  )}
                </button>
              );
            })}
          </nav>

          {/* Operational Controls & Auth */}
          <div className="command-controls">
            <button
              onClick={handleRefreshClick}
              className="command-refresh-pill"
              title={`Last synchronized: ${lastUpdated.toLocaleTimeString()}`}
            >
              <RefreshCw size={13} className={isRefreshing ? 'spin-anim' : ''} />
              <span>Sync</span>
            </button>

            {user ? (
              <div className="command-user-group">
                <span className="command-user-badge">
                  <Shield size={12} />
                  <span>{user.email.split('@')[0]}</span>
                </span>
                <button
                  onClick={handleLogout}
                  className="command-logout-btn"
                  title="Sign out of organizer account"
                >
                  Log Out
                </button>
              </div>
            ) : (
              <button
                onClick={() => setShowLoginModal(true)}
                className="command-login-btn"
              >
                Organizer Log In
              </button>
            )}
          </div>
        </div>
      </header>

      {/* Login Modal */}
      {showLoginModal && (
        <div className="retro-modal-overlay" onClick={() => setShowLoginModal(false)}>
          <div className="retro-modal-window" onClick={(e) => e.stopPropagation()}>
            <div className="retro-modal-header">
              <div className="retro-modal-dots">
                <span className="window-dot red" onClick={() => setShowLoginModal(false)} style={{ cursor: 'pointer' }}></span>
                <span className="window-dot yellow"></span>
                <span className="window-dot green"></span>
              </div>
              <div className="retro-modal-title">Organizer Sign In</div>
            </div>

            <form onSubmit={handleLogin}>
              <div className="retro-modal-body">
                {loginError && (
                  <div style={{
                    background: '#fad2cf',
                    border: '2px solid #202124',
                    borderRadius: '8px',
                    padding: '0.5rem',
                    fontSize: '0.8rem',
                    marginBottom: '1rem',
                    fontWeight: 700,
                    color: '#c5221f',
                  }}>
                    {loginError}
                  </div>
                )}
                <div style={{ marginBottom: '1rem' }}>
                  <label style={{ display: 'block', fontWeight: 800, fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                    Organizer Email
                  </label>
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="retro-pill-input"
                    required
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontWeight: 800, fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                    Password
                  </label>
                  <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="retro-pill-input"
                    required
                  />
                </div>
              </div>

              <div className="retro-modal-footer">
                <button
                  type="button"
                  onClick={() => setShowLoginModal(false)}
                  className="retro-pill-btn"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={loading}
                  className="retro-modal-ok-btn"
                >
                  {loading ? 'Authenticating...' : 'Sign In'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </>
  );
};
export default Navbar;
