package models

import (
	"time"

	"github.com/google/uuid"
)

type SensorStatus string

const (
	SensorStatusActive      SensorStatus = "active"
	SensorStatusOffline     SensorStatus = "offline"
	SensorStatusMaintenance SensorStatus = "maintenance"
)

type Sensor struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Campus     string                 `json:"campus"`
	Building   string                 `json:"building"`
	Location   string                 `json:"location"`
	SensorType string                 `json:"sensor_type"`
	Unit       string                 `json:"unit"`
	Status     SensorStatus           `json:"status"`
	APIKeyHash string                 `json:"-"`
	Metadata   map[string]interface{} `json:"metadata"`
	LastSeenAt *time.Time             `json:"last_seen_at,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type Reading struct {
	ID             int64                  `json:"id,omitempty"`
	SensorID       string                 `json:"sensor_id"`
	ObservedAt     time.Time              `json:"observed_at"`
	ReceivedAt     time.Time              `json:"received_at"`
	MetricType     string                 `json:"metric_type"`
	Value          float64                `json:"value"`
	Quality        string                 `json:"quality"`
	IdempotencyKey string                 `json:"idempotency_key,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type IngestReadingRequest struct {
	SensorID       string                 `json:"sensorId"`
	Type           string                 `json:"type"`
	Value          float64                `json:"value"`
	Timestamp      string                 `json:"timestamp"`
	Quality        string                 `json:"quality,omitempty"`
	IdempotencyKey string                 `json:"idempotencyKey,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type BatchIngestRequest struct {
	Readings []IngestReadingRequest `json:"readings"`
}

type BatchItemResult struct {
	Index   int    `json:"index"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type BatchIngestResponse struct {
	TotalProcessed int               `json:"total_processed"`
	AcceptedCount  int               `json:"accepted_count"`
	FailedCount    int               `json:"failed_count"`
	DurationMs     int64             `json:"duration_ms"`
	Errors         []BatchItemResult `json:"errors,omitempty"`
}

type RuleClass string

const (
	RuleClassSingleReading RuleClass = "single_reading"
	RuleClassTimeWindow    RuleClass = "time_window"
)

type ScopeType string

const (
	ScopeGlobal   ScopeType = "global"
	ScopeSensor   ScopeType = "sensor"
	ScopeTypeKind ScopeType = "type"
	ScopeBuilding ScopeType = "building"
	ScopeCampus   ScopeType = "campus"
)

type WindowAggregation string

const (
	AggAvg   WindowAggregation = "avg"
	AggMin   WindowAggregation = "min"
	AggMax   WindowAggregation = "max"
	AggCount WindowAggregation = "count"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

type AlertRule struct {
	ID                    uuid.UUID         `json:"id"`
	Name                  string            `json:"name"`
	Description           string            `json:"description"`
	Enabled               bool              `json:"enabled"`
	ScopeType             ScopeType         `json:"scope_type"`
	ScopeValue            string            `json:"scope_value"`
	Metric                string            `json:"metric"`
	RuleClass             RuleClass         `json:"rule_class"`
	Operator              string            `json:"operator"` // '>', '>=', '<', '<=', '==', '!='
	Threshold             float64           `json:"threshold"`
	WindowDurationSeconds int               `json:"window_duration_seconds"`
	WindowAggregation     WindowAggregation `json:"window_aggregation"`
	Severity              Severity          `json:"severity"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
}

type AlertStatus string

const (
	AlertStatusOpen         AlertStatus = "OPEN"
	AlertStatusAcknowledged AlertStatus = "ACKNOWLEDGED"
	AlertStatusResolved     AlertStatus = "RESOLVED"
)

type Alert struct {
	ID               uuid.UUID   `json:"id"`
	RuleID           uuid.UUID   `json:"rule_id"`
	RuleName         string      `json:"rule_name,omitempty"`
	SensorID         string      `json:"sensor_id"`
	SensorName       string      `json:"sensor_name,omitempty"`
	Campus           string      `json:"campus,omitempty"`
	Building         string      `json:"building,omitempty"`
	DeduplicationKey string      `json:"deduplication_key"`
	Severity         Severity    `json:"severity"`
	Status           AlertStatus `json:"status"`
	TriggerValue     float64     `json:"trigger_value"`
	TriggerTime      time.Time   `json:"trigger_time"`
	WindowStart      *time.Time  `json:"window_start,omitempty"`
	WindowEnd        *time.Time  `json:"window_end,omitempty"`
	AcknowledgedBy   string      `json:"acknowledged_by,omitempty"`
	AcknowledgedAt   *time.Time  `json:"acknowledged_at,omitempty"`
	ResolvedBy       string      `json:"resolved_by,omitempty"`
	ResolvedAt       *time.Time  `json:"resolved_at,omitempty"`
	ResolutionNotes  string      `json:"resolution_notes,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

type AuditAction string

const (
	AuditCreate      AuditAction = "CREATE"
	AuditUpdate      AuditAction = "UPDATE"
	AuditEnable      AuditAction = "ENABLE"
	AuditDisable     AuditAction = "DISABLE"
	AuditAcknowledge AuditAction = "ACKNOWLEDGE"
	AuditResolve     AuditAction = "RESOLVE"
	AuditReopen      AuditAction = "REOPEN"
)

type AuditLog struct {
	ID         uuid.UUID              `json:"id"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Action     AuditAction            `json:"action"`
	Actor      string                 `json:"actor"`
	Details    map[string]interface{} `json:"details"`
	CreatedAt  time.Time              `json:"created_at"`
}

type DashboardSummary struct {
	TotalSensors      int                `json:"total_sensors"`
	ActiveSensors     int                `json:"active_sensors"`
	OfflineSensors    int                `json:"offline_sensors"`
	TotalReadings     int64              `json:"total_readings"`
	ActiveAlertsCount int                `json:"active_alerts_count"`
	CriticalAlerts    int                `json:"critical_alerts"`
	WarningAlerts     int                `json:"warning_alerts"`
	InfoAlerts        int                `json:"info_alerts"`
	RecentAlerts      []Alert            `json:"recent_alerts"`
	SensorsByBuilding map[string]int     `json:"sensors_by_building"`
	ReadingsPerMinute float64            `json:"readings_per_minute"`
	LastIngestAt      *time.Time         `json:"last_ingest_at,omitempty"`
}

type TrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Avg       float64   `json:"avg"`
	Min       float64   `json:"min"`
	Max       float64   `json:"max"`
	Count     int       `json:"count"`
}

type SimulatorConfig struct {
	Running       bool    `json:"running"`
	RatePerSecond int     `json:"rate_per_second"`
	TriggerSpike  bool    `json:"trigger_spike"`
	SpikeMetric   string  `json:"spike_metric"`
	SpikeSensorID string  `json:"spike_sensor_id"`
	SpikeValue    float64 `json:"spike_value"`
	ActiveSensors int     `json:"active_sensors"`
	TotalEmitted  int64   `json:"total_emitted"`
	LastEmittedAt *time.Time `json:"last_emitted_at,omitempty"`
}
