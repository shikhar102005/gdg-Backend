export type SensorStatus = 'active' | 'offline' | 'maintenance';

export interface Sensor {
  id: string;
  name: string;
  campus: string;
  building: string;
  location: string;
  sensor_type: string;
  unit: string;
  status: SensorStatus;
  metadata?: Record<string, any>;
  last_seen_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Reading {
  id?: number;
  sensor_id: string;
  observed_at: string;
  received_at: string;
  metric_type: string;
  value: number;
  quality: string;
  idempotency_key?: string;
  metadata?: Record<string, any>;
}

export type RuleClass = 'single_reading' | 'time_window';
export type ScopeType = 'global' | 'sensor' | 'type' | 'building' | 'campus';
export type WindowAggregation = 'avg' | 'min' | 'max' | 'count';
export type Severity = 'info' | 'warning' | 'critical';

export interface AlertRule {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  scope_type: ScopeType;
  scope_value: string;
  metric: string;
  rule_class: RuleClass;
  operator: string;
  threshold: number;
  window_duration_seconds: number;
  window_aggregation: WindowAggregation;
  severity: Severity;
  created_at: string;
  updated_at: string;
}

export type AlertStatus = 'OPEN' | 'ACKNOWLEDGED' | 'RESOLVED';

export interface Alert {
  id: string;
  rule_id: string;
  rule_name?: string;
  sensor_id: string;
  sensor_name?: string;
  campus?: string;
  building?: string;
  deduplication_key: string;
  severity: Severity;
  status: AlertStatus;
  trigger_value: number;
  trigger_time: string;
  window_start?: string;
  window_end?: string;
  acknowledged_by?: string;
  acknowledged_at?: string;
  resolved_by?: string;
  resolved_at?: string;
  resolution_notes?: string;
  created_at: string;
  updated_at: string;
}

export interface DashboardSummary {
  total_sensors: number;
  active_sensors: number;
  offline_sensors: number;
  total_readings: number;
  active_alerts_count: number;
  critical_alerts: number;
  warning_alerts: number;
  info_alerts: number;
  recent_alerts: Alert[];
  sensors_by_building: Record<string, number>;
  readings_per_minute: number;
  last_ingest_at?: string;
}

export interface TrendPoint {
  timestamp: string;
  avg: number;
  min: number;
  max: number;
  count: number;
}

export interface SimulatorStatus {
  running: boolean;
  rate_per_second: number;
  trigger_spike: boolean;
  spike_metric: string;
  spike_sensor_id: string;
  spike_value: number;
  active_sensors: number;
  total_emitted: number;
  last_emitted_at?: string;
}

export interface UserProfile {
  email: string;
  role: 'admin' | 'organizer' | 'viewer';
  token: string;
}
