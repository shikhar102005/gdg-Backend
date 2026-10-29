package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

type AlertHandler struct {
	store store.Store
}

func NewAlertHandler(s store.Store) *AlertHandler {
	return &AlertHandler{store: s}
}

func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	severity := r.URL.Query().Get("severity")
	building := r.URL.Query().Get("building")
	sensorID := r.URL.Query().Get("sensor_id")

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	alerts, err := h.store.ListAlerts(r.Context(), status, severity, building, sensorID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	totalActive, crit, warn, info, _ := h.store.GetActiveAlertsCount(r.Context())

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"alerts":              alerts,
		"count":               len(alerts),
		"active_alerts_total": totalActive,
		"critical_count":      crit,
		"warning_count":       warn,
		"info_count":          info,
	})
}

func (h *AlertHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert id")
		return
	}

	alert, err := h.store.GetAlert(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "alert not found")
		return
	}
	writeJSON(w, http.StatusOK, alert)
}

func (h *AlertHandler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert id")
		return
	}

	actor := "organizer"
	if claims := GetUserClaims(r); claims != nil {
		actor = claims.Email
	}

	alert, err := h.store.AcknowledgeAlert(r.Context(), id, actor)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = h.store.InsertAuditLog(r.Context(), &models.AuditLog{
		EntityType: "alert",
		EntityID:   id.String(),
		Action:     models.AuditAcknowledge,
		Actor:      actor,
		Details:    map[string]interface{}{"status": alert.Status},
		CreatedAt:  time.Now(),
	})

	writeJSON(w, http.StatusOK, alert)
}

func (h *AlertHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert id")
		return
	}

	var payload struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	actor := "organizer"
	if claims := GetUserClaims(r); claims != nil {
		actor = claims.Email
	}

	alert, err := h.store.ResolveAlert(r.Context(), id, actor, payload.Notes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = h.store.InsertAuditLog(r.Context(), &models.AuditLog{
		EntityType: "alert",
		EntityID:   id.String(),
		Action:     models.AuditResolve,
		Actor:      actor,
		Details: map[string]interface{}{
			"status": alert.Status,
			"notes":  payload.Notes,
		},
		CreatedAt: time.Now(),
	})

	writeJSON(w, http.StatusOK, alert)
}

func (h *AlertHandler) Reopen(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid alert id")
		return
	}

	actor := "organizer"
	if claims := GetUserClaims(r); claims != nil {
		actor = claims.Email
	}

	alert, err := h.store.ReopenAlert(r.Context(), id, actor)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = h.store.InsertAuditLog(r.Context(), &models.AuditLog{
		EntityType: "alert",
		EntityID:   id.String(),
		Action:     models.AuditReopen,
		Actor:      actor,
		Details:    map[string]interface{}{"status": alert.Status},
		CreatedAt:  time.Now(),
	})

	writeJSON(w, http.StatusOK, alert)
}
