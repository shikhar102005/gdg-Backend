package api

import (
	"net/http"
	"time"

	"github.com/shikhar102005/gdg-Backend/internal/store"
)

type ReadingsHandler struct {
	store store.Store
}

func NewReadingsHandler(s store.Store) *ReadingsHandler {
	return &ReadingsHandler{store: s}
}

// GetTrends handles GET /api/v1/readings/trends
func (h *ReadingsHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "temperature"
	}
	sensorID := r.URL.Query().Get("sensor_id")
	building := r.URL.Query().Get("building")

	duration := 1 * time.Hour
	if dStr := r.URL.Query().Get("duration"); dStr != "" {
		if d, err := time.ParseDuration(dStr); err == nil && d > 0 {
			duration = d
		}
	}

	bucketSize := 1 * time.Minute
	if bStr := r.URL.Query().Get("bucket"); bStr != "" {
		if b, err := time.ParseDuration(bStr); err == nil && b > 0 {
			bucketSize = b
		}
	}

	points, err := h.store.GetMetricTrends(r.Context(), metric, sensorID, building, duration, bucketSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"metric":      metric,
		"sensor_id":   sensorID,
		"building":    building,
		"duration":    duration.String(),
		"bucket_size": bucketSize.String(),
		"points":      points,
		"count":       len(points),
	})
}
