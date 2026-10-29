package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shikhar102005/gdg-Backend/internal/auth"
	"github.com/shikhar102005/gdg-Backend/internal/config"
	"github.com/shikhar102005/gdg-Backend/internal/ingestion"
	"github.com/shikhar102005/gdg-Backend/internal/simulator"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

type ServerComponents struct {
	Config    *config.Config
	Store     store.Store
	AuthSvc   *auth.AuthService
	IngestSvc *ingestion.IngestionService
	Simulator *simulator.Simulator
}

func NewRouter(c *ServerComponents) http.Handler {
	r := chi.NewRouter()

	// Base middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(LoggingMiddleware)
	r.Use(CORSMiddleware(c.Config.CORSAllowedOrigins))

	healthHandler := NewHealthHandler(c.Store)
	authHandler := NewAuthHandler(c.AuthSvc)
	ingestHandler := NewIngestHandler(c.IngestSvc)
	sensorHandler := NewSensorHandler(c.Store)
	readingsHandler := NewReadingsHandler(c.Store)
	rulesHandler := NewRulesHandler(c.Store)
	alertHandler := NewAlertHandler(c.Store)
	dashboardHandler := NewDashboardHandler(c.Store)
	simHandler := NewSimulatorHandler(c.Simulator)

	// Liveness & readiness probes
	r.Get("/healthz", healthHandler.Healthz)
	r.Get("/readyz", healthHandler.Readyz)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes
		r.Post("/auth/login", authHandler.Login)
		r.With(RequireAuth(c.AuthSvc)).Get("/auth/me", authHandler.Me)

		// Sensor Ingestion (Burst optimized)
		r.Post("/readings", ingestHandler.IngestSingle)
		r.Post("/readings/batch", ingestHandler.IngestBatch)

		// Trends & analytics
		r.Get("/readings/trends", readingsHandler.GetTrends)

		// Sensors inventory & history
		r.Get("/sensors", sensorHandler.List)
		r.Get("/sensors/{id}", sensorHandler.Get)
		r.Get("/sensors/{id}/history", sensorHandler.GetHistory)
		r.Post("/sensors", sensorHandler.CreateOrUpdate)

		// Dashboard summary & audit
		r.Get("/dashboard/summary", dashboardHandler.GetSummary)
		r.Get("/audit-logs", dashboardHandler.ListAuditLogs)

		// Alert rules
		r.Get("/rules", rulesHandler.List)
		r.Get("/rules/{id}", rulesHandler.Get)
		r.Post("/rules", rulesHandler.Create)
		r.Put("/rules/{id}", rulesHandler.Update)
		r.Patch("/rules/{id}/toggle", rulesHandler.Toggle)

		// Alerts lifecycle
		r.Get("/alerts", alertHandler.List)
		r.Get("/alerts/{id}", alertHandler.Get)
		r.Post("/alerts/{id}/acknowledge", alertHandler.Acknowledge)
		r.Post("/alerts/{id}/resolve", alertHandler.Resolve)
		r.Post("/alerts/{id}/reopen", alertHandler.Reopen)

		// Simulator controls
		r.Get("/simulator/status", simHandler.GetStatus)
		r.Post("/simulator/start", simHandler.Start)
		r.Post("/simulator/stop", simHandler.Stop)
		r.Post("/simulator/trigger-spike", simHandler.TriggerSpike)
	})

	return r
}
