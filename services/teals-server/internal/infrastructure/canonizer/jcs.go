package canonizer

import (
	"encoding/json"
	"time"

	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service/model"
	"github.com/andrlijirka/dp-teals/services/teals-server/internal/service/ports"
	"github.com/gowebpki/jcs"
)

// JcsSerializer implements the Serializer interface using JSON Canonicalization Scheme (JCS).
type JcsSerializer struct{}

// NewJcsSerializer creates a new instance of a Serializer.
func NewJcsSerializer() ports.Serializer {
	return &JcsSerializer{}
}

// SerializeCanonicalAuditEvent serializes an AuditEvent into a canonical JSON form using JCS.
func (js *JcsSerializer) SerializeCanonicalAuditEvent(event *model.AuditEvent) ([]byte, error) {
	dto := toPayload(event)
	jsonData, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	return jcs.Transform(jsonData)
}

// toPayload converts an AuditEvent to a payload object suitable for canonization.
func toPayload(event *model.AuditEvent) *auditEventPayload {
	dto := &auditEventPayload{
		ID:        event.ID.String(),
		Timestamp: event.Timestamp.UTC().Format(time.RFC3339Nano),
		Actor: actorPayload{
			Type: string(event.Actor.Type),
			ID:   event.Actor.ID,
		},
		Subject: subjectPayload{
			ID: event.Subject.ID,
		},
		Action: string(event.Action),
		Resource: resourcePayload{
			ID:     event.Resource.ID,
			Name:   event.Resource.Name,
			Fields: event.Resource.Fields,
		},
		Result: resultPayload{
			Status: string(event.Result.Status),
			Reason: event.Result.Reason,
		},
		Metadata: event.Metadata,
	}

	if event.Environment != nil {
		dto.Environment = &environmentPayload{
			Service: event.Environment.Service,
			TraceID: event.Environment.TraceID,
			SpanID:  event.Environment.SpanID,
		}
	}

	return dto
}

// auditEventPayload is the data transfer object for an AuditEvent.
type auditEventPayload struct {
	ID          string              `json:"id"`
	Timestamp   string              `json:"timestamp"`
	Environment *environmentPayload `json:"environment,omitempty"`
	Actor       actorPayload        `json:"actor"`
	Subject     subjectPayload      `json:"subject"`
	Action      string              `json:"action"`
	Resource    resourcePayload     `json:"resource"`
	Result      resultPayload       `json:"result"`
	Metadata    map[string]any      `json:"metadata,omitempty"`
}

// environmentPayload is the DTO for Environment.
type environmentPayload struct {
	Service string `json:"service"`
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
}

// actorPayload is the DTO for Actor.
type actorPayload struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// subjectPayload is the DTO for Subject.
type subjectPayload struct {
	ID string `json:"id"`
}

// resourcePayload is the DTO for Resource.
type resourcePayload struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Fields []string `json:"fields,omitempty"`
}

// resultPayload is the DTO for Result.
type resultPayload struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}
