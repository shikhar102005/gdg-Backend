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
        {/* Top Navbar Row matching Original Layout with Google Palette */}
        <div className="retro-top-bar">
          {/* Left: Hex badge + Overview View button */}
          <div className="retro-brand-group">
            <div className="retro-hex-logo" title="Campus Sensor Hub">
              ⬡
            </div>
            <button
              onClick={() => setActiveTab('overview')}
              className="retro-pill-btn yellow"
            >
              Overview View
            </button>
          </div>

          {/* Center: Notice banner pill matching Original Layout */}
          <div className="retro-banner-pill">
            <span>📡 Telemetry Stream Active: Continuous Ingestion to Supabase</span>
            <button
              onClick={onRefresh}
              style={{
                background: 'transparent',
                border: 'none',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                padding: '0 0.25rem',
              }}
              title="Refresh telemetry"
            >
              <RefreshCw size={13} />
            </button>
          </div>

          {/* Right: Active Alerts button + Log In button */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>
            <button
              onClick={() => setActiveTab('alerts')}
              className="retro-pill-btn coral"
            >
              Active Alerts {activeAlertsCount > 0 ? `(${activeAlertsCount})` : ''}
            </button>

            {user ? (
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <span className="retro-badge active" style={{ fontSize: '0.8rem', padding: '0.35rem 0.75rem' }}>
                  <Shield size={12} style={{ marginRight: '0.25rem' }} /> {user.email.split('@')[0]}
                </span>
                <button
                  onClick={handleLogout}
                  className="retro-pill-btn"
                  style={{ padding: '0.35rem 0.85rem', fontSize: '0.8rem' }}
                >
                  Log Out
                </button>
              </div>
            ) : (
              <button
                onClick={() => setShowLoginModal(true)}
                className="retro-pill-btn green"
              >
                Log In
              </button>
            )}
          </div>
        </div>

        {/* Bottom Navigation Pills Row (Original Layout) */}
        <nav className="retro-nav-tabs">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => setActiveTab(item.id)}
                className={`retro-nav-tab ${isActive ? 'active' : ''}`}
              >
                <Icon size={15} />
                <span>{item.label}</span>
                {item.badge !== undefined && item.badge > 0 && (
                  <span
                    style={{
                      background: isActive ? '#ffffff' : '#ea4335',
                      color: isActive ? '#ea4335' : '#ffffff',
                      borderRadius: '9999px',
                      padding: '1px 7px',
                      fontSize: '0.72rem',
                      fontWeight: 800,
                      marginLeft: '0.2rem',
                    }}
                  >
                    {item.badge}
                  </span>
                )}
              </button>
            );
          })}
        </nav>
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
