package store

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shikhar102005/gdg-Backend/internal/models"
)

type MemoryStore struct {
	mu           sync.RWMutex
	sensors      map[string]*models.Sensor
	readings     []models.Reading
	idempKeys    map[string]struct{} // "sensorID:idempKey"
	rules        map[uuid.UUID]*models.AlertRule
	alerts       map[uuid.UUID]*models.Alert
	activeDedup  map[string]uuid.UUID // deduplicationKey -> alertID
	auditLogs    []models.AuditLog
	lastIngestAt *time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sensors:     make(map[string]*models.Sensor),
		readings:    make([]models.Reading, 0, 100000),
		idempKeys:   make(map[string]struct{}),
		rules:       make(map[uuid.UUID]*models.AlertRule),
		alerts:      make(map[uuid.UUID]*models.Alert),
		activeDedup: make(map[string]uuid.UUID),
		auditLogs:   make([]models.AuditLog, 0),
	}
}

func (s *MemoryStore) Ping(ctx context.Context) error {
	return nil
}

func (s *MemoryStore) Close() {}

// ---------------- Sensors ----------------

func (s *MemoryStore) GetSensor(ctx context.Context, id string) (*models.Sensor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sensor, ok := s.sensors[id]
	if !ok {
		return nil, fmt.Errorf("sensor not found: %s", id)
	}
	cp := *sensor
	return &cp, nil
}

func (s *MemoryStore) ListSensors(ctx context.Context, building, sensorType, status string) ([]models.Sensor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Sensor, 0, len(s.sensors))
	for _, sen := range s.sensors {
		if building != "" && !strings.EqualFold(sen.Building, building) {
			continue
		}
		if sensorType != "" && !strings.EqualFold(sen.SensorType, sensorType) {
			continue
		}
		if status != "" && !strings.EqualFold(string(sen.Status), status) {
			continue
		}
		result = append(result, *sen)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (s *MemoryStore) UpsertSensor(ctx context.Context, sensor *models.Sensor) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if existing, ok := s.sensors[sensor.ID]; ok {
		existing.Name = sensor.Name
		existing.Campus = sensor.Campus
		existing.Building = sensor.Building
		existing.Location = sensor.Location
		existing.SensorType = sensor.SensorType
		existing.Unit = sensor.Unit
		existing.Status = sensor.Status
		if sensor.APIKeyHash != "" {
			existing.APIKeyHash = sensor.APIKeyHash
		}
		if sensor.Metadata != nil {
			existing.Metadata = sensor.Metadata
		}
		existing.UpdatedAt = now
		return nil
	}

	cp := *sensor
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = now
	}
	cp.UpdatedAt = now
	s.sensors[sensor.ID] = &cp
	return nil
}

