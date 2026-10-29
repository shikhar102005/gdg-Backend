package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/shikhar102005/gdg-Backend/internal/config"
	"github.com/shikhar102005/gdg-Backend/internal/database"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

func main() {
	readingsCountFlag := flag.Int("readings", 200, "Number of historic readings to generate per sensor")
	resetFlag := flag.Bool("reset", false, "Reset tables before seeding")
	flag.Parse()

	cfg := config.Load()

	ctx := context.Background()
	log.Println("[seed] Connecting to store...")

	var appStore store.Store
	if cfg.DatabaseURL != "" {
		pgStore, err := store.NewPostgresStore(ctx, cfg)
		if err != nil {
			log.Fatalf("[seed] Could not connect to PostgreSQL: %v", err)
		}
		appStore = pgStore
		defer appStore.Close()

		if *resetFlag {
			log.Println("[seed] Resetting schema...")
			// Apply migrations fresh
			_ = database.RunMigrations(ctx, pgStore.Pool(), "migrations")
		} else {
			_ = database.RunMigrations(ctx, pgStore.Pool(), "migrations")
		}
	} else {
		log.Println("[seed] Note: DATABASE_URL not set. Running against memory store (for verification).")
		appStore = store.NewMemoryStore()
		defer appStore.Close()
	}

	sensors := []models.Sensor{
		{ID: "lib-fl2-temp", Name: "Library 2nd Floor Temp", Campus: "Vellore Campus", Building: "Central Library", Location: "Floor 2 - Reading Hall", SensorType: "temperature", Unit: "celsius", Status: models.SensorStatusActive},
		{ID: "lib-fl1-co2", Name: "Library Ground Floor CO2", Campus: "Vellore Campus", Building: "Central Library", Location: "Ground Floor - Study Pods", SensorType: "co2", Unit: "ppm", Status: models.SensorStatusActive},
		{ID: "tt-lab3-temp", Name: "Tech Tower Lab 304 Temp", Campus: "Vellore Campus", Building: "Technology Tower", Location: "Room 304 - Cloud Computing Lab", SensorType: "temperature", Unit: "celsius", Status: models.SensorStatusActive},
		{ID: "tt-lab3-hum", Name: "Tech Tower Lab 304 Humidity", Campus: "Vellore Campus", Building: "Technology Tower", Location: "Room 304 - Cloud Computing Lab", SensorType: "humidity", Unit: "percent", Status: models.SensorStatusActive},
		{ID: "sjt-aud-occ", Name: "Silver Jubilee Tower Auditorium Occupancy", Campus: "Vellore Campus", Building: "Silver Jubilee Tower", Location: "Ground Floor - Main Hall", SensorType: "occupancy", Unit: "count", Status: models.SensorStatusActive},
		{ID: "sjt-fl5-noise", Name: "Silver Jubilee Tower Floor 5 Noise", Campus: "Vellore Campus", Building: "Silver Jubilee Tower", Location: "Floor 5 - Study Lounge", SensorType: "noise", Unit: "dB", Status: models.SensorStatusActive},
		{ID: "mb-srv-temp", Name: "Main Building Server Room Temp", Campus: "Vellore Campus", Building: "Main Building", Location: "Basement 1 - Rack A", SensorType: "temperature", Unit: "celsius", Status: models.SensorStatusActive},
		{ID: "mb-air-pm25", Name: "Main Building Courtyard PM2.5", Campus: "Vellore Campus", Building: "Main Building", Location: "Outdoor Courtyard", SensorType: "pm25", Unit: "ug/m3", Status: models.SensorStatusActive},
	}

	for _, sen := range sensors {
		if err := appStore.UpsertSensor(ctx, &sen); err != nil {
			log.Fatalf("failed inserting sensor %s: %v", sen.ID, err)
		}
	}
	log.Printf("[seed] Inserted %d sensors.", len(sensors))

	rules := []models.AlertRule{
		{
			ID:          uuid.New(),
			Name:        "Critical Server Room Temperature",
			Description: "Immediate alert when server room temperature exceeds safe thermal limits",
			Enabled:     true,
			ScopeType:   models.ScopeSensor,
			ScopeValue:  "mb-srv-temp",
			Metric:      "temperature",
			RuleClass:   models.RuleClassSingleReading,
			Operator:    ">",
			Threshold:   28.0,
			Severity:    models.SeverityCritical,
		},
		{
			ID:          uuid.New(),
			Name:        "High Ambient Library Temperature",
			Description: "Alerts when study hall temperature rises above comfortable levels",
			Enabled:     true,
			ScopeType:   models.ScopeBuilding,
			ScopeValue:  "Central Library",
			Metric:      "temperature",
			RuleClass:   models.RuleClassSingleReading,
			Operator:    ">",
			Threshold:   29.0,
			Severity:    models.SeverityWarning,
		},
		{
			ID:                    uuid.New(),
			Name:                  "High Indoor CO2 Level (10m Average)",
			Description:           "Alerts when rolling 10-minute average CO2 level exceeds 750 ppm",
			Enabled:               true,
			ScopeType:             models.ScopeGlobal,
			Metric:                "co2",
			RuleClass:             models.RuleClassTimeWindow,
			Operator:              ">",
			Threshold:             750.0,
			WindowDurationSeconds: 600,
			WindowAggregation:     models.AggAvg,
			Severity:              models.SeverityWarning,
		},
	}

	for _, rule := range rules {
		if err := appStore.CreateAlertRule(ctx, &rule); err != nil {
			log.Fatalf("failed inserting rule %s: %v", rule.Name, err)
		}
	}
	log.Printf("[seed] Inserted %d alert rules.", len(rules))

	// Generate historical readings
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	totalReadings := 0
	batch := make([]models.Reading, 0, 500)

	for _, sen := range sensors {
		for i := 0; i < *readingsCountFlag; i++ {
			observedAt := time.Now().UTC().Add(-time.Duration(*readingsCountFlag-i) * 2 * time.Minute)
			val := 22.0 + rng.Float64()*4.0
			if sen.SensorType == "co2" {
				val = 450.0 + rng.Float64()*200.0
			} else if sen.SensorType == "humidity" {
				val = 45.0 + rng.Float64()*10.0
			} else if sen.SensorType == "occupancy" {
				val = float64(rng.Intn(40))
			} else if sen.SensorType == "noise" {
				val = 40.0 + rng.Float64()*20.0
			} else if sen.SensorType == "pm25" {
				val = 12.0 + rng.Float64()*10.0
			}

			batch = append(batch, models.Reading{
				SensorID:       sen.ID,
				ObservedAt:     observedAt,
				ReceivedAt:     observedAt.Add(50 * time.Millisecond),
				MetricType:     sen.SensorType,
				Value:          val,
				Quality:        "good",
				IdempotencyKey: fmt.Sprintf("seed-%s-%d", sen.ID, i),
			})

			if len(batch) >= 500 {
				n, _ := appStore.InsertReadingBatch(ctx, batch)
				totalReadings += n
				batch = batch[:0]
			}
		}
	}

	if len(batch) > 0 {
		n, _ := appStore.InsertReadingBatch(ctx, batch)
		totalReadings += n
	}

	log.Printf("[seed] Inserted %d historical readings.", totalReadings)
	log.Println("[seed] Seeding complete!")
}
