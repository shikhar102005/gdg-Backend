package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/shikhar102005/gdg-Backend/internal/alerts"
	"github.com/shikhar102005/gdg-Backend/internal/api"
	"github.com/shikhar102005/gdg-Backend/internal/auth"
	"github.com/shikhar102005/gdg-Backend/internal/config"
	"github.com/shikhar102005/gdg-Backend/internal/database"
	"github.com/shikhar102005/gdg-Backend/internal/ingestion"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/simulator"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

func main() {
	seedFlag := flag.Bool("seed", false, "Seed demo data on startup")
	simRateFlag := flag.Int("sim-rate", 5, "Initial simulator rate (readings/sec). Set 0 to disable auto-sim")
	flag.Parse()

	cfg := config.Load()

	log.Println("=========================================================")
	log.Println("  Campus Sensor Monitoring & Alerts Engine (GDG Backend)  ")
	log.Println("=========================================================")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var appStore store.Store

	if cfg.DatabaseURL != "" {
		log.Printf("[database] Attempting connection to PostgreSQL/Supabase...")
		pgStore, err := store.NewPostgresStore(ctx, cfg)
		if err != nil {
			log.Printf("[database] Warning: Could not connect to PostgreSQL (%v). Falling back to high-fidelity In-Memory Store for demo.", err)
			appStore = store.NewMemoryStore()
		} else {
			log.Println("[database] Successfully connected to PostgreSQL!")
			appStore = pgStore

			// Run database migrations
			if err := database.RunMigrations(ctx, pgStore.Pool(), "migrations"); err != nil {
				log.Printf("[database] Migration note: %v", err)
			}
		}
	} else {
		log.Println("[database] DATABASE_URL not set. Running with In-Memory Store (Zero-dependency demo mode).")
		appStore = store.NewMemoryStore()
	}
	defer appStore.Close()

	// Seed demo data if flag is set or store is empty
	if *seedFlag || isStoreEmpty(ctx, appStore) {
		log.Println("[seed] Seeding initial campus sensors and alert rules...")
		seedDemoData(ctx, appStore)
	}

	authSvc := auth.NewAuthService(
		cfg.OrganizerJWTSecret,
		cfg.SensorKeySalt,
		cfg.OrganizerAdminEmail,
		cfg.OrganizerAdminPassword,
		cfg.OrganizerViewerEmail,
		cfg.OrganizerViewerPassword,
	)

	evaluator := alerts.NewEvaluator(appStore)
	windowEvaluator := alerts.NewWindowEvaluator(appStore)

	ingestSvc := ingestion.NewIngestionService(
		appStore,
		evaluator,
		cfg.IngestionBatchMaxSize,
		cfg.IngestionQueueBuffer,
		cfg.IngestionWorkers,
	)
	defer ingestSvc.Close()

	sim := simulator.NewSimulator(appStore, ingestSvc)
	if *simRateFlag > 0 {
		log.Printf("[simulator] Starting background sensor simulator at %d readings/sec...", *simRateFlag)
		sim.Start(*simRateFlag)
	}

	// Background ticker for rolling time-window alert evaluation
	go func() {
		ticker := time.NewTicker(cfg.AlertWindowEvaluationInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				evalCtx, evalCancel := context.WithTimeout(ctx, 10*time.Second)
				_, _ = windowEvaluator.EvaluateTimeWindowRules(evalCtx)
				evalCancel()
			}
		}
	}()

	router := api.NewRouter(&api.ServerComponents{
		Config:    cfg,
		Store:     appStore,
		AuthSvc:   authSvc,
		IngestSvc: ingestSvc,
		Simulator: sim,
	})

	serverAddr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
		IdleTimeout:  cfg.ServerIdleTimeout,
	}

	// Graceful shutdown channel
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[http] Server listening on http://%s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[http] Listen error: %v", err)
		}
	}()

	<-stopChan
	log.Println("[http] Graceful shutdown initiated...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	sim.Stop()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[http] Server shutdown error: %v", err)
	}
	log.Println("[http] Server stopped cleanly.")
}

func isStoreEmpty(ctx context.Context, s store.Store) bool {
	sensors, err := s.ListSensors(ctx, "", "", "")
	return err != nil || len(sensors) == 0
}

func seedDemoData(ctx context.Context, s store.Store) {
	sensors := []models.Sensor{
		{
			ID: "lib-fl2-temp", Name: "Library 2nd Floor Temp", Campus: "Vellore Campus",
			Building: "Central Library", Location: "Floor 2 - Reading Hall", SensorType: "temperature", Unit: "celsius", Status: models.SensorStatusActive,
		},
		{
			ID: "lib-fl1-co2", Name: "Library Ground Floor CO2", Campus: "Vellore Campus",
			Building: "Central Library", Location: "Ground Floor - Study Pods", SensorType: "co2", Unit: "ppm", Status: models.SensorStatusActive,
		},
		{
			ID: "tt-lab3-temp", Name: "Tech Tower Lab 304 Temp", Campus: "Vellore Campus",
			Building: "Technology Tower", Location: "Room 304 - Cloud Computing Lab", SensorType: "temperature", Unit: "celsius", Status: models.SensorStatusActive,
		},
		{
			ID: "tt-lab3-hum", Name: "Tech Tower Lab 304 Humidity", Campus: "Vellore Campus",
			Building: "Technology Tower", Location: "Room 304 - Cloud Computing Lab", SensorType: "humidity", Unit: "percent", Status: models.SensorStatusActive,
		},
		{
			ID: "sjt-aud-occ", Name: "Silver Jubilee Tower Auditorium Occupancy", Campus: "Vellore Campus",
			Building: "Silver Jubilee Tower", Location: "Ground Floor - Main Hall", SensorType: "occupancy", Unit: "count", Status: models.SensorStatusActive,
		},
		{
			ID: "sjt-fl5-noise", Name: "Silver Jubilee Tower Floor 5 Noise", Campus: "Vellore Campus",
			Building: "Silver Jubilee Tower", Location: "Floor 5 - Study Lounge", SensorType: "noise", Unit: "dB", Status: models.SensorStatusActive,
		},
		{
			ID: "mb-srv-temp", Name: "Main Building Server Room Temp", Campus: "Vellore Campus",
			Building: "Main Building", Location: "Basement 1 - Rack A", SensorType: "temperature", Unit: "celsius", Status: models.SensorStatusActive,
		},
		{
			ID: "mb-air-pm25", Name: "Main Building Courtyard PM2.5", Campus: "Vellore Campus",
			Building: "Main Building", Location: "Outdoor Courtyard", SensorType: "pm25", Unit: "ug/m3", Status: models.SensorStatusActive,
		},
	}

	for _, sen := range sensors {
		_ = s.UpsertSensor(ctx, &sen)
	}

	// Seed alert rules
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
			Description:           "Alerts when rolling 10-minute average CO2 level exceeds 800 ppm",
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
		{
			ID:          uuid.New(),
			Name:        "Auditorium Overcrowding",
			Description: "Occupancy exceeding maximum recommended venue threshold",
			Enabled:     true,
			ScopeType:   models.ScopeSensor,
			ScopeValue:  "sjt-aud-occ",
			Metric:      "occupancy",
			RuleClass:   models.RuleClassSingleReading,
			Operator:    ">=",
			Threshold:   40.0,
			Severity:    models.SeverityWarning,
		},
	}

	for _, rule := range rules {
		_ = s.CreateAlertRule(ctx, &rule)
	}

	log.Printf("[seed] Finished seeding %d sensors and %d alert rules.", len(sensors), len(rules))
}