func (s *MemoryStore) UpdateSensorLastSeen(ctx context.Context, id string, lastSeen time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sen, ok := s.sensors[id]; ok {
		sen.LastSeenAt = &lastSeen
		sen.Status = models.SensorStatusActive
		sen.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("sensor not found: %s", id)
}

func (s *MemoryStore) UpdateSensorStatus(ctx context.Context, id string, status models.SensorStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sen, ok := s.sensors[id]; ok {
		sen.Status = status
		sen.UpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("sensor not found: %s", id)
}

// ---------------- Readings ----------------

func (s *MemoryStore) InsertReadingBatch(ctx context.Context, batch []models.Reading) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accepted := 0
	now := time.Now()
	s.lastIngestAt = &now

	for _, r := range batch {
		if r.IdempotencyKey != "" {
			key := fmt.Sprintf("%s:%s", r.SensorID, r.IdempotencyKey)
			if _, exists := s.idempKeys[key]; exists {
				continue // Skip duplicate idempotent event
			}
			s.idempKeys[key] = struct{}{}
		}

		if sen, ok := s.sensors[r.SensorID]; ok {
			sen.LastSeenAt = &r.ObservedAt
			sen.Status = models.SensorStatusActive
		}

		s.readings = append(s.readings, r)
		accepted++
	}

	// Keep bounded memory for readings if exceeding 500,000 in memory
	if len(s.readings) > 500000 {
		s.readings = s.readings[len(s.readings)-400000:]
	}

	return accepted, nil
}

func (s *MemoryStore) GetSensorReadings(ctx context.Context, sensorID string, limit int, since *time.Time) ([]models.Reading, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	result := make([]models.Reading, 0, limit)
	for i := len(s.readings) - 1; i >= 0 && len(result) < limit; i-- {
		r := s.readings[i]
		if r.SensorID != sensorID {
			continue
		}
		if since != nil && r.ObservedAt.Before(*since) {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

func (s *MemoryStore) GetMetricTrends(ctx context.Context, metric string, sensorID string, building string, duration time.Duration, bucketSize time.Duration) ([]models.TrendPoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoff := time.Now().Add(-duration)
	if bucketSize <= 0 {
		bucketSize = time.Minute
	}

	type bucketAgg struct {
		sum   float64
		min   float64
		max   float64
		count int
	}

	buckets := make(map[int64]*bucketAgg)

	for _, r := range s.readings {
		if r.ObservedAt.Before(cutoff) {
			continue
		}
		if metric != "" && !strings.EqualFold(r.MetricType, metric) {
			continue
		}
		if sensorID != "" && r.SensorID != sensorID {
			continue
		}
		if building != "" {
			sen, ok := s.sensors[r.SensorID]
			if !ok || !strings.EqualFold(sen.Building, building) {
				continue
			}
		}

		tBucket := r.ObservedAt.Truncate(bucketSize).Unix()
		b, exists := buckets[tBucket]
		if !exists {
			b = &bucketAgg{sum: r.Value, min: r.Value, max: r.Value, count: 1}
			buckets[tBucket] = b
		} else {
			b.sum += r.Value
			if r.Value < b.min {
				b.min = r.Value
			}
			if r.Value > b.max {
				b.max = r.Value
			}
			b.count++
		}
	}

	points := make([]models.TrendPoint, 0, len(buckets))
	for tUnix, b := range buckets {
		points = append(points, models.TrendPoint{
			Timestamp: time.Unix(tUnix, 0).UTC(),
			Avg:       b.sum / float64(b.count),
			Min:       b.min,
			Max:       b.max,
			Count:     b.count,
		})
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].Timestamp.Before(points[j].Timestamp)
	})

	return points, nil
}

func (s *MemoryStore) GetTotalReadingsCount(ctx context.Context) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int64(len(s.readings)), nil
}

func (s *MemoryStore) GetReadingsInWindow(ctx context.Context, sensorID string, metric string, start, end time.Time) ([]models.Reading, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.Reading
	for _, r := range s.readings {
		if r.SensorID == sensorID && strings.EqualFold(r.MetricType, metric) {
			if (r.ObservedAt.Equal(start) || r.ObservedAt.After(start)) && (r.ObservedAt.Equal(end) || r.ObservedAt.Before(end)) {
				result = append(result, r)
			}
		}
	}
	return result, nil
}

// ---------------- Alert Rules ----------------

func (s *MemoryStore) ListAlertRules(ctx context.Context, enabledOnly bool) ([]models.AlertRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.AlertRule, 0, len(s.rules))
	for _, r := range s.rules {
		if enabledOnly && !r.Enabled {
			continue
		}
		result = append(result, *r)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result, nil
}

func (s *MemoryStore) GetAlertRule(ctx context.Context, id uuid.UUID) (*models.AlertRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rule, ok := s.rules[id]
	if !ok {
		return nil, fmt.Errorf("rule not found: %s", id)
	}
	cp := *rule
	return &cp, nil
}

func (s *MemoryStore) CreateAlertRule(ctx context.Context, rule *models.AlertRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	cp := *rule
	s.rules[rule.ID] = &cp
	return nil
}

func (s *MemoryStore) UpdateAlertRule(ctx context.Context, rule *models.AlertRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.rules[rule.ID]
	if !ok {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}

	existing.Name = rule.Name
	existing.Description = rule.Description
	existing.Enabled = rule.Enabled
	existing.ScopeType = rule.ScopeType
	existing.ScopeValue = rule.ScopeValue
	existing.Metric = rule.Metric
	existing.RuleClass = rule.RuleClass
	existing.Operator = rule.Operator
	existing.Threshold = rule.Threshold
	existing.WindowDurationSeconds = rule.WindowDurationSeconds
	existing.WindowAggregation = rule.WindowAggregation
	existing.Severity = rule.Severity
	existing.UpdatedAt = time.Now()
	return nil
}

func (s *MemoryStore) ToggleAlertRule(ctx context.Context, id uuid.UUID, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.rules[id]
	if !ok {
		return fmt.Errorf("rule not found: %s", id)
	}
	existing.Enabled = enabled
	existing.UpdatedAt = time.Now()
	return nil
}

// ---------------- Alerts ----------------

func (s *MemoryStore) ListAlerts(ctx context.Context, status, severity, building, sensorID string, limit, offset int) ([]models.Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matched []models.Alert
	for _, al := range s.alerts {
		if status != "" && !strings.EqualFold(string(al.Status), status) {
			continue
		}
		if severity != "" && !strings.EqualFold(string(al.Severity), severity) {
			continue
		}
		if sensorID != "" && al.SensorID != sensorID {
			continue
		}
		if building != "" && !strings.EqualFold(al.Building, building) {
			continue
		}
		matched = append(matched, *al)
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})

	if offset >= len(matched) {
		return []models.Alert{}, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(matched) {
		end = len(matched)
	}
	return matched[offset:end], nil
}

func (s *MemoryStore) GetAlert(ctx context.Context, id uuid.UUID) (*models.Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	al, ok := s.alerts[id]
	if !ok {
		return nil, fmt.Errorf("alert not found: %s", id)
	}
	cp := *al
	return &cp, nil
}

func (s *MemoryStore) CreateAlertIfNotActive(ctx context.Context, alert *models.Alert) (bool, *models.Alert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if active alert already exists with this deduplication key
	if existingID, exists := s.activeDedup[alert.DeduplicationKey]; exists {
		if existing, ok := s.alerts[existingID]; ok && (existing.Status == models.AlertStatusOpen || existing.Status == models.AlertStatusAcknowledged) {
			cp := *existing
			return false, &cp, nil
		}
	}

	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	now := time.Now()
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = now
	}
	alert.UpdatedAt = now
	if alert.Status == "" {
		alert.Status = models.AlertStatusOpen
	}

	// Populate sensor / rule details if available
	if sen, ok := s.sensors[alert.SensorID]; ok {
		alert.SensorName = sen.Name
		alert.Campus = sen.Campus
		alert.Building = sen.Building
	}
	if rule, ok := s.rules[alert.RuleID]; ok {
		alert.RuleName = rule.Name
	}

	cp := *alert
	s.alerts[alert.ID] = &cp
	s.activeDedup[alert.DeduplicationKey] = alert.ID

	return true, &cp, nil
}

