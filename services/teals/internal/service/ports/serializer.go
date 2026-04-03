package ports

import (
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
)

// Serializer serializes data into a canonical form.
type Serializer interface {
	SerializeCanonicalAuditEvent(event *model.AuditEvent) ([]byte, error)
}
