package api

import (
	"encoding/json"
	"net/http"

	"github.com/shikhar102005/gdg-Backend/internal/ingestion"
	"github.com/shikhar102005/gdg-Backend/internal/models"
)

type IngestHandler struct {
	ingestSvc *ingestion.IngestionService
}

func NewIngestHandler(ingestSvc *ingestion.IngestionService) *IngestHandler {
	return &IngestHandler{ingestSvc: ingestSvc}
}

// IngestSingle handles POST /api/v1/readings
func (h *IngestHandler) IngestSingle(w http.ResponseWriter, r *http.Request) {
	var req models.IngestReadingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload: "+err.Error())
		return
	}

	reading, err := h.ingestSvc.IngestSingle(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":  "accepted",
		"reading": reading,
	})
}

// IngestBatch handles POST /api/v1/readings/batch
func (h *IngestHandler) IngestBatch(w http.ResponseWriter, r *http.Request) {
	var req models.BatchIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid batch payload: "+err.Error())
		return
	}

	resp, err := h.ingestSvc.IngestBatch(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	status := http.StatusOK
	if resp.FailedCount > 0 && resp.AcceptedCount == 0 {
		status = http.StatusBadRequest
	} else if resp.FailedCount > 0 {
		status = http.StatusMultiStatus
	}

	writeJSON(w, status, resp)
}
