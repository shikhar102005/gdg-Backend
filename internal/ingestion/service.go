package ingestion

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shikhar102005/gdg-Backend/internal/alerts"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

type ReadingWithSensor struct {
	Reading models.Reading
	Sensor  *models.Sensor
}

type IngestionService struct {
	store        store.Store
	evaluator    *alerts.Evaluator
	maxBatchSize int
	alertChan    chan ReadingWithSensor
	stopChan     chan struct{}
	wg           sync.WaitGroup
	sensorCache  sync.Map
	provisionMu  sync.Mutex
}

func NewIngestionService(s store.Store, evaluator *alerts.Evaluator, maxBatchSize int, queueBuffer int, workers int) *IngestionService {
	if maxBatchSize <= 0 {
		maxBatchSize = 5000
	}
	if queueBuffer <= 0 {
		queueBuffer = 50000
	}
	if workers <= 0 {
		workers = 4
	}

	svc := &IngestionService{
		store:        s,
		evaluator:    evaluator,
		maxBatchSize: maxBatchSize,
		alertChan:    make(chan ReadingWithSensor, queueBuffer),
		stopChan:     make(chan struct{}),
	}

	// Pre-populate sensorCache from existing sensors in store
	sensors, err := s.ListSensors(context.Background(), "", "", "")
	if err == nil {
		for i := range sensors {
			svc.sensorCache.Store(sensors[i].ID, &sensors[i])
		}
	}

	// Start background alert workers
	for i := 0; i < workers; i++ {
		svc.wg.Add(1)
		go svc.alertWorker()
	}

	return svc
}

func (s *IngestionService) Close() {
	close(s.stopChan)
	s.wg.Wait()
}

func (s *IngestionService) alertWorker() {
	defer s.wg.Done()

	// Cache rules periodically to avoid excessive DB reads during high ingestion
	var cachedRules []models.AlertRule
	var lastRulesFetch time.Time

	for {
		select {
		case <-s.stopChan:
			return
		case item, ok := <-s.alertChan:
			if !ok {
				return
			}

			if time.Since(lastRulesFetch) > 3*time.Second || cachedRules == nil {
				rules, err := s.store.ListAlertRules(context.Background(), true)
				if err == nil {
					cachedRules = rules
					lastRulesFetch = time.Now()
				}
			}

			if len(cachedRules) > 0 && item.Sensor != nil {
				_, _ = s.evaluator.EvaluateSingleReading(context.Background(), &item.Reading, item.Sensor, cachedRules)
			}
		}
	}
}

// IngestSingle processes a single reading request
func (s *IngestionService) IngestSingle(ctx context.Context, req *models.IngestReadingRequest) (*models.Reading, error) {
	observedAt, err := ValidateReading(req)
	if err != nil {
		return nil, err
	}

	sensor, err := s.ensureSensorExists(ctx, req.SensorID, req.Type)
	if err != nil {
		return nil, fmt.Errorf("sensor validation failed: %w", err)
	}

	reading := models.Reading{
		SensorID:       req.SensorID,
		ObservedAt:     observedAt,
		ReceivedAt:     time.Now().UTC(),
		MetricType:     strings.ToLower(req.Type),
		Value:          req.Value,
		Quality:        req.Quality,
		IdempotencyKey: req.IdempotencyKey,
		Metadata:       req.Metadata,
	}

	// Persist to store
	_, err = s.store.InsertReadingBatch(ctx, []models.Reading{reading})
	if err != nil {
		return nil, fmt.Errorf("persistence error: %w", err)
	}

	// Non-blocking dispatch to alert evaluation
	select {
	case s.alertChan <- ReadingWithSensor{Reading: reading, Sensor: sensor}:
	default:
		// Queue full under extreme burst; reading is already safely durable
	}

	return &reading, nil
}

