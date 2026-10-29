package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shikhar102005/gdg-Backend/internal/models"
)

type Store interface {
	// Sensors
	GetSensor(ctx context.Context, id string) (*models.Sensor, error)
	ListSensors(ctx context.Context, building, sensorType, status string) ([]models.Sensor, error)
	UpsertSensor(ctx context.Context, sensor *models.Sensor) error
	UpdateSensorLastSeen(ctx context.Context, id string, lastSeen time.Time) error
	UpdateSensorStatus(ctx context.Context, id string, status models.SensorStatus) error

	// Readings
	InsertReadingBatch(ctx context.Context, readings []models.Reading) (accepted int, err error)
	GetSensorReadings(ctx context.Context, sensorID string, limit int, since *time.Time) ([]models.Reading, error)
	GetMetricTrends(ctx context.Context, metric string, sensorID string, building string, duration time.Duration, bucketSize time.Duration) ([]models.TrendPoint, error)
	GetTotalReadingsCount(ctx context.Context) (int64, error)
	GetReadingsInWindow(ctx context.Context, sensorID string, metric string, start, end time.Time) ([]models.Reading, error)

	// Alert Rules
	ListAlertRules(ctx context.Context, enabledOnly bool) ([]models.AlertRule, error)
	GetAlertRule(ctx context.Context, id uuid.UUID) (*models.AlertRule, error)
	CreateAlertRule(ctx context.Context, rule *models.AlertRule) error
	UpdateAlertRule(ctx context.Context, rule *models.AlertRule) error
	ToggleAlertRule(ctx context.Context, id uuid.UUID, enabled bool) error

	// Alerts
	ListAlerts(ctx context.Context, status, severity, building, sensorID string, limit, offset int) ([]models.Alert, error)
	GetAlert(ctx context.Context, id uuid.UUID) (*models.Alert, error)
	CreateAlertIfNotActive(ctx context.Context, alert *models.Alert) (created bool, existing *models.Alert, err error)
	AcknowledgeAlert(ctx context.Context, id uuid.UUID, actor string) (*models.Alert, error)
	ResolveAlert(ctx context.Context, id uuid.UUID, actor, notes string) (*models.Alert, error)
	ReopenAlert(ctx context.Context, id uuid.UUID, actor string) (*models.Alert, error)
	GetActiveAlertsCount(ctx context.Context) (total int, critical int, warning int, info int, err error)

	// Audit Logs
	InsertAuditLog(ctx context.Context, log *models.AuditLog) error
	ListAuditLogs(ctx context.Context, entityType, entityID string, limit int) ([]models.AuditLog, error)

	// Dashboard Summary
	GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error)

	// Health
	Ping(ctx context.Context) error
	Close()
}
