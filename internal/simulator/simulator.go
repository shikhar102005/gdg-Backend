package simulator

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/shikhar102005/gdg-Backend/internal/ingestion"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

type Simulator struct {
	mu            sync.RWMutex
	store         store.Store
	ingestSvc     *ingestion.IngestionService
	running       bool
	ratePerSec    int
	triggerSpike  bool
	spikeMetric   string
	spikeSensorID string
	spikeValue    float64
	totalEmitted  int64
	lastEmittedAt *time.Time
	stopChan      chan struct{}
}

func NewSimulator(s store.Store, ingestSvc *ingestion.IngestionService) *Simulator {
	return &Simulator{
		store:      s,
		ingestSvc:  ingestSvc,
		ratePerSec: 5,
	}
}

func (sim *Simulator) Start(rate int) {
	sim.mu.Lock()
	if sim.running {
		sim.mu.Unlock()
		return
	}
	if rate <= 0 {
		rate = 5
	}
	sim.ratePerSec = rate
	sim.running = true
	sim.stopChan = make(chan struct{})
	sim.mu.Unlock()

	go sim.loop()
}

func (sim *Simulator) Stop() {
	sim.mu.Lock()
	defer sim.mu.Unlock()
	if !sim.running {
		return
	}
	sim.running = false
	close(sim.stopChan)
}

func (sim *Simulator) TriggerSpike(metric, sensorID string, value float64) {
	sim.mu.Lock()
	defer sim.mu.Unlock()
	sim.triggerSpike = true
	sim.spikeMetric = metric
	sim.spikeSensorID = sensorID
	sim.spikeValue = value
}

func (sim *Simulator) GetStatus() models.SimulatorConfig {
	sim.mu.RLock()
	defer sim.mu.RUnlock()

	sensors, _ := sim.store.ListSensors(context.Background(), "", "", "")

	return models.SimulatorConfig{
		Running:       sim.running,
		RatePerSecond: sim.ratePerSec,
		TriggerSpike:  sim.triggerSpike,
		SpikeMetric:   sim.spikeMetric,
		SpikeSensorID: sim.spikeSensorID,
		SpikeValue:    sim.spikeValue,
		ActiveSensors: len(sensors),
		TotalEmitted:  sim.totalEmitted,
		LastEmittedAt: sim.lastEmittedAt,
	}
}

func (sim *Simulator) loop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		select {
		case <-sim.stopChan:
			return
		case <-ticker.C:
			sim.mu.RLock()
			rate := sim.ratePerSec
			hasSpike := sim.triggerSpike
			spikeMetric := sim.spikeMetric
			spikeSensor := sim.spikeSensorID
			spikeVal := sim.spikeValue
			sim.mu.RUnlock()

			sensors, err := sim.store.ListSensors(context.Background(), "", "", "")
			if err != nil || len(sensors) == 0 {
				continue
			}

			batch := make([]models.IngestReadingRequest, 0, rate)
			now := time.Now().UTC()

			if hasSpike {
				// Inject spike reading
				targetSensor := sensors[0].ID
				if spikeSensor != "" {
					targetSensor = spikeSensor
				}
				targetMetric := "temperature"
				if spikeMetric != "" {
					targetMetric = spikeMetric
				}
				batch = append(batch, models.IngestReadingRequest{
					SensorID:  targetSensor,
					Type:      targetMetric,
					Value:     spikeVal,
					Timestamp: now.Format(time.RFC3339),
					Quality:   "spike_simulated",
				})

				sim.mu.Lock()
				sim.triggerSpike = false // Reset spike trigger after emitting
				sim.mu.Unlock()
			}

			// Generate normal readings for random sensors
			for i := len(batch); i < rate; i++ {
				sen := sensors[rng.Intn(len(sensors))]
				val := generateMetricValue(sen.SensorType, rng)

				batch = append(batch, models.IngestReadingRequest{
					SensorID:  sen.ID,
					Type:      sen.SensorType,
					Value:     val,
					Timestamp: now.Format(time.RFC3339),
					Quality:   "simulated",
				})
			}

			resp, err := sim.ingestSvc.IngestBatch(context.Background(), &models.BatchIngestRequest{Readings: batch})
			if err == nil && resp != nil {
				sim.mu.Lock()
				sim.totalEmitted += int64(resp.AcceptedCount)
				sim.lastEmittedAt = &now
				sim.mu.Unlock()
			}
		}
	}
}

func generateMetricValue(sensorType string, rng *rand.Rand) float64 {
	switch sensorType {
	case "temperature":
		// Normal between 21.0 and 26.0 °C
		return 21.0 + rng.Float64()*5.0
	case "humidity":
		// Normal between 40% and 55%
		return 40.0 + rng.Float64()*15.0
	case "co2":
		// Normal between 400 and 750 ppm
		return 400.0 + rng.Float64()*350.0
	case "occupancy":
		// Room occupancy count 0 to 45
		return float64(rng.Intn(45))
	case "noise":
		// Noise level 35 to 65 dB
		return 35.0 + rng.Float64()*30.0
	case "pm25":
		// PM2.5 5 to 25 ug/m3
		return 5.0 + rng.Float64()*20.0
	default:
		return 10.0 + rng.Float64()*20.0
	}
}