// IngestBatch processes a batch of readings efficiently
func (s *IngestionService) IngestBatch(ctx context.Context, req *models.BatchIngestRequest) (*models.BatchIngestResponse, error) {
	start := time.Now()
	total := len(req.Readings)

	if total == 0 {
		return &models.BatchIngestResponse{
			TotalProcessed: 0,
			AcceptedCount:  0,
			FailedCount:    0,
			DurationMs:     time.Since(start).Milliseconds(),
		}, nil
	}

	if total > s.maxBatchSize {
		return nil, fmt.Errorf("%w: received %d, limit is %d", ErrBatchTooLarge, total, s.maxBatchSize)
	}

	validReadings := make([]models.Reading, 0, total)
	itemResults := make([]models.BatchItemResult, 0)
	sensorMap := make(map[string]*models.Sensor)

	for i, r := range req.Readings {
		observedAt, err := ValidateReading(&r)
		if err != nil {
			itemResults = append(itemResults, models.BatchItemResult{
				Index:   i,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		sen, exists := sensorMap[r.SensorID]
		if !exists {
			sen, err = s.ensureSensorExists(ctx, r.SensorID, r.Type)
			if err != nil {
				itemResults = append(itemResults, models.BatchItemResult{
					Index:   i,
					Success: false,
					Error:   err.Error(),
				})
				continue
			}
			sensorMap[r.SensorID] = sen
		}

		reading := models.Reading{
			SensorID:       r.SensorID,
			ObservedAt:     observedAt,
			ReceivedAt:     time.Now().UTC(),
			MetricType:     strings.ToLower(r.Type),
			Value:          r.Value,
			Quality:        r.Quality,
			IdempotencyKey: r.IdempotencyKey,
			Metadata:       r.Metadata,
		}
		validReadings = append(validReadings, reading)
	}

	// Bulk write valid readings
	accepted, err := s.store.InsertReadingBatch(ctx, validReadings)
	if err != nil {
		return nil, fmt.Errorf("batch insert failed: %w", err)
	}

	// Dispatch alerts
	for _, vr := range validReadings {
		sen := sensorMap[vr.SensorID]
		select {
		case s.alertChan <- ReadingWithSensor{Reading: vr, Sensor: sen}:
		default:
		}
	}

	failed := total - accepted
	return &models.BatchIngestResponse{
		TotalProcessed: total,
		AcceptedCount:  accepted,
		FailedCount:    failed,
		DurationMs:     time.Since(start).Milliseconds(),
		Errors:         itemResults,
	}, nil
}

func (s *IngestionService) ensureSensorExists(ctx context.Context, sensorID, metricType string) (*models.Sensor, error) {
	if val, ok := s.sensorCache.Load(sensorID); ok {
		return val.(*models.Sensor), nil
	}

	s.provisionMu.Lock()
	defer s.provisionMu.Unlock()

	// Double check cache after acquiring lock
	if val, ok := s.sensorCache.Load(sensorID); ok {
		return val.(*models.Sensor), nil
	}

	sen, err := s.store.GetSensor(ctx, sensorID)
	if err == nil && sen != nil {
		s.sensorCache.Store(sensorID, sen)
		return sen, nil
	}

	// Auto-provision sensor if not existing
	unit := "units"
	switch strings.ToLower(metricType) {
	case "temperature":
		unit = "celsius"
	case "humidity":
		unit = "percent"
	case "co2":
		unit = "ppm"
	case "occupancy":
		unit = "count"
	case "noise":
		unit = "dB"
	case "pm25":
		unit = "ug/m3"
	}

	newSensor := &models.Sensor{
		ID:         sensorID,
		Name:       fmt.Sprintf("Sensor %s", sensorID),
		Campus:     "Main Campus",
		Building:   "Academic Block",
		Location:   "Room Auto-Reg",
		SensorType: strings.ToLower(metricType),
		Unit:       unit,
		Status:     models.SensorStatusActive,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_ = s.store.UpsertSensor(ctx, newSensor)
	s.sensorCache.Store(sensorID, newSensor)
	return newSensor, nil
}