func (s *MemoryStore) AcknowledgeAlert(ctx context.Context, id uuid.UUID, actor string) (*models.Alert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	al, ok := s.alerts[id]
	if !ok {
		return nil, fmt.Errorf("alert not found: %s", id)
	}
	if al.Status == models.AlertStatusResolved {
		return nil, fmt.Errorf("cannot acknowledge resolved alert")
	}

	now := time.Now()
	al.Status = models.AlertStatusAcknowledged
	al.AcknowledgedBy = actor
	al.AcknowledgedAt = &now
	al.UpdatedAt = now

	cp := *al
	return &cp, nil
}

func (s *MemoryStore) ResolveAlert(ctx context.Context, id uuid.UUID, actor, notes string) (*models.Alert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	al, ok := s.alerts[id]
	if !ok {
		return nil, fmt.Errorf("alert not found: %s", id)
	}

	now := time.Now()
	al.Status = models.AlertStatusResolved
	al.ResolvedBy = actor
	al.ResolvedAt = &now
	al.ResolutionNotes = notes
	al.UpdatedAt = now

	// Clear active deduplication entry so future violations can trigger a new alert
	delete(s.activeDedup, al.DeduplicationKey)

	cp := *al
	return &cp, nil
}

func (s *MemoryStore) ReopenAlert(ctx context.Context, id uuid.UUID, actor string) (*models.Alert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	al, ok := s.alerts[id]
	if !ok {
		return nil, fmt.Errorf("alert not found: %s", id)
	}
	if al.Status != models.AlertStatusResolved {
		return nil, fmt.Errorf("only resolved alerts can be reopened")
	}

	// Check if another alert already has the active dedup key
	if existingID, exists := s.activeDedup[al.DeduplicationKey]; exists && existingID != al.ID {
		return nil, fmt.Errorf("another active alert already exists for this rule and sensor")
	}

	now := time.Now()
	al.Status = models.AlertStatusOpen
	al.UpdatedAt = now
	s.activeDedup[al.DeduplicationKey] = al.ID

	cp := *al
	return &cp, nil
}

