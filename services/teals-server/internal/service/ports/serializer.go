package ports

import (
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service/model"
)

// Serializer serializes data into a canonical form.
type Serializer interface {
	SerializeCanonicalAuditEvent(event *model.AuditEvent) ([]byte, error)
}
