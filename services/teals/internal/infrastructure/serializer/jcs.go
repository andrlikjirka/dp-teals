package serializer

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	pkgcanon "github.com/andrlikjirka/dp-teals/pkg/canonical"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model/enum"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/google/uuid"
)

// JcsSerializer implements the Serializer interface using JSON Canonicalization Scheme (JCS).
type JcsSerializer struct{}

// NewJcsSerializer creates a new instance of a Serializer.
func NewJcsSerializer() ports.Serializer {
	return &JcsSerializer{}
}

// SerializeCanonicalAuditEvent maps a service model AuditEvent to the shared canonical DTO and delegates serialization to pkg/canonicalizer.
func (js *JcsSerializer) SerializeCanonicalAuditEvent(event *model.AuditEvent) (json.RawMessage, error) {
	payload := toPayload(event)
	return pkgcanon.CanonicalizeAuditEvent(payload)
}

// DeserializeCanonicalAuditEvent takes a byte slice representing a canonicalized audit event and converts it back into an AuditEvent struct. It returns the deserialized AuditEvent or an error if deserialization fails.
func (js *JcsSerializer) DeserializeCanonicalAuditEvent(payload json.RawMessage) (*model.AuditEvent, error) {
	var p pkgcanon.AuditEventPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("unmarshal canonical audit event: %w", err)
	}
	return fromPayload(&p)
}

// SerializeCanonicalProtectedAuditEvent maps a service model ProtectedAuditEvent to the shared canonical protected DTO and delegates serialization to pkg/canonicalizer, encoding the protected metadata fields as base64 strings.
func (js *JcsSerializer) SerializeCanonicalProtectedAuditEvent(event *model.ProtectedAuditEvent) (json.RawMessage, error) {
	payload := toProtectedPayload(event)
	return pkgcanon.CanonicalizeProtectedAuditEvent(payload)
}

// DeserializeCanonicalProtectedAuditEvent takes a byte slice representing a canonicalized protected audit event and converts it back into a ProtectedAuditEvent struct. It returns the deserialized ProtectedAuditEvent or an error if deserialization fails.
func (js *JcsSerializer) DeserializeCanonicalProtectedAuditEvent(payload json.RawMessage) (*model.ProtectedAuditEvent, error) {
	var p pkgcanon.ProtectedAuditEventPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("unmarshal protected audit event: %w", err)
	}
	return fromProtectedPayload(&p)
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

// fromPayload converts a canonical payload back to an AuditEvent.
func fromPayload(p *pkgcanon.AuditEventPayload) (*model.AuditEvent, error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("parse event id: %w", err)
	}

	ts, err := time.Parse(time.RFC3339Nano, p.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("parse timestamp: %w", err)
	}

	var env *model.Environment
	if p.Environment != nil {
		env = &model.Environment{
			Service: p.Environment.Service,
			TraceID: p.Environment.TraceID,
			SpanID:  p.Environment.SpanID,
		}
	}

	return model.NewAuditEvent(model.CreateAuditEventParams{
		BaseEventParams: model.BaseEventParams{
			ID:          id,
			Timestamp:   ts,
			Environment: env,
			Actor: model.Actor{
				Type: enum.ActorType(p.Actor.Type),
				ID:   p.Actor.ID,
			},
			Subject: model.Subject{ID: p.Subject.ID},
			Action:  enum.ActionType(p.Action),
			Resource: model.Resource{
				ID:     p.Resource.ID,
				Name:   p.Resource.Name,
				Fields: p.Resource.Fields,
			},
			Result: model.Result{
				Status: enum.ResultStatusType(p.Result.Status),
				Reason: p.Result.Reason,
			},
		},
		Metadata: p.Metadata,
	})
}

// toProtectedPayload converts a ProtectedAuditEvent to a payload object suitable for canonization, encoding the protected metadata fields as base64 strings.
func toProtectedPayload(event *model.ProtectedAuditEvent) *pkgcanon.ProtectedAuditEventPayload {
	dto := &pkgcanon.ProtectedAuditEventPayload{
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
	}

	if event.Environment != nil {
		dto.Environment = &pkgcanon.EnvironmentPayload{
			Service: event.Environment.Service,
			TraceID: event.Environment.TraceID,
			SpanID:  event.Environment.SpanID,
		}
	}

	if event.ProtectedMetadata != nil {
		dto.ProtectedMetadata = &pkgcanon.ProtectedMetadataPayload{
			Ciphertext: base64.StdEncoding.EncodeToString(event.ProtectedMetadata.Ciphertext),
			WrappedDEK: base64.StdEncoding.EncodeToString(event.ProtectedMetadata.WrappedDEK),
			Commitment: hex.EncodeToString(event.ProtectedMetadata.Commitment),
		}
	}

	return dto
}

// fromProtectedPayload converts a canonical protected payload back to a ProtectedAuditEvent.
func fromProtectedPayload(p *pkgcanon.ProtectedAuditEventPayload) (*model.ProtectedAuditEvent, error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, fmt.Errorf("parse event id: %w", err)
	}

	ts, err := time.Parse(time.RFC3339Nano, p.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("parse timestamp: %w", err)
	}

	var env *model.Environment
	if p.Environment != nil {
		env = &model.Environment{
			Service: p.Environment.Service,
			TraceID: p.Environment.TraceID,
			SpanID:  p.Environment.SpanID,
		}
	}

	var pm *model.ProtectedMetadata
	if p.ProtectedMetadata != nil {
		ciphertext, err := base64.StdEncoding.DecodeString(p.ProtectedMetadata.Ciphertext)
		if err != nil {
			return nil, fmt.Errorf("decode ciphertext: %w", err)
		}
		wrappedDEK, err := base64.StdEncoding.DecodeString(p.ProtectedMetadata.WrappedDEK)
		if err != nil {
			return nil, fmt.Errorf("decode wrapped dek: %w", err)
		}
		commitment, err := hex.DecodeString(p.ProtectedMetadata.Commitment)
		if err != nil {
			return nil, fmt.Errorf("decode commitment: %w", err)
		}
		pm = &model.ProtectedMetadata{
			Ciphertext: ciphertext,
			WrappedDEK: wrappedDEK,
			Commitment: commitment,
		}
	}

	return model.NewProtectedAuditEvent(model.CreateProtectedAuditEventParams{
		BaseEventParams: model.BaseEventParams{
			ID:          id,
			Timestamp:   ts,
			Environment: env,
			Actor: model.Actor{
				Type: enum.ActorType(p.Actor.Type),
				ID:   p.Actor.ID,
			},
			Subject: model.Subject{ID: p.Subject.ID},
			Action:  enum.ActionType(p.Action),
			Resource: model.Resource{
				ID:     p.Resource.ID,
				Name:   p.Resource.Name,
				Fields: p.Resource.Fields},
			Result: model.Result{
				Status: enum.ResultStatusType(p.Result.Status),
				Reason: p.Result.Reason,
			},
		},
		ProtectedMetadata: pm,
	})
}
