package serializer

import (
	"time"

	pkgcanon "github.com/andrlikjirka/dp-teals/pkg/canonicalizer"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
)

// JcsSerializer implements the Serializer interface using JSON Canonicalization Scheme (JCS).
type JcsSerializer struct{}

// NewJcsSerializer creates a new instance of a Serializer.
func NewJcsSerializer() ports.Serializer {
	return &JcsSerializer{}
}

// SerializeCanonicalAuditEvent maps a service model AuditEvent to the shared canonical DTO and delegates serialization to pkg/canonicalizer.
func (js *JcsSerializer) SerializeCanonicalAuditEvent(event *model.AuditEvent) ([]byte, error) {
	payload := toPayload(event)
	return pkgcanon.Canonicalize(payload)
}

// toPayload converts an AuditEvent to a payload object suitable for canonization.
func toPayload(event *model.AuditEvent) *pkgcanon.AuditEventPayload {
	dto := &pkgcanon.AuditEventPayload{
		ID:        event.ID.String(),
		Timestamp: event.Timestamp.UTC().Format(time.RFC3339Nano),
		Actor: pkgcanon.ActorPayload{
			Type: string(event.Actor.Type),
			ID:   event.Actor.ID,
		},
		Subject: pkgcanon.SubjectPayload{
			ID: event.Subject.ID,
		},
		Action: string(event.Action),
		Resource: pkgcanon.ResourcePayload{
			ID:     event.Resource.ID,
			Name:   event.Resource.Name,
			Fields: event.Resource.Fields,
		},
		Result: pkgcanon.ResultPayload{
			Status: string(event.Result.Status),
			Reason: event.Result.Reason,
		},
		Metadata: event.Metadata,
	}

	if event.Environment != nil {
		dto.Environment = &pkgcanon.EnvironmentPayload{
			Service: event.Environment.Service,
			TraceID: event.Environment.TraceID,
			SpanID:  event.Environment.SpanID,
		}
	}

	return dto
}
