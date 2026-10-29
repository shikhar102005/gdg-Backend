import React from 'react';
import { LucideIcon } from 'lucide-react';

interface MetricCardProps {
  title: string;
  value: string | number;
  subtext?: string;
  icon: LucideIcon;
  badge?: {
    text: string;
    variant: 'active' | 'warning' | 'critical' | 'info';
  };
}

export const MetricCard: React.FC<MetricCardProps> = ({ title, value, subtext, icon: Icon, badge }) => {
  return (
    <div className="retro-metric-card">
      <div className="retro-metric-header">
        <span className="retro-metric-title">{title}</span>
        <div style={{
          background: '#f8fafd',
          border: '1.5px solid #202124',
          borderRadius: '8px',
          padding: '4px',
          display: 'flex',
          boxShadow: '1.5px 1.5px 0px #202124',
        }}>
          <Icon size={16} color="#202124" />
        </div>
      </div>
      <div className="retro-metric-value">{value}</div>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginTop: '0.5rem' }}>
        {subtext && <span className="retro-metric-sub">{subtext}</span>}
        {badge && (
          <span className={`retro-badge ${badge.variant}`}>
            {badge.text}
          </span>
        )}
      </div>
    </div>
  );
};
export default MetricCard;
