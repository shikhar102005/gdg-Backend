import { Alert, AlertRule, DashboardSummary, Reading, Sensor, SimulatorStatus, TrendPoint, UserProfile } from '../types';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8081/api/v1';

class ApiClient {
  private token: string | null = null;

  constructor() {
    this.token = localStorage.getItem('campus_auth_token');
  }

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem('campus_auth_token', token);
    } else {
      localStorage.removeItem('campus_auth_token');
    }
  }

  getToken(): string | null {
    return this.token;
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(errorData.error || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // Auth
  async login(email: string, pass: string): Promise<UserProfile> {
    const data = await this.request<{ token: string; email: string; role: 'admin' | 'organizer' | 'viewer' }>(
      '/auth/login',
      {
        method: 'POST',
        body: JSON.stringify({ email, password: pass }),
      }
    );
    this.setToken(data.token);
    return data;
  }

  async getMe(): Promise<UserProfile> {
    const data = await this.request<{ user_id: string; email: string; role: 'admin' | 'organizer' | 'viewer' }>('/auth/me');
    return {
      email: data.email,
      role: data.role,
      token: this.token || '',
    };
  }

  // Dashboard
  async getDashboardSummary(): Promise<DashboardSummary> {
    const data = await this.request<DashboardSummary>('/dashboard/summary');
    return {
      ...data,
      recent_alerts: data.recent_alerts || [],
      sensors_by_building: data.sensors_by_building || {},
    };
  }

  // Sensors
  async getSensors(building?: string, sensorType?: string, status?: string): Promise<{ sensors: Sensor[]; count: number }> {
    const params = new URLSearchParams();
    if (building) params.append('building', building);
    if (sensorType) params.append('sensor_type', sensorType);
    if (status) params.append('status', status);
    const query = params.toString() ? `?${params.toString()}` : '';
    const res = await this.request<{ sensors: Sensor[]; count: number }>(`/sensors${query}`);
    return { sensors: res.sensors || [], count: res.count || (res.sensors ? res.sensors.length : 0) };
  }

  async getSensor(id: string): Promise<Sensor> {
    return this.request<Sensor>(`/sensors/${id}`);
  }

  async getSensorHistory(id: string, limit: number = 50): Promise<{ readings: Reading[]; count: number }> {
    const res = await this.request<{ readings: Reading[]; count: number }>(`/sensors/${id}/history?limit=${limit}`);
    return { readings: res.readings || [], count: res.count || (res.readings ? res.readings.length : 0) };
  }

  // Trends
  async getMetricTrends(metric: string, building?: string, sensorId?: string, duration: string = '1h'): Promise<{ points: TrendPoint[]; metric: string }> {
    const params = new URLSearchParams({ metric, duration });
    if (building) params.append('building', building);
    if (sensorId) params.append('sensor_id', sensorId);
    const res = await this.request<{ points: TrendPoint[]; metric: string }>(`/readings/trends?${params.toString()}`);
    return { points: res.points || [], metric: res.metric || metric };
  }

  // Alerts
  async getAlerts(status?: string, severity?: string, building?: string, sensorId?: string): Promise<{ alerts: Alert[]; count: number }> {
    const params = new URLSearchParams();
    if (status) params.append('status', status);
    if (severity) params.append('severity', severity);
    if (building) params.append('building', building);
    if (sensorId) params.append('sensor_id', sensorId);
    const query = params.toString() ? `?${params.toString()}` : '';
    const res = await this.request<{ alerts: Alert[]; count: number }>(`/alerts${query}`);
    return { alerts: res.alerts || [], count: res.count || (res.alerts ? res.alerts.length : 0) };
  }

  async acknowledgeAlert(id: string): Promise<Alert> {
    return this.request<Alert>(`/alerts/${id}/acknowledge`, { method: 'POST' });
  }

  async resolveAlert(id: string, notes: string): Promise<Alert> {
    return this.request<Alert>(`/alerts/${id}/resolve`, {
      method: 'POST',
      body: JSON.stringify({ notes }),
    });
  }

  async reopenAlert(id: string): Promise<Alert> {
    return this.request<Alert>(`/alerts/${id}/reopen`, { method: 'POST' });
  }

  // Rules
  async getRules(enabledOnly: boolean = false): Promise<{ rules: AlertRule[]; count: number }> {
    const res = await this.request<{ rules: AlertRule[]; count: number }>(`/rules?enabled_only=${enabledOnly}`);
    return { rules: res.rules || [], count: res.count || (res.rules ? res.rules.length : 0) };
  }

  async createRule(rule: Partial<AlertRule>): Promise<AlertRule> {
    return this.request<AlertRule>('/rules', {
      method: 'POST',
      body: JSON.stringify(rule),
    });
  }

  async updateRule(id: string, rule: Partial<AlertRule>): Promise<AlertRule> {
    return this.request<AlertRule>(`/rules/${id}`, {
      method: 'PUT',
      body: JSON.stringify(rule),
    });
  }

  async toggleRule(id: string, enabled: boolean): Promise<{ id: string; enabled: boolean }> {
    return this.request<{ id: string; enabled: boolean }>(`/rules/${id}/toggle`, {
      method: 'PATCH',
      body: JSON.stringify({ enabled }),
    });
  }

  // Simulator
  async getSimulatorStatus(): Promise<SimulatorStatus> {
    return this.request<SimulatorStatus>('/simulator/status');
  }

  async startSimulator(rate: number): Promise<{ status: string }> {
    return this.request<{ status: string }>('/simulator/start', {
      method: 'POST',
      body: JSON.stringify({ rate_per_second: rate }),
    });
  }

  async stopSimulator(): Promise<{ status: string }> {
    return this.request<{ status: string }>('/simulator/stop', { method: 'POST' });
  }

  async triggerSpike(metric: string, sensorId: string, value: number): Promise<{ status: string }> {
    return this.request<{ status: string }>('/simulator/trigger-spike', {
      method: 'POST',
      body: JSON.stringify({ metric, sensor_id: sensorId, value }),
    });
  }
}

export const api = new ApiClient();
