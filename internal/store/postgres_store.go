package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shikhar102005/gdg-Backend/internal/config"
	"github.com/shikhar102005/gdg-Backend/internal/models"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, cfg *config.Config) (*PostgresStore, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database url: %w", err)
	}

	poolConfig.MaxConns = cfg.DatabaseMaxConns
	poolConfig.MinConns = cfg.DatabaseMinConns
	poolConfig.MaxConnLifetime = cfg.DatabaseMaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.DatabaseMaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *PostgresStore) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

// ---------------- Sensors ----------------

func (s *PostgresStore) GetSensor(ctx context.Context, id string) (*models.Sensor, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, campus, building, location, sensor_type, unit, status, api_key_hash, metadata, last_seen_at, created_at, updated_at
		FROM sensors WHERE id = $1
	`, id)

	var sen models.Sensor
	var metaBytes []byte
	err := row.Scan(
		&sen.ID, &sen.Name, &sen.Campus, &sen.Building, &sen.Location,
		&sen.SensorType, &sen.Unit, &sen.Status, &sen.APIKeyHash,
		&metaBytes, &sen.LastSeenAt, &sen.CreatedAt, &sen.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(metaBytes) > 0 {
		_ = json.Unmarshal(metaBytes, &sen.Metadata)
	}
	return &sen, nil
}

func (s *PostgresStore) ListSensors(ctx context.Context, building, sensorType, status string) ([]models.Sensor, error) {
	query := `SELECT id, name, campus, building, location, sensor_type, unit, status, api_key_hash, metadata, last_seen_at, created_at, updated_at FROM sensors WHERE 1=1`
	var args []interface{}
	idx := 1

	if building != "" {
		query += fmt.Sprintf(" AND LOWER(building) = LOWER($%d)", idx)
		args = append(args, building)
		idx++
	}
	if sensorType != "" {
		query += fmt.Sprintf(" AND LOWER(sensor_type) = LOWER($%d)", idx)
		args = append(args, sensorType)
		idx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND LOWER(status) = LOWER($%d)", idx)
		args = append(args, status)
		idx++
	}
	query += " ORDER BY id ASC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.Sensor, 0)
	for rows.Next() {
		var sen models.Sensor
		var metaBytes []byte
		if err := rows.Scan(
			&sen.ID, &sen.Name, &sen.Campus, &sen.Building, &sen.Location,
			&sen.SensorType, &sen.Unit, &sen.Status, &sen.APIKeyHash,
			&metaBytes, &sen.LastSeenAt, &sen.CreatedAt, &sen.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if len(metaBytes) > 0 {
			_ = json.Unmarshal(metaBytes, &sen.Metadata)
		}
		result = append(result, sen)
	}
	return result, nil
}

func (s *PostgresStore) UpsertSensor(ctx context.Context, sen *models.Sensor) error {
	metaJSON, _ := json.Marshal(sen.Metadata)
	query := `
		INSERT INTO sensors (id, name, campus, building, location, sensor_type, unit, status, api_key_hash, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			campus = EXCLUDED.campus,
			building = EXCLUDED.building,
			location = EXCLUDED.location,
			sensor_type = EXCLUDED.sensor_type,
			unit = EXCLUDED.unit,
			status = EXCLUDED.status,
			api_key_hash = CASE WHEN EXCLUDED.api_key_hash != '' THEN EXCLUDED.api_key_hash ELSE sensors.api_key_hash END,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
	`
	_, err := s.pool.Exec(ctx, query,
		sen.ID, sen.Name, sen.Campus, sen.Building, sen.Location,
		sen.SensorType, sen.Unit, sen.Status, sen.APIKeyHash, metaJSON,
	)
	return err
}

func (s *PostgresStore) UpdateSensorLastSeen(ctx context.Context, id string, lastSeen time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sensors SET last_seen_at = $2, status = 'active', updated_at = NOW()
		WHERE id = $1
	`, id, lastSeen)
	return err
}

