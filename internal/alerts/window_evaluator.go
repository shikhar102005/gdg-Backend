package alerts

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

type WindowEvaluator struct {
	store store.Store
}

func NewWindowEvaluator(s store.Store) *WindowEvaluator {
	return &WindowEvaluator{store: s}
}

// EvaluateTimeWindowRules iterates through all sensors and active time-window rules
func (we *WindowEvaluator) EvaluateTimeWindowRules(ctx context.Context) ([]*models.Alert, error) {
	rules, err := we.store.ListAlertRules(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}

	sensors, err := we.store.ListSensors(ctx, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to list sensors: %w", err)
	}

	var createdAlerts []*models.Alert
	now := time.Now().UTC()

	for _, rule := range rules {
		if rule.RuleClass != models.RuleClassTimeWindow || !rule.Enabled {
			continue
		}

		duration := time.Duration(rule.WindowDurationSeconds) * time.Second
		if duration <= 0 {
			duration = 10 * time.Minute
		}
		windowStart := now.Add(-duration)

		for _, sensor := range sensors {
			if !MatchesScope(&rule, &sensor) {
				continue
			}

			readings, err := we.store.GetReadingsInWindow(ctx, sensor.ID, rule.Metric, windowStart, now)
			if err != nil {
				continue
			}

			if len(readings) == 0 {
				continue
			}

			aggValue, hasValue := ComputeAggregation(readings, rule.WindowAggregation)
			if !hasValue {
				continue
			}

			triggered := CompareOperator(aggValue, rule.Operator, rule.Threshold)
			if !triggered {
				continue
			}

			dedupKey := GenerateDeduplicationKey(&rule, &sensor)
			alert := &models.Alert{
				ID:               uuid.New(),
				RuleID:           rule.ID,
				RuleName:         rule.Name,
				SensorID:         sensor.ID,
				SensorName:       sensor.Name,
				Campus:           sensor.Campus,
				Building:         sensor.Building,
				DeduplicationKey: dedupKey,
				Severity:         rule.Severity,
				Status:           models.AlertStatusOpen,
				TriggerValue:     aggValue,
				TriggerTime:      now,
				WindowStart:      &windowStart,
				WindowEnd:        &now,
			}

			created, _, err := we.store.CreateAlertIfNotActive(ctx, alert)
			if err != nil {
				continue
			}
			if created {
				_ = we.store.InsertAuditLog(ctx, &models.AuditLog{
					EntityType: "alert",
					EntityID:   alert.ID.String(),
					Action:     models.AuditCreate,
					Actor:      "system:window-engine",
					Details: map[string]interface{}{
						"rule_id":            rule.ID.String(),
						"rule_name":          rule.Name,
						"sensor_id":          sensor.ID,
						"window_aggregation": rule.WindowAggregation,
						"window_seconds":     rule.WindowDurationSeconds,
						"trigger_value":      aggValue,
						"threshold":          rule.Threshold,
						"readings_evaluated": len(readings),
					},
					CreatedAt: time.Now(),
				})
				createdAlerts = append(createdAlerts, alert)
			}
		}
	}

	return createdAlerts, nil
}

// ComputeAggregation calculates avg, min, max, or count over a set of readings
func ComputeAggregation(readings []models.Reading, agg models.WindowAggregation) (float64, bool) {
	if len(readings) == 0 {
		return 0, false
	}

	switch agg {
	case models.AggCount:
		return float64(len(readings)), true

	case models.AggMin:
		minVal := readings[0].Value
		for _, r := range readings[1:] {
			if r.Value < minVal {
				minVal = r.Value
			}
		}
		return minVal, true

	case models.AggMax:
		maxVal := readings[0].Value
		for _, r := range readings[1:] {
			if r.Value > maxVal {
				maxVal = r.Value
			}
		}
		return maxVal, true

	case models.AggAvg:
		fallthrough
	default:
		sum := 0.0
		for _, r := range readings {
			sum += r.Value
		}
		avg := sum / float64(len(readings))
		return math.Round(avg*100) / 100, true
	}
}