func (s *MemoryStore) GetActiveAlertsCount(ctx context.Context) (int, int, int, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total, crit, warn, info := 0, 0, 0, 0
	for _, al := range s.alerts {
		if al.Status == models.AlertStatusOpen || al.Status == models.AlertStatusAcknowledged {
			total++
			switch al.Severity {
			case models.SeverityCritical:
				crit++
			case models.SeverityWarning:
				warn++
			case models.SeverityInfo:
				info++
			}
		}
	}
	return total, crit, warn, info, nil
}

// ---------------- Audit Logs ----------------

func (s *MemoryStore) InsertAuditLog(ctx context.Context, log *models.AuditLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	s.auditLogs = append(s.auditLogs, *log)
	return nil
}

func (s *MemoryStore) ListAuditLogs(ctx context.Context, entityType, entityID string, limit int) ([]models.AuditLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}

	var result []models.AuditLog
	for i := len(s.auditLogs) - 1; i >= 0 && len(result) < limit; i-- {
		l := s.auditLogs[i]
		if entityType != "" && !strings.EqualFold(l.EntityType, entityType) {
			continue
		}
		if entityID != "" && l.EntityID != entityID {
			continue
		}
		result = append(result, l)
	}
	return result, nil
}

// ---------------- Dashboard Summary ----------------

func (s *MemoryStore) GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeSensors := 0
	offlineSensors := 0
	byBuilding := make(map[string]int)

	for _, sen := range s.sensors {
		if sen.Status == models.SensorStatusActive {
			activeSensors++
		} else {
			offlineSensors++
		}
		byBuilding[sen.Building]++
	}

	totalAlerts, crit, warn, info := 0, 0, 0, 0
	var recentAlerts []models.Alert
	for _, al := range s.alerts {
		if al.Status == models.AlertStatusOpen || al.Status == models.AlertStatusAcknowledged {
			totalAlerts++
			switch al.Severity {
			case models.SeverityCritical:
				crit++
			case models.SeverityWarning:
				warn++
			case models.SeverityInfo:
				info++
			}
		}
		recentAlerts = append(recentAlerts, *al)
	}

	sort.Slice(recentAlerts, func(i, j int) bool {
		return recentAlerts[i].CreatedAt.After(recentAlerts[j].CreatedAt)
	})
	if len(recentAlerts) > 5 {
		recentAlerts = recentAlerts[:5]
	}

	// Calculate readings rate in last 1 minute
	oneMinAgo := time.Now().Add(-1 * time.Minute)
	countLastMin := 0
	for i := len(s.readings) - 1; i >= 0; i-- {
		if s.readings[i].ObservedAt.After(oneMinAgo) {
			countLastMin++
		} else {
			break
		}
	}

	return &models.DashboardSummary{
		TotalSensors:      len(s.sensors),
		ActiveSensors:     activeSensors,
		OfflineSensors:    offlineSensors,
		TotalReadings:     int64(len(s.readings)),
		ActiveAlertsCount: totalAlerts,
		CriticalAlerts:    crit,
		WarningAlerts:     warn,
		InfoAlerts:        info,
		RecentAlerts:      recentAlerts,
		SensorsByBuilding: byBuilding,
		ReadingsPerMinute: float64(countLastMin),
		LastIngestAt:      s.lastIngestAt,
	}, nil
}
