package api

import (
	"encoding/json"
	"net/http"

	"github.com/shikhar102005/gdg-Backend/internal/simulator"
)

type SimulatorHandler struct {
	sim *simulator.Simulator
}

func NewSimulatorHandler(sim *simulator.Simulator) *SimulatorHandler {
	return &SimulatorHandler{sim: sim}
}

func (h *SimulatorHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.sim.GetStatus())
}

func (h *SimulatorHandler) Start(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		RatePerSecond int `json:"rate_per_second"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	rate := payload.RatePerSecond
	if rate <= 0 {
		rate = 10
	}
	h.sim.Start(rate)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "started",
		"rate_per_second": rate,
	})
}

func (h *SimulatorHandler) Stop(w http.ResponseWriter, r *http.Request) {
	h.sim.Stop()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "stopped",
	})
}

func (h *SimulatorHandler) TriggerSpike(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Metric   string  `json:"metric"`
		SensorID string  `json:"sensor_id"`
		Value    float64 `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if payload.Metric == "" {
		payload.Metric = "temperature"
	}
	if payload.Value == 0 {
		payload.Value = 38.5 // Trigger critical heat alert
	}

	h.sim.TriggerSpike(payload.Metric, payload.SensorID, payload.Value)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "spike_queued",
		"metric":    payload.Metric,
		"sensor_id": payload.SensorID,
		"value":     payload.Value,
	})
}