func (s *PostgresStore) UpdateSensorStatus(ctx context.Context, id string, status models.SensorStatus) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sensors SET status = $2, updated_at = NOW()
		WHERE id = $1
	`, id, status)
	return err
}

// ---------------- Readings ----------------

func (s *PostgresStore) InsertReadingBatch(ctx context.Context, readings []models.Reading) (int, error) {
	if len(readings) == 0 {
		return 0, nil
	}

	// High-performance CopyFrom bulk ingestion
	rows := make([][]interface{}, 0, len(readings))
	sensorLastSeenMap := make(map[string]time.Time)

	for _, r := range readings {
		metaJSON, _ := json.Marshal(r.Metadata)
		var idemp *string
		if r.IdempotencyKey != "" {
			k := r.IdempotencyKey
			idemp = &k
		}
		rows = append(rows, []interface{}{
			r.SensorID,
			r.ObservedAt,
			r.ReceivedAt,
			r.MetricType,
			r.Value,
			r.Quality,
			idemp,
			metaJSON,
		})
		sensorLastSeenMap[r.SensorID] = r.ObservedAt
	}

	copyCount, err := s.pool.CopyFrom(
		ctx,
		pgx.Identifier{"readings"},
		[]string{"sensor_id", "observed_at", "received_at", "metric_type", "value", "quality", "idempotency_key", "metadata"},
		pgx.CopyFromRows(rows),
	)
	if err == nil {
		s.asyncUpdateSensorLastSeen(sensorLastSeenMap)
		return int(copyCount), nil
	}

	// Fallback to high-speed chunked multi-row INSERT (compatible with PgBouncer & transaction poolers)
	inserted, err := s.insertMultiRow(ctx, readings)
	if err != nil {
		log.Printf("[store] insertMultiRow failed: %v", err)
		return inserted, err
	}
	s.asyncUpdateSensorLastSeen(sensorLastSeenMap)
	return inserted, nil
}

func (s *PostgresStore) asyncUpdateSensorLastSeen(m map[string]time.Time) {
	if len(m) == 0 {
		return
	}
	go func(touched map[string]time.Time) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		batch := &pgx.Batch{}
		for sid, ls := range touched {
			batch.Queue(`UPDATE sensors SET last_seen_at = $2, status = 'active', updated_at = NOW() WHERE id = $1`, sid, ls)
		}
		br := s.pool.SendBatch(bgCtx, batch)
		_ = br.Close()
	}(m)
}

func (s *PostgresStore) insertMultiRow(ctx context.Context, readings []models.Reading) (int, error) {
	if len(readings) == 0 {
		return 0, nil
	}

	chunkSize := 500
	totalInserted := 0

	for i := 0; i < len(readings); i += chunkSize {
		end := i + chunkSize
		if end > len(readings) {
			end = len(readings)
		}
		chunk := readings[i:end]

		var sb strings.Builder
		sb.WriteString("INSERT INTO readings (sensor_id, observed_at, received_at, metric_type, value, quality, idempotency_key, metadata) VALUES ")
		args := make([]interface{}, 0, len(chunk)*8)

		paramIdx := 1
		for j, r := range chunk {
			if j > 0 {
				sb.WriteString(", ")
			}
			metaJSON, _ := json.Marshal(r.Metadata)
			var idemp *string
			if r.IdempotencyKey != "" {
				k := r.IdempotencyKey
				idemp = &k
			}
			fmt.Fprintf(&sb, "($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				paramIdx, paramIdx+1, paramIdx+2, paramIdx+3, paramIdx+4, paramIdx+5, paramIdx+6, paramIdx+7)
			args = append(args, r.SensorID, r.ObservedAt, r.ReceivedAt, r.MetricType, r.Value, r.Quality, idemp, metaJSON)
			paramIdx += 8
		}
		sb.WriteString(" ON CONFLICT (sensor_id, idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING")

		ct, err := s.pool.Exec(ctx, sb.String(), args...)
		if err != nil {
			return totalInserted, fmt.Errorf("multi-row insert failed: %w", err)
		}
		totalInserted += int(ct.RowsAffected())
	}

	return totalInserted, nil
}

func (s *PostgresStore) GetSensorReadings(ctx context.Context, sensorID string, limit int, since *time.Time) ([]models.Reading, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT id, sensor_id, observed_at, received_at, metric_type, value, quality, idempotency_key, metadata
		FROM readings
		WHERE sensor_id = $1
	`
	args := []interface{}{sensorID}
	if since != nil {
		query += " AND observed_at >= $2"
		args = append(args, *since)
	}
	query += fmt.Sprintf(" ORDER BY observed_at DESC LIMIT %d", limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.Reading, 0)
	for rows.Next() {
		var r models.Reading
		var idemp *string
		var metaBytes []byte
		if err := rows.Scan(
			&r.ID, &r.SensorID, &r.ObservedAt, &r.ReceivedAt, &r.MetricType,
			&r.Value, &r.Quality, &idemp, &metaBytes,
		); err != nil {
			return nil, err
		}
		if idemp != nil {
			r.IdempotencyKey = *idemp
		}
		if len(metaBytes) > 0 {
			_ = json.Unmarshal(metaBytes, &r.Metadata)
		}
		result = append(result, r)
	}
	return result, nil
}

