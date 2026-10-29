package ingestion

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/shikhar102005/gdg-Backend/internal/models"
)

var (
	ErrEmptySensorID     = errors.New("sensorId cannot be empty")
	ErrSensorIDTooLong   = errors.New("sensorId cannot exceed 64 characters")
	ErrEmptyMetricType   = errors.New("metric type cannot be empty")
	ErrInvalidValue      = errors.New("reading value must be a valid finite number")
	ErrInvalidTimestamp  = errors.New("timestamp must be RFC3339 formatted")
	ErrFutureTimestamp   = errors.New("timestamp cannot be more than 24 hours in the future")
	ErrPastTimestamp     = errors.New("timestamp cannot be older than 30 days")
	ErrBatchTooLarge     = errors.New("batch exceeds maximum allowed size")
)

// ValidateReading validates an incoming reading request payload
func ValidateReading(req *models.IngestReadingRequest) (time.Time, error) {
	req.SensorID = strings.TrimSpace(req.SensorID)
	if req.SensorID == "" {
		return time.Time{}, ErrEmptySensorID
	}
	if len(req.SensorID) > 64 {
		return time.Time{}, ErrSensorIDTooLong
	}

	req.Type = strings.TrimSpace(req.Type)
	if req.Type == "" {
		return time.Time{}, ErrEmptyMetricType
	}

	if math.IsNaN(req.Value) || math.IsInf(req.Value, 0) {
		return time.Time{}, ErrInvalidValue
	}

	var observedAt time.Time
	if strings.TrimSpace(req.Timestamp) == "" {
		observedAt = time.Now().UTC()
	} else {
		var err error
		observedAt, err = time.Parse(time.RFC3339, req.Timestamp)
		if err != nil {
			// Try fallback format
			observedAt, err = time.Parse("2006-01-02T15:04:05Z07:00", req.Timestamp)
			if err != nil {
				return time.Time{}, ErrInvalidTimestamp
			}
		}
	}

	now := time.Now().UTC()
	if observedAt.After(now.Add(24 * time.Hour)) {
		return time.Time{}, ErrFutureTimestamp
	}
	if observedAt.Before(now.Add(-30 * 24 * time.Hour)) {
		return time.Time{}, ErrPastTimestamp
	}

	if req.Quality == "" {
		req.Quality = "good"
	}

	return observedAt, nil
}
