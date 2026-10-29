package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

type SensorHandler struct {
	store store.Store
}

func NewSensorHandler(s store.Store) *SensorHandler {
	return &SensorHandler{store: s}
}

func (h *SensorHandler) List(w http.ResponseWriter, r *http.Request) {
	building := r.URL.Query().Get("building")
	sensorType := r.URL.Query().Get("sensor_type")
	status := r.URL.Query().Get("status")

	sensors, err := h.store.ListSensors(r.Context(), building, sensorType, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sensors": sensors,
		"count":   len(sensors),
	})
}

func (h *SensorHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sensor, err := h.store.GetSensor(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "sensor not found: "+id)
		return
	}
	writeJSON(w, http.StatusOK, sensor)
}

func (h *SensorHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	var since *time.Time
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = &t
		}
	}

	readings, err := h.store.GetSensorReadings(r.Context(), id, limit, since)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sensor_id": id,
		"readings":  readings,
		"count":     len(readings),
	})
}

func (h *SensorHandler) CreateOrUpdate(w http.ResponseWriter, r *http.Request) {
	var sen models.Sensor
	if err := json.NewDecoder(r.Body).Decode(&sen); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if sen.ID == "" {
		writeError(w, http.StatusBadRequest, "sensor id is required")
		return
	}
	if sen.Name == "" {
		sen.Name = "Sensor " + sen.ID
	}
	if sen.Campus == "" {
		sen.Campus = "Main Campus"
	}
	if sen.Building == "" {
		sen.Building = "Academic Block"
	}
	if sen.Location == "" {
		sen.Location = "Floor 1"
	}
	if sen.Status == "" {
		sen.Status = models.SensorStatusActive
	}

	if err := h.store.UpsertSensor(r.Context(), &sen); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	claims := GetUserClaims(r)
	actor := "anonymous"
	if claims != nil {
		actor = claims.Email
	}

	_ = h.store.InsertAuditLog(r.Context(), &models.AuditLog{
		EntityType: "sensor",
		EntityID:   sen.ID,
		Action:     models.AuditCreate,
		Actor:      actor,
		Details: map[string]interface{}{
			"name":        sen.Name,
			"sensor_type": sen.SensorType,
			"building":    sen.Building,
		},
		CreatedAt: time.Now(),
	})

	writeJSON(w, http.StatusOK, sen)
}
