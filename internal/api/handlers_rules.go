package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shikhar102005/gdg-Backend/internal/models"
	"github.com/shikhar102005/gdg-Backend/internal/store"
)

type RulesHandler struct {
	store store.Store
}

func NewRulesHandler(s store.Store) *RulesHandler {
	return &RulesHandler{store: s}
}

func (h *RulesHandler) List(w http.ResponseWriter, r *http.Request) {
	enabledOnly := r.URL.Query().Get("enabled_only") == "true"
	rules, err := h.store.ListAlertRules(r.Context(), enabledOnly)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

func (h *RulesHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rule id format")
		return
	}

	rule, err := h.store.GetAlertRule(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (h *RulesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var rule models.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if rule.Name == "" {
		writeError(w, http.StatusBadRequest, "rule name is required")
		return
	}
	if rule.Metric == "" {
		writeError(w, http.StatusBadRequest, "metric is required")
		return
	}
	if rule.Operator == "" {
		rule.Operator = ">"
	}
	if rule.Severity == "" {
		rule.Severity = models.SeverityWarning
	}
	if rule.RuleClass == "" {
		rule.RuleClass = models.RuleClassSingleReading
	}
	if rule.RuleClass == models.RuleClassTimeWindow {
		if rule.WindowDurationSeconds <= 0 {
			rule.WindowDurationSeconds = 600
		}
		if rule.WindowAggregation == "" {
			rule.WindowAggregation = models.AggAvg
		}
	}

	if err := h.store.CreateAlertRule(r.Context(), &rule); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	claims := GetUserClaims(r)
	actor := "organizer"
	if claims != nil {
		actor = claims.Email
	}

	_ = h.store.InsertAuditLog(r.Context(), &models.AuditLog{
		EntityType: "rule",
		EntityID:   rule.ID.String(),
		Action:     models.AuditCreate,
		Actor:      actor,
		Details: map[string]interface{}{
			"name":       rule.Name,
			"metric":     rule.Metric,
			"operator":   rule.Operator,
			"threshold":  rule.Threshold,
			"rule_class": rule.RuleClass,
		},
		CreatedAt: time.Now(),
	})

	writeJSON(w, http.StatusCreated, rule)
}

func (h *RulesHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rule id")
		return
	}

	var rule models.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rule.ID = id

	if err := h.store.UpdateAlertRule(r.Context(), &rule); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	claims := GetUserClaims(r)
	actor := "organizer"
	if claims != nil {
		actor = claims.Email
	}

	_ = h.store.InsertAuditLog(r.Context(), &models.AuditLog{
		EntityType: "rule",
		EntityID:   rule.ID.String(),
		Action:     models.AuditUpdate,
		Actor:      actor,
		Details: map[string]interface{}{
			"name":      rule.Name,
			"threshold": rule.Threshold,
			"enabled":   rule.Enabled,
		},
		CreatedAt: time.Now(),
	})

	writeJSON(w, http.StatusOK, rule)
}

func (h *RulesHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rule id")
		return
	}

	var payload struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.store.ToggleAlertRule(r.Context(), id, payload.Enabled); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	action := models.AuditEnable
	if !payload.Enabled {
		action = models.AuditDisable
	}

	claims := GetUserClaims(r)
	actor := "organizer"
	if claims != nil {
		actor = claims.Email
	}

	_ = h.store.InsertAuditLog(r.Context(), &models.AuditLog{
		EntityType: "rule",
		EntityID:   id.String(),
		Action:     action,
		Actor:      actor,
		Details:    map[string]interface{}{"enabled": payload.Enabled},
		CreatedAt:  time.Now(),
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"enabled": payload.Enabled,
	})
}
