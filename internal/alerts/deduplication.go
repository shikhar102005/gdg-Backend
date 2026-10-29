package alerts

import (
	"fmt"

	"github.com/shikhar102005/gdg-Backend/internal/models"
)

// GenerateDeduplicationKey produces a deterministic correlation key for an alert.
// This prevents alert spam while an incident is continuously active.
func GenerateDeduplicationKey(rule *models.AlertRule, sensor *models.Sensor) string {
	return fmt.Sprintf("rule:%s:sensor:%s", rule.ID.String(), sensor.ID)
}
