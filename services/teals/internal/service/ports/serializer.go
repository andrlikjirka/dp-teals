package ports

import (
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
)

// Serializer serializes data into a canonical form.
type Serializer interface {
	// SerializeCanonicalAuditEvent takes an AuditEvent and converts it into a canonical byte representation. It returns the serialized data or an error if serialization fails.
	SerializeCanonicalAuditEvent(event *model.AuditEvent) ([]byte, error)
}
