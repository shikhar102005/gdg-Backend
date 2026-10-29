package api

import (
	"context"
	"net/http"
	"time"

	"github.com/shikhar102005/gdg-Backend/internal/store"
)

type HealthHandler struct {
	store store.Store
}

func NewHealthHandler(s store.Store) *HealthHandler {
	return &HealthHandler{store: s}
}

// Healthz indicates process liveness
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "campus-sensor-backend",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Readyz indicates database connectivity and operational readiness
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.store.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":    "unready",
			"error":     err.Error(),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ready",
		"database":  "connected",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
