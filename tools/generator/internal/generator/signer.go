package generator

import (
	"fmt"
	"time"

	pkgcanon "github.com/andrlikjirka/dp-teals/pkg/canonicalizer"
	"github.com/andrlikjirka/dp-teals/pkg/jws"
	"github.com/andrlikjirka/dp-teals/tools/generator/internal/model"
)

// signer defines the interface for signing audit events before sending them to the ingestion service. It abstracts the signing mechanism, allowing for different implementations (e.g., using JWS or other signing methods) without affecting the rest of the event generation and sending logic.
type signer interface {
	// Sign takes a byte slice representing the payload of an audit event and returns a string containing the signature (e.g., a JWS token) or an error if the signing process fails. This method is crucial for ensuring the integrity and authenticity of the audit events being sent to the ingestion service.
	Sign(event *model.AuditEvent) (string, error)
}

// EventSigner is an implementation of the signer interface that uses a JWS signer to create signatures for audit events. It encapsulates the logic for converting audit events into a canonical format suitable for signing and delegates the actual signing process to the provided JWS signer.
type EventSigner struct {
	jwsSigner jws.Signer
}

// NewEventSigner creates a new instance of EventSigner with the provided JWS signer. This allows the generator to sign audit events using the specified JWS signing mechanism, ensuring that the events can be verified by the ingestion service upon receipt.
func NewEventSigner(s jws.Signer) *EventSigner {
	return &EventSigner{jwsSigner: s}
}

// Sign takes an audit event, converts it to a canonical payload format, and uses the JWS signer to create a signature for the event. It returns the generated signature as a string or an error if any step of the process fails, such as issues with mapping the event to a payload or problems during the signing process.
func (s *EventSigner) Sign(event *model.AuditEvent) (string, error) {
	payload, err := toPayload(event)
	if err != nil {
		return "", fmt.Errorf("sign: map to canonical payload: %w", err)
	}

	bytes, err := pkgcanon.CanonicalizeAuditEvent(payload)
	if err != nil {
		return "", fmt.Errorf("sign: canonicalize: %w", err)
	}
	return s.jwsSigner.Sign(bytes)
}

// toPayload converts an AuditEvent to a payload object suitable for JSON canonization. It maps the fields of the service model AuditEvent to the corresponding fields in the pkg/canonicalizer's AuditEventPayload structure, ensuring that all necessary information is included for accurate representation and signing of the event.
func toPayload(event *model.AuditEvent) (*pkgcanon.AuditEventPayload, error) {
	p := &pkgcanon.AuditEventPayload{
		ID:        event.ID.String(),
		Timestamp: event.Timestamp.UTC().Format(time.RFC3339Nano),
		Actor:     pkgcanon.ActorPayload{Type: string(event.Actor.Type), ID: event.Actor.ID},
		Subject:   pkgcanon.SubjectPayload{ID: event.Subject.ID},
		Action:    string(event.Action),
		Resource:  pkgcanon.ResourcePayload{ID: event.Resource.ID, Name: event.Resource.Name, Fields: event.Resource.Fields},
		Result:    pkgcanon.ResultPayload{Status: string(event.Result.Status), Reason: event.Result.Reason},
		Metadata:  event.Metadata,
	}
	if event.Environment != nil {
		p.Environment = &pkgcanon.EnvironmentPayload{
			Service: event.Environment.Service,
			TraceID: event.Environment.TraceID,
			SpanID:  event.Environment.SpanID,
		}
	}
	return p, nil
}
