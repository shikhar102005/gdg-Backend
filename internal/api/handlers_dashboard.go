package api

import (
	"net/http"
	"strconv"

	"github.com/shikhar102005/gdg-Backend/internal/store"
)

type DashboardHandler struct {
	store store.Store
}

func NewDashboardHandler(s store.Store) *DashboardHandler {
	return &DashboardHandler{store: s}
}

// GetSummary handles GET /api/v1/dashboard/summary
func (h *DashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.store.GetDashboardSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// ListAuditLogs handles GET /api/v1/audit-logs
func (h *DashboardHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")
	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := h.store.ListAuditLogs(r.Context(), entityType, entityID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"audit_logs": logs,
		"count":      len(logs),
	})
}