func (s *PostgresStore) GetMetricTrends(ctx context.Context, metric string, sensorID string, building string, duration time.Duration, bucketSize time.Duration) ([]models.TrendPoint, error) {
	cutoff := time.Now().Add(-duration)
	bucketSec := int(bucketSize.Seconds())
	if bucketSec <= 0 {
		bucketSec = 60
	}

	query := fmt.Sprintf(`
		SELECT 
			to_timestamp(floor(extract(epoch from r.observed_at) / %d) * %d) AS bucket_time,
			AVG(r.value) AS avg_val,
			MIN(r.value) AS min_val,
			MAX(r.value) AS max_val,
			COUNT(r.id) AS cnt
		FROM readings r
		JOIN sensors s ON r.sensor_id = s.id
		WHERE r.observed_at >= $1
	`, bucketSec, bucketSec)

	var args []interface{}
	args = append(args, cutoff)
	idx := 2

	if metric != "" {
		query += fmt.Sprintf(" AND LOWER(r.metric_type) = LOWER($%d)", idx)
		args = append(args, metric)
		idx++
	}
	if sensorID != "" {
		query += fmt.Sprintf(" AND r.sensor_id = $%d", idx)
		args = append(args, sensorID)
		idx++
	}
	if building != "" {
		query += fmt.Sprintf(" AND LOWER(s.building) = LOWER($%d)", idx)
		args = append(args, building)
		idx++
	}

	query += " GROUP BY bucket_time ORDER BY bucket_time ASC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]models.TrendPoint, 0)
	for rows.Next() {
		var pt models.TrendPoint
		if err := rows.Scan(&pt.Timestamp, &pt.Avg, &pt.Min, &pt.Max, &pt.Count); err != nil {
			return nil, err
		}
		points = append(points, pt)
	}
	return points, nil
}

func (s *PostgresStore) GetTotalReadingsCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM readings").Scan(&count)
	return count, err
}

