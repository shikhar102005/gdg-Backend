package alerts

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

type Evaluator struct {
	store store.Store
}

func NewEvaluator(s store.Store) *Evaluator {
	return &Evaluator{store: s}
}

// EvaluateSingleReading checks a single incoming reading against all enabled single-reading rules
func (e *Evaluator) EvaluateSingleReading(ctx context.Context, reading *models.Reading, sensor *models.Sensor, rules []models.AlertRule) ([]*models.Alert, error) {
	var triggeredAlerts []*models.Alert

	for _, rule := range rules {
		if !rule.Enabled || rule.RuleClass != models.RuleClassSingleReading {
			continue
		}

		if !MatchesScope(&rule, sensor) {
			continue
		}

		if !strings.EqualFold(rule.Metric, reading.MetricType) {
			continue
		}

		triggered := CompareOperator(reading.Value, rule.Operator, rule.Threshold)
		if !triggered {
			continue
		}

		dedupKey := GenerateDeduplicationKey(&rule, sensor)
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
			TriggerValue:     reading.Value,
			TriggerTime:      reading.ObservedAt,
		}

		created, _, err := e.store.CreateAlertIfNotActive(ctx, alert)
		if err != nil {
			return nil, err
		}
		if created {
			// Record audit log for alert creation
			_ = e.store.InsertAuditLog(ctx, &models.AuditLog{
				EntityType: "alert",
				EntityID:   alert.ID.String(),
				Action:     models.AuditCreate,
				Actor:      "system:alert-engine",
				Details: map[string]interface{}{
					"rule_id":       rule.ID.String(),
					"rule_name":     rule.Name,
					"sensor_id":     sensor.ID,
					"trigger_value": reading.Value,
					"threshold":     rule.Threshold,
					"operator":      rule.Operator,
				},
				CreatedAt: time.Now(),
			})
			triggeredAlerts = append(triggeredAlerts, alert)
		}
	}

	return triggeredAlerts, nil
}

// MatchesScope checks whether a rule scope applies to the given sensor
func MatchesScope(rule *models.AlertRule, sensor *models.Sensor) bool {
	switch rule.ScopeType {
	case models.ScopeGlobal, "":
		return true
	case models.ScopeSensor:
		return strings.EqualFold(rule.ScopeValue, sensor.ID)
	case models.ScopeTypeKind:
		return strings.EqualFold(rule.ScopeValue, sensor.SensorType)
	case models.ScopeBuilding:
		return strings.EqualFold(rule.ScopeValue, sensor.Building)
	case models.ScopeCampus:
		return strings.EqualFold(rule.ScopeValue, sensor.Campus)
	default:
		return false
	}
}

// CompareOperator evaluates standard arithmetic comparison operators
func CompareOperator(value float64, operator string, threshold float64) bool {
	const epsilon = 0.000001
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold || math.Abs(value-threshold) < epsilon
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold || math.Abs(value-threshold) < epsilon
	case "==", "=":
		return math.Abs(value-threshold) < epsilon
	case "!=":
		return math.Abs(value-threshold) >= epsilon
	default:
		return false
	}
}
