package ingestion

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shikhar102005/gdg-Backend/internal/alerts"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

func TestValidateReading(t *testing.T) {
	valid := &models.IngestReadingRequest{
		SensorID:  "sensor-1",
		Type:      "temperature",
		Value:     24.5,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	_, err := ValidateReading(valid)
	if err != nil {
		t.Fatalf("expected valid, got %v", err)
	}

	emptyID := &models.IngestReadingRequest{
		SensorID:  "",
		Type:      "temperature",
		Value:     24.5,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := ValidateReading(emptyID); err != ErrEmptySensorID {
		t.Fatalf("expected ErrEmptySensorID, got %v", err)
	}

	invalidTime := &models.IngestReadingRequest{
		SensorID:  "sensor-1",
		Type:      "temperature",
		Value:     24.5,
		Timestamp: "invalid-date",
	}
	if _, err := ValidateReading(invalidTime); err != ErrInvalidTimestamp {
		t.Fatalf("expected ErrInvalidTimestamp, got %v", err)
	}
}

func TestIngestionService_SingleAndBatch(t *testing.T) {
	ctx := context.Background()
	memStore := store.NewMemoryStore()
	evaluator := alerts.NewEvaluator(memStore)
	ingestSvc := NewIngestionService(memStore, evaluator, 5000, 1000, 2)
	defer ingestSvc.Close()

	// Single reading
	singleReq := &models.IngestReadingRequest{
		SensorID:  "sensor-test-single",
		Type:      "temperature",
		Value:     22.3,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	r, err := ingestSvc.IngestSingle(ctx, singleReq)
	if err != nil {
		t.Fatalf("failed single ingest: %v", err)
	}
	if r.SensorID != "sensor-test-single" || r.Value != 22.3 {
		t.Fatalf("unexpected reading output: %+v", r)
	}

	// Batch reading
	batchReq := &models.BatchIngestRequest{
		Readings: make([]models.IngestReadingRequest, 100),
	}
	for i := 0; i < 100; i++ {
		batchReq.Readings[i] = models.IngestReadingRequest{
			SensorID:  fmt.Sprintf("sensor-batch-%d", i%5),
			Type:      "humidity",
			Value:     float64(40 + (i % 20)),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}

	resp, err := ingestSvc.IngestBatch(ctx, batchReq)
	if err != nil {
		t.Fatalf("failed batch ingest: %v", err)
	}
	if resp.AcceptedCount != 100 {
		t.Fatalf("expected 100 accepted, got %d", resp.AcceptedCount)
	}

	count, _ := memStore.GetTotalReadingsCount(ctx)
	if count != 101 {
		t.Fatalf("expected 101 total readings in store, got %d", count)
	}
}