func (s *PostgresStore) GetReadingsInWindow(ctx context.Context, sensorID string, metric string, start, end time.Time) ([]models.Reading, error) {
	query := `
		SELECT id, sensor_id, observed_at, received_at, metric_type, value, quality
		FROM readings
		WHERE sensor_id = $1 AND LOWER(metric_type) = LOWER($2) AND observed_at >= $3 AND observed_at <= $4
		ORDER BY observed_at ASC
	`
	rows, err := s.pool.Query(ctx, query, sensorID, metric, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Reading
	for rows.Next() {
		var r models.Reading
		if err := rows.Scan(&r.ID, &r.SensorID, &r.ObservedAt, &r.ReceivedAt, &r.MetricType, &r.Value, &r.Quality); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// ---------------- Alert Rules ----------------

func (s *PostgresStore) ListAlertRules(ctx context.Context, enabledOnly bool) ([]models.AlertRule, error) {
	query := `
		SELECT id, name, description, enabled, scope_type, scope_value, metric, rule_class, operator, threshold, window_duration_seconds, window_aggregation, severity, created_at, updated_at
		FROM alert_rules
	`
	if enabledOnly {
		query += " WHERE enabled = true"
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]models.AlertRule, 0)
	for rows.Next() {
		var r models.AlertRule
		if err := rows.Scan(
			&r.ID, &r.Name, &r.Description, &r.Enabled, &r.ScopeType, &r.ScopeValue,
			&r.Metric, &r.RuleClass, &r.Operator, &r.Threshold, &r.WindowDurationSeconds,
			&r.WindowAggregation, &r.Severity, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *PostgresStore) GetAlertRule(ctx context.Context, id uuid.UUID) (*models.AlertRule, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, description, enabled, scope_type, scope_value, metric, rule_class, operator, threshold, window_duration_seconds, window_aggregation, severity, created_at, updated_at
		FROM alert_rules WHERE id = $1
	`, id)

	var r models.AlertRule
	err := row.Scan(
		&r.ID, &r.Name, &r.Description, &r.Enabled, &r.ScopeType, &r.ScopeValue,
		&r.Metric, &r.RuleClass, &r.Operator, &r.Threshold, &r.WindowDurationSeconds,
		&r.WindowAggregation, &r.Severity, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *PostgresStore) CreateAlertRule(ctx context.Context, rule *models.AlertRule) error {
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	query := `
		INSERT INTO alert_rules (id, name, description, enabled, scope_type, scope_value, metric, rule_class, operator, threshold, window_duration_seconds, window_aggregation, severity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	return s.pool.QueryRow(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Enabled, rule.ScopeType, rule.ScopeValue,
		rule.Metric, rule.RuleClass, rule.Operator, rule.Threshold, rule.WindowDurationSeconds,
		rule.WindowAggregation, rule.Severity,
	).Scan(&rule.CreatedAt, &rule.UpdatedAt)
}

func (s *PostgresStore) UpdateAlertRule(ctx context.Context, rule *models.AlertRule) error {
	query := `
		UPDATE alert_rules SET
			name = $2, description = $3, enabled = $4, scope_type = $5, scope_value = $6,
			metric = $7, rule_class = $8, operator = $9, threshold = $10,
			window_duration_seconds = $11, window_aggregation = $12, severity = $13, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	return s.pool.QueryRow(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Enabled, rule.ScopeType, rule.ScopeValue,
		rule.Metric, rule.RuleClass, rule.Operator, rule.Threshold, rule.WindowDurationSeconds,
		rule.WindowAggregation, rule.Severity,
	).Scan(&rule.UpdatedAt)
}

func (s *PostgresStore) ToggleAlertRule(ctx context.Context, id uuid.UUID, enabled bool) error {
	_, err := s.pool.Exec(ctx, "UPDATE alert_rules SET enabled = $2, updated_at = NOW() WHERE id = $1", id, enabled)
	return err
}

// ---------------- Alerts ----------------

func (s *PostgresStore) ListAlerts(ctx context.Context, status, severity, building, sensorID string, limit, offset int) ([]models.Alert, error) {
	query := `
		SELECT 
			a.id, a.rule_id, r.name as rule_name, a.sensor_id, s.name as sensor_name, s.campus, s.building,
			a.deduplication_key, a.severity, a.status, a.trigger_value, a.trigger_time, a.window_start, a.window_end,
			a.acknowledged_by, a.acknowledged_at, a.resolved_by, a.resolved_at, a.resolution_notes,
			a.created_at, a.updated_at
		FROM alerts a
		JOIN alert_rules r ON a.rule_id = r.id
		JOIN sensors s ON a.sensor_id = s.id
		WHERE 1=1
	`
	var args []interface{}
	idx := 1

	if status != "" {
		query += fmt.Sprintf(" AND a.status = $%d", idx)
		args = append(args, strings.ToUpper(status))
		idx++
	}
	if severity != "" {
		query += fmt.Sprintf(" AND a.severity = $%d", idx)
		args = append(args, strings.ToLower(severity))
		idx++
	}
	if building != "" {
		query += fmt.Sprintf(" AND LOWER(s.building) = LOWER($%d)", idx)
		args = append(args, building)
		idx++
	}
	if sensorID != "" {
		query += fmt.Sprintf(" AND a.sensor_id = $%d", idx)
		args = append(args, sensorID)
		idx++
	}

	query += " ORDER BY a.created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := make([]models.Alert, 0)
	for rows.Next() {
		var a models.Alert
		var ackBy, resBy, notes *string
		if err := rows.Scan(
			&a.ID, &a.RuleID, &a.RuleName, &a.SensorID, &a.SensorName, &a.Campus, &a.Building,
			&a.DeduplicationKey, &a.Severity, &a.Status, &a.TriggerValue, &a.TriggerTime, &a.WindowStart, &a.WindowEnd,
			&ackBy, &a.AcknowledgedAt, &resBy, &a.ResolvedAt, &notes,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if ackBy != nil {
			a.AcknowledgedBy = *ackBy
		}
		if resBy != nil {
			a.ResolvedBy = *resBy
		}
		if notes != nil {
			a.ResolutionNotes = *notes
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

func (s *PostgresStore) GetAlert(ctx context.Context, id uuid.UUID) (*models.Alert, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT 
			a.id, a.rule_id, r.name as rule_name, a.sensor_id, s.name as sensor_name, s.campus, s.building,
			a.deduplication_key, a.severity, a.status, a.trigger_value, a.trigger_time, a.window_start, a.window_end,
			a.acknowledged_by, a.acknowledged_at, a.resolved_by, a.resolved_at, a.resolution_notes,
			a.created_at, a.updated_at
		FROM alerts a
		JOIN alert_rules r ON a.rule_id = r.id
		JOIN sensors s ON a.sensor_id = s.id
		WHERE a.id = $1
	`, id)

	var a models.Alert
	var ackBy, resBy, notes *string
	err := row.Scan(
		&a.ID, &a.RuleID, &a.RuleName, &a.SensorID, &a.SensorName, &a.Campus, &a.Building,
		&a.DeduplicationKey, &a.Severity, &a.Status, &a.TriggerValue, &a.TriggerTime, &a.WindowStart, &a.WindowEnd,
		&ackBy, &a.AcknowledgedAt, &resBy, &a.ResolvedAt, &notes,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ackBy != nil {
		a.AcknowledgedBy = *ackBy
	}
	if resBy != nil {
		a.ResolvedBy = *resBy
	}
	if notes != nil {
		a.ResolutionNotes = *notes
	}
	return &a, nil
}

func (s *PostgresStore) CreateAlertIfNotActive(ctx context.Context, alert *models.Alert) (bool, *models.Alert, error) {
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	query := `
		INSERT INTO alerts (id, rule_id, sensor_id, deduplication_key, severity, status, trigger_value, trigger_time, window_start, window_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'OPEN', $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (deduplication_key) WHERE status IN ('OPEN', 'ACKNOWLEDGED') DO NOTHING
		RETURNING id, created_at, updated_at
	`
	err := s.pool.QueryRow(ctx, query,
		alert.ID, alert.RuleID, alert.SensorID, alert.DeduplicationKey,
		alert.Severity, alert.TriggerValue, alert.TriggerTime, alert.WindowStart, alert.WindowEnd,
	).Scan(&alert.ID, &alert.CreatedAt, &alert.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			// An active alert already exists! Query existing
			existing, getErr := s.getActiveAlertByDedupKey(ctx, alert.DeduplicationKey)
			return false, existing, getErr
		}
		return false, nil, err
	}

	return true, alert, nil
}

func (s *PostgresStore) getActiveAlertByDedupKey(ctx context.Context, key string) (*models.Alert, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, rule_id, sensor_id, deduplication_key, severity, status, trigger_value, trigger_time, window_start, window_end,
		       acknowledged_by, acknowledged_at, resolved_by, resolved_at, resolution_notes, created_at, updated_at
		FROM alerts
		WHERE deduplication_key = $1 AND status IN ('OPEN', 'ACKNOWLEDGED')
		LIMIT 1
	`, key)
	var a models.Alert
	var ackBy, resBy, notes *string
	err := row.Scan(
		&a.ID, &a.RuleID, &a.SensorID, &a.DeduplicationKey, &a.Severity, &a.Status,
		&a.TriggerValue, &a.TriggerTime, &a.WindowStart, &a.WindowEnd,
		&ackBy, &a.AcknowledgedAt, &resBy, &a.ResolvedAt, &notes,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ackBy != nil {
		a.AcknowledgedBy = *ackBy
	}
	if resBy != nil {
		a.ResolvedBy = *resBy
	}
	if notes != nil {
		a.ResolutionNotes = *notes
	}
	return &a, nil
}

func (s *PostgresStore) AcknowledgeAlert(ctx context.Context, id uuid.UUID, actor string) (*models.Alert, error) {
	query := `
		UPDATE alerts SET
			status = 'ACKNOWLEDGED', acknowledged_by = $2, acknowledged_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status != 'RESOLVED'
		RETURNING id, rule_id, sensor_id, deduplication_key, severity, status, trigger_value, trigger_time,
		          window_start, window_end, acknowledged_by, acknowledged_at, resolved_by, resolved_at, resolution_notes,
		          created_at, updated_at
	`
	var a models.Alert
	var ackBy, resBy, notes *string
	err := s.pool.QueryRow(ctx, query, id, actor).Scan(
		&a.ID, &a.RuleID, &a.SensorID, &a.DeduplicationKey, &a.Severity, &a.Status,
		&a.TriggerValue, &a.TriggerTime, &a.WindowStart, &a.WindowEnd,
		&ackBy, &a.AcknowledgedAt, &resBy, &a.ResolvedAt, &notes,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ackBy != nil {
		a.AcknowledgedBy = *ackBy
	}
	if resBy != nil {
		a.ResolvedBy = *resBy
	}
	if notes != nil {
		a.ResolutionNotes = *notes
	}
	return &a, nil
}

func (s *PostgresStore) ResolveAlert(ctx context.Context, id uuid.UUID, actor, notes string) (*models.Alert, error) {
	query := `
		UPDATE alerts SET
			status = 'RESOLVED', resolved_by = $2, resolved_at = NOW(), resolution_notes = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, rule_id, sensor_id, deduplication_key, severity, status, trigger_value, trigger_time,
		          window_start, window_end, acknowledged_by, acknowledged_at, resolved_by, resolved_at, resolution_notes,
		          created_at, updated_at
	`
	var a models.Alert
	var ackBy, resBy, noteStr *string
	err := s.pool.QueryRow(ctx, query, id, actor, notes).Scan(
		&a.ID, &a.RuleID, &a.SensorID, &a.DeduplicationKey, &a.Severity, &a.Status,
		&a.TriggerValue, &a.TriggerTime, &a.WindowStart, &a.WindowEnd,
		&ackBy, &a.AcknowledgedAt, &resBy, &a.ResolvedAt, &noteStr,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ackBy != nil {
		a.AcknowledgedBy = *ackBy
	}
	if resBy != nil {
		a.ResolvedBy = *resBy
	}
	if noteStr != nil {
		a.ResolutionNotes = *noteStr
	}
	return &a, nil
}

func (s *PostgresStore) ReopenAlert(ctx context.Context, id uuid.UUID, actor string) (*models.Alert, error) {
	query := `
		UPDATE alerts SET
			status = 'OPEN', updated_at = NOW()
		WHERE id = $1 AND status = 'RESOLVED'
		RETURNING id, rule_id, sensor_id, deduplication_key, severity, status, trigger_value, trigger_time,
		          window_start, window_end, acknowledged_by, acknowledged_at, resolved_by, resolved_at, resolution_notes,
		          created_at, updated_at
	`
	var a models.Alert
	var ackBy, resBy, notes *string
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.RuleID, &a.SensorID, &a.DeduplicationKey, &a.Severity, &a.Status,
		&a.TriggerValue, &a.TriggerTime, &a.WindowStart, &a.WindowEnd,
		&ackBy, &a.AcknowledgedAt, &resBy, &a.ResolvedAt, &notes,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ackBy != nil {
		a.AcknowledgedBy = *ackBy
	}
	if resBy != nil {
		a.ResolvedBy = *resBy
	}
	if notes != nil {
		a.ResolutionNotes = *notes
	}
	return &a, nil
}

func (s *PostgresStore) GetActiveAlertsCount(ctx context.Context) (int, int, int, int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT severity, COUNT(*) 
		FROM alerts 
		WHERE status IN ('OPEN', 'ACKNOWLEDGED') 
		GROUP BY severity
	`)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer rows.Close()

	total, crit, warn, info := 0, 0, 0, 0
	for rows.Next() {
		var sev string
		var cnt int
		if err := rows.Scan(&sev, &cnt); err != nil {
			return 0, 0, 0, 0, err
		}
		total += cnt
		switch strings.ToLower(sev) {
		case "critical":
			crit += cnt
		case "warning":
			warn += cnt
		case "info":
			info += cnt
		}
	}
	return total, crit, warn, info, nil
}

// ---------------- Audit Logs ----------------

func (s *PostgresStore) InsertAuditLog(ctx context.Context, log *models.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	detailsJSON, _ := json.Marshal(log.Details)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_logs (id, entity_type, entity_id, action, actor, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, log.ID, log.EntityType, log.EntityID, log.Action, log.Actor, detailsJSON)
	return err
}

func (s *PostgresStore) ListAuditLogs(ctx context.Context, entityType, entityID string, limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, entity_type, entity_id, action, actor, details, created_at FROM audit_logs WHERE 1=1`
	var args []interface{}
	idx := 1
	if entityType != "" {
		query += fmt.Sprintf(" AND entity_type = $%d", idx)
		args = append(args, entityType)
		idx++
	}
	if entityID != "" {
		query += fmt.Sprintf(" AND entity_id = $%d", idx)
		args = append(args, entityID)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d", limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		var detBytes []byte
		if err := rows.Scan(&l.ID, &l.EntityType, &l.EntityID, &l.Action, &l.Actor, &detBytes, &l.CreatedAt); err != nil {
			return nil, err
		}
		if len(detBytes) > 0 {
			_ = json.Unmarshal(detBytes, &l.Details)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// ---------------- Dashboard Summary ----------------

func (s *PostgresStore) GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error) {
	var wg sync.WaitGroup
	var totalSensors, activeSensors, offlineSensors int
	var totalReadings int64
	var totalAlerts, crit, warn, info int
	var recentAlerts []models.Alert
	byBuilding := make(map[string]int)
	var readingsLastMin float64
	var lastIngestAt *time.Time
	var buildingMu sync.Mutex

	wg.Add(5)

	// 1. Sensors total & active
	go func() {
		defer wg.Done()
		var tot, act int
		_ = s.pool.QueryRow(ctx, "SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'active') FROM sensors").Scan(&tot, &act)
		totalSensors = tot
		activeSensors = act
		offlineSensors = tot - act
	}()

	// 2. Total readings
	go func() {
		defer wg.Done()
		_ = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM readings").Scan(&totalReadings)
	}()

	// 3. Alerts counts & recent alerts
	go func() {
		defer wg.Done()
		totalAlerts, crit, warn, info, _ = s.GetActiveAlertsCount(ctx)
		al, _ := s.ListAlerts(ctx, "", "", "", "", 5, 0)
		if al == nil {
			al = []models.Alert{}
		}
		recentAlerts = al
	}()

	// 4. Sensors by building
	go func() {
		defer wg.Done()
		rows, err := s.pool.Query(ctx, "SELECT building, COUNT(*) FROM sensors GROUP BY building")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var b string
				var count int
				if err := rows.Scan(&b, &count); err == nil {
					buildingMu.Lock()
					byBuilding[b] = count
					buildingMu.Unlock()
				}
			}
		}
	}()

	// 5. Ingestion rate & latest ingest timestamp
	go func() {
		defer wg.Done()
		var count float64
		var t *time.Time
		_ = s.pool.QueryRow(ctx, `
			SELECT 
				COUNT(*) FILTER (WHERE observed_at >= NOW() - INTERVAL '1 minute'),
				MAX(received_at)
			FROM readings
		`).Scan(&count, &t)
		readingsLastMin = count
		lastIngestAt = t
	}()

	wg.Wait()

	return &models.DashboardSummary{
		TotalSensors:      totalSensors,
		ActiveSensors:     activeSensors,
		OfflineSensors:    offlineSensors,
		TotalReadings:     totalReadings,
		ActiveAlertsCount: totalAlerts,
		CriticalAlerts:    crit,
		WarningAlerts:     warn,
		InfoAlerts:        info,
		RecentAlerts:      recentAlerts,
		SensorsByBuilding: byBuilding,
		ReadingsPerMinute: readingsLastMin,
		LastIngestAt:      lastIngestAt,
	}, nil
}
