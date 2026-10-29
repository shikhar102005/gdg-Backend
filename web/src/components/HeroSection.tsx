import React from 'react';

interface HeroSectionProps {
  onStartClick: () => void;
  totalSensors: number;
  totalReadings: number;
}

export const HeroSection: React.FC<HeroSectionProps> = ({ onStartClick, totalSensors, totalReadings }) => {
  return (
    <section className="retro-hero">
      {/* Floating Letter Badges with Google themed colors */}
      <div className="floating-tile tile-blue" style={{ top: '15px', left: '12%', transform: 'rotate(-8deg)' }}>
        C
      </div>
      <div className="floating-tile tile-red" style={{ top: '8px', left: '32%', transform: 'rotate(5deg)' }}>
        H
      </div>
      <div className="floating-tile tile-yellow" style={{ top: '18px', right: '12%', transform: 'rotate(10deg)' }}>
        E
      </div>
      <div className="floating-tile tile-blue" style={{ bottom: '25px', left: '8%', transform: 'rotate(6deg)' }}>
        I
      </div>
      <div className="floating-tile tile-green" style={{ bottom: '30px', right: '16%', transform: 'rotate(-12deg)' }}>
        F
      </div>

      {/* Decorative Doodles */}
      {/* Paper note with pencil doodle (left) */}
      <div style={{ position: 'absolute', top: '10px', left: '18%', width: '60px', opacity: 0.95, pointerEvents: 'none' }}>
        <svg viewBox="0 0 100 100" fill="none" stroke="#202124" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M20 15 L70 15 L85 30 L85 90 L20 90 Z" fill="#ffffff" />
          <path d="M70 15 L70 30 L85 30" fill="#f8fafd" />
          <line x1="30" y1="40" x2="65" y2="40" />
          <line x1="30" y1="52" x2="65" y2="52" />
          <line x1="30" y1="64" x2="55" y2="64" />
          {/* Pencil in Google Red */}
          <path d="M75 75 L88 62 L94 68 L81 81 Z" fill="#ea4335" />
          <polygon points="75,75 71,83 79,79" fill="#202124" />
        </svg>
      </div>

      {/* Paper airplane doodle (right) in Google Yellow */}
      <div style={{ position: 'absolute', top: '15px', right: '22%', width: '56px', opacity: 0.95, pointerEvents: 'none' }}>
        <svg viewBox="0 0 100 80" fill="none" stroke="#202124" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M10 40 L88 12 L50 68 L42 45 Z" fill="#fbbc04" />
          <line x1="88" y1="12" x2="42" y2="45" stroke="#202124" />
          <path d="M42 45 L32 58 L30 42" fill="#fef7e0" />
          {/* Trailing loop squiggle */}
          <path d="M8 45 Q-5 55 10 65 Q25 75 15 82" stroke="#202124" strokeDasharray="3 3" />
        </svg>
      </div>

      {/* Center Hero Content */}
      <div className="retro-hero-inner">
        {/* Isometric 3D Cube Icon with Google Colors */}
        <div className="isometric-cube">
          <svg viewBox="0 0 100 100" fill="none" stroke="#202124" strokeWidth="3">
            {/* Top face: Google Blue */}
            <polygon points="50,10 88,30 50,50 12,30" fill="#4285f4" />
            {/* Left face: Google Green */}
            <polygon points="12,30 50,50 50,92 12,72" fill="#34a853" />
            {/* Right face: Google Red */}
            <polygon points="50,50 88,30 88,72 50,92" fill="#ea4335" />
            {/* Inner sensor grid lines in Google Yellow */}
            <line x1="31" y1="20" x2="69" y2="40" stroke="#fbbc04" strokeWidth="2.5" />
            <line x1="69" y1="20" x2="31" y2="40" stroke="#fbbc04" strokeWidth="2.5" />
          </svg>
        </div>

        <h1 className="retro-hero-title">CampusPulse</h1>
        <p className="retro-hero-subtitle">
          Real-Time IoT Environmental Telemetry & Incident Control
        </p>

        {/* Start / Jump to dashboard pill */}
        <button
          onClick={onStartClick}
          className="retro-pill-btn blue"
          style={{
            fontSize: '1.05rem',
            padding: '0.65rem 2rem',
            boxShadow: '3px 3px 0px #202124',
            fontWeight: 800,
          }}
        >
          Start &nbsp;↓
        </button>

        <div style={{ marginTop: '0.85rem', fontSize: '0.82rem', fontWeight: 700, color: '#5f6368' }}>
          <span>{totalSensors} Sensors Active</span>
          <span style={{ margin: '0 0.5rem' }}>&bull;</span>
          <span>{totalReadings.toLocaleString()} Readings Ingested</span>
          <span style={{ margin: '0 0.5rem' }}>&bull;</span>
          <span>10K Readings/sec Capable</span>
        </div>
      </div>
    </section>
  );
};
export default HeroSection;
