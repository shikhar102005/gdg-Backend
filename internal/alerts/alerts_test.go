package alerts

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

func TestCompareOperator(t *testing.T) {
	tests := []struct {
		val      float64
		op       string
		thresh   float64
		expected bool
	}{
		{35.5, ">", 30.0, true},
		{30.0, ">", 30.0, false},
		{30.0, ">=", 30.0, true},
		{29.9, ">=", 30.0, false},
		{15.0, "<", 20.0, true},
		{20.0, "<", 20.0, false},
		{20.0, "<=", 20.0, true},
		{20.1, "<=", 20.0, false},
		{25.0, "==", 25.0, true},
		{25.0000001, "==", 25.0, true}, // epsilon
		{25.1, "==", 25.0, false},
		{25.1, "!=", 25.0, true},
		{25.0, "!=", 25.0, false},
	}

	for _, tt := range tests {
		got := CompareOperator(tt.val, tt.op, tt.thresh)
		if got != tt.expected {
			t.Errorf("CompareOperator(%f, %s, %f) = %v; want %v", tt.val, tt.op, tt.thresh, got, tt.expected)
		}
	}
}

func TestMatchesScope(t *testing.T) {
	sensor := &models.Sensor{
		ID:         "sensor-tech-101",
		SensorType: "temperature",
		Building:   "Technology Tower",
		Campus:     "Vellore",
	}

	tests := []struct {
		scopeType  models.ScopeType
		scopeValue string
		expected   bool
	}{
		{models.ScopeGlobal, "", true},
		{models.ScopeSensor, "sensor-tech-101", true},
		{models.ScopeSensor, "other-sensor", false},
		{models.ScopeTypeKind, "temperature", true},
		{models.ScopeTypeKind, "humidity", false},
		{models.ScopeBuilding, "Technology Tower", true},
		{models.ScopeBuilding, "Main Building", false},
		{models.ScopeCampus, "Vellore", true},
		{models.ScopeCampus, "Chennai", false},
	}

	for _, tt := range tests {
		rule := &models.AlertRule{
			ScopeType:  tt.scopeType,
			ScopeValue: tt.scopeValue,
		}
		if got := MatchesScope(rule, sensor); got != tt.expected {
			t.Errorf("MatchesScope(scope=%s, val=%s) = %v; want %v", tt.scopeType, tt.scopeValue, got, tt.expected)
		}
	}
}

func TestComputeAggregation(t *testing.T) {
	readings := []models.Reading{
		{Value: 10.0},
		{Value: 20.0},
		{Value: 30.0},
		{Value: 40.0},
	}

	avg, ok := ComputeAggregation(readings, models.AggAvg)
	if !ok || avg != 25.0 {
		t.Fatalf("expected avg 25.0, got %f", avg)
	}

	min, ok := ComputeAggregation(readings, models.AggMin)
	if !ok || min != 10.0 {
		t.Fatalf("expected min 10.0, got %f", min)
	}

	max, ok := ComputeAggregation(readings, models.AggMax)
	if !ok || max != 40.0 {
		t.Fatalf("expected max 40.0, got %f", max)
	}

	cnt, ok := ComputeAggregation(readings, models.AggCount)
	if !ok || cnt != 4.0 {
		t.Fatalf("expected count 4.0, got %f", cnt)
	}
}

func TestEvaluator_SingleReadingAndDeduplication(t *testing.T) {
	ctx := context.Background()
	memStore := store.NewMemoryStore()
	evaluator := NewEvaluator(memStore)

	sensor := &models.Sensor{
		ID:         "sensor-lib-01",
		Name:       "Library Floor 2",
		Campus:     "Main Campus",
		Building:   "Library",
		SensorType: "temperature",
		Unit:       "celsius",
		Status:     models.SensorStatusActive,
	}
	_ = memStore.UpsertSensor(ctx, sensor)

	ruleID := uuid.New()
	rule := models.AlertRule{
		ID:          ruleID,
		Name:        "High Library Temperature",
		Enabled:     true,
		ScopeType:   models.ScopeGlobal,
		Metric:      "temperature",
		RuleClass:   models.RuleClassSingleReading,
		Operator:    ">",
		Threshold:   32.0,
		Severity:    models.SeverityCritical,
		CreatedAt:   time.Now(),
	}
	_ = memStore.CreateAlertRule(ctx, &rule)

	// Reading 1: Value 35.0 (exceeds threshold 32.0) -> Should create 1 Alert
	r1 := &models.Reading{
		SensorID:   "sensor-lib-01",
		ObservedAt: time.Now(),
		MetricType: "temperature",
		Value:      35.0,
	}

	alerts1, err := evaluator.EvaluateSingleReading(ctx, r1, sensor, []models.AlertRule{rule})
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	if len(alerts1) != 1 {
		t.Fatalf("expected 1 alert triggered, got %d", len(alerts1))
	}
	if alerts1[0].Status != models.AlertStatusOpen {
		t.Fatalf("expected alert status OPEN, got %s", alerts1[0].Status)
	}

	// Reading 2: Value 36.0 (still exceeding threshold) -> Should deduplicate and NOT create another alert!
	r2 := &models.Reading{
		SensorID:   "sensor-lib-01",
		ObservedAt: time.Now(),
		MetricType: "temperature",
		Value:      36.0,
	}
	alerts2, err := evaluator.EvaluateSingleReading(ctx, r2, sensor, []models.AlertRule{rule})
	if err != nil {
		t.Fatalf("evaluation 2 failed: %v", err)
	}
	if len(alerts2) != 0 {
		t.Fatalf("expected 0 new alerts due to active deduplication, got %d", len(alerts2))
	}

	// Now resolve the incident
	_, err = memStore.ResolveAlert(ctx, alerts1[0].ID, "organizer-1", "Air conditioning repaired")
	if err != nil {
		t.Fatalf("failed to resolve alert: %v", err)
	}

	// Reading 3: Value 37.0 (exceeds threshold after resolution) -> Should now create a new alert!
	r3 := &models.Reading{
		SensorID:   "sensor-lib-01",
		ObservedAt: time.Now(),
		MetricType: "temperature",
		Value:      37.0,
	}
	alerts3, err := evaluator.EvaluateSingleReading(ctx, r3, sensor, []models.AlertRule{rule})
	if err != nil {
		t.Fatalf("evaluation 3 failed: %v", err)
	}
	if len(alerts3) != 1 {
		t.Fatalf("expected 1 new alert after resolution, got %d", len(alerts3))
	}
}
