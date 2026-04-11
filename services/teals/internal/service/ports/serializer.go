package ports

import (
	"encoding/json"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
)

// Serializer serializes data into a canonical form.
type Serializer interface {
	// SerializeCanonicalAuditEvent takes an AuditEvent and converts it into a canonical byte representation. It returns the serialized data or an error if serialization fails.
	SerializeCanonicalAuditEvent(event *model.AuditEvent) (json.RawMessage, error)
	// DeserializeCanonicalAuditEvent takes a byte slice representing a canonicalized audit event and converts it back into an AuditEvent struct. It returns the deserialized AuditEvent or an error if deserialization fails.
	DeserializeCanonicalAuditEvent(data json.RawMessage) (*model.AuditEvent, error)
}
