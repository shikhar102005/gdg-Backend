package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shikhar102005/gdg-Backend/internal/alerts"
	"github.com/shikhar102005/gdg-Backend/internal/auth"
	"github.com/shikhar102005/gdg-Backend/internal/config"
	"github.com/shikhar102005/gdg-Backend/internal/ingestion"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/simulator"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

func setupTestServer() (http.Handler, *auth.AuthService, store.Store) {
	cfg := config.Load()
	memStore := store.NewMemoryStore()
	authSvc := auth.NewAuthService(cfg.OrganizerJWTSecret, cfg.SensorKeySalt, cfg.OrganizerAdminEmail, cfg.OrganizerAdminPassword, cfg.OrganizerViewerEmail, cfg.OrganizerViewerPassword)
	evaluator := alerts.NewEvaluator(memStore)
	ingestSvc := ingestion.NewIngestionService(memStore, evaluator, 5000, 1000, 2)
	sim := simulator.NewSimulator(memStore, ingestSvc)

	router := NewRouter(&ServerComponents{
		Config:    cfg,
		Store:     memStore,
		AuthSvc:   authSvc,
		IngestSvc: ingestSvc,
		Simulator: sim,
	})

	return router, authSvc, memStore
}

func TestHealthEndpoints(t *testing.T) {
	router, _, _ := setupTestServer()

	// GET /healthz
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /healthz, got %d", rr.Code)
	}

	// GET /readyz
	req, _ = http.NewRequest(http.MethodGet, "/readyz", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /readyz, got %d", rr.Code)
	}
}

func TestAuthLogin(t *testing.T) {
	router, _, _ := setupTestServer()

	// Valid login
	body, _ := json.Marshal(map[string]string{
		"email":    "organizer@campus.edu",
		"password": "CampusOrganizerSecure2026!",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for valid login, got %d", rr.Code)
	}

	var resp LoginResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Token == "" || resp.Role != auth.RoleAdmin {
		t.Fatalf("unexpected login response: %+v", resp)
	}

	// Invalid login
	body, _ = json.Marshal(map[string]string{
		"email":    "organizer@campus.edu",
		"password": "wrongpassword",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for bad login, got %d", rr.Code)
	}
}

func TestIngestBatchEndpoint(t *testing.T) {
	router, _, _ := setupTestServer()

	batch := models.BatchIngestRequest{
		Readings: []models.IngestReadingRequest{
			{
				SensorID:  "sensor-bench-1",
				Type:      "temperature",
				Value:     23.5,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
			{
				SensorID:  "sensor-bench-2",
				Type:      "humidity",
				Value:     48.0,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	body, _ := json.Marshal(batch)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/readings/batch", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for batch ingest, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.BatchIngestResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp.AcceptedCount != 2 {
		t.Fatalf("expected 2 accepted readings, got %d", resp.AcceptedCount)
	}
}

func TestAlertRulesAndLifecycle(t *testing.T) {
	router, _, _ := setupTestServer()

	// Create rule
	rule := models.AlertRule{
		Name:      "Server Room Overheat",
		Metric:    "temperature",
		Operator:  ">",
		Threshold: 30.0,
		Severity:  models.SeverityCritical,
		RuleClass: models.RuleClassSingleReading,
		Enabled:   true,
	}
	body, _ := json.Marshal(rule)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/rules", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for rule creation, got %d: %s", rr.Code, rr.Body.String())
	}

	var createdRule models.AlertRule
	_ = json.NewDecoder(rr.Body).Decode(&createdRule)

	// Ingest reading exceeding threshold -> triggers alert
	reading := models.IngestReadingRequest{
		SensorID:  "sensor-server-room",
		Type:      "temperature",
		Value:     36.5,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	body, _ = json.Marshal(reading)
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/readings", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted for single reading, got %d", rr.Code)
	}

	// Give alert worker a few milliseconds to process
	time.Sleep(50 * time.Millisecond)

	// List alerts
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/alerts", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var alertListResp struct {
		Alerts []models.Alert `json:"alerts"`
		Count  int            `json:"count"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&alertListResp)

	if alertListResp.Count == 0 {
		t.Fatalf("expected triggered alert to be found, got 0")
	}

	alertID := alertListResp.Alerts[0].ID.String()

	// Acknowledge alert
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/alerts/"+alertID+"/acknowledge", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("failed acknowledging alert: %d", rr.Code)
	}

	// Resolve alert
	resolveBody, _ := json.Marshal(map[string]string{"notes": "Server HVAC restarted"})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/alerts/"+alertID+"/resolve", bytes.NewBuffer(resolveBody))
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("failed resolving alert: %d", rr.Code)
	}

	// Check dashboard summary
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("failed getting dashboard summary: %d", rr.Code)
	}
}
