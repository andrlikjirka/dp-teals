package model

import (
	"time"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model/enum"
	"github.com/google/uuid"
)

// baseEvent holds the fields common to both AuditEvent and ProtectedAuditEvent.
type baseEvent struct {
	ID          uuid.UUID
	Timestamp   time.Time
	Environment *Environment
	Actor       Actor
	Subject     Subject
	Action      enum.ActionType
	Resource    Resource
	Result      Result
}

// AuditEvent represents an audit event with all relevant details.
type AuditEvent struct {
	baseEvent
	Metadata map[string]any
}

// Environment is the service and trace context where the activity occurred.
type Environment struct {
	Service string
	TraceID string
	SpanID  string
}

// Actor is who performed the action (user or system).
type Actor struct {
	Type enum.ActorType
	ID   string
}

// Subject is the data subject whose personal data was processed.
type Subject struct {
	ID string
}

// Resource is the affected resource (record, dataset, or other).
type Resource struct {
	ID     string
	Name   string
	Fields []string
}

// Result captures the outcome of the action.
type Result struct {
	Status enum.ResultStatusType
	Reason string
}

// BaseEventParams holds the parameters common to both factory methods.
type BaseEventParams struct {
	ID          uuid.UUID
	Timestamp   time.Time
	Environment *Environment
	Actor       Actor
	Subject     Subject
	Action      enum.ActionType
	Resource    Resource
	Result      Result
}

// CreateAuditEventParams encapsulates the parameters needed to create an AuditEvent.
type CreateAuditEventParams struct {
	BaseEventParams
	Metadata map[string]any
}

// validateBaseEvent checks the common fields for validity and returns an error if any required fields are missing or invalid.
func validateBaseEvent(p BaseEventParams) error {
	if p.ID == uuid.Nil {
		return errors.ErrInvalidEventID
	}
	if p.Timestamp.IsZero() {
		return errors.ErrMissingTimestamp
	}
	if !p.Actor.Type.IsValid() {
		return errors.ErrInvalidActorType
	}
	if p.Actor.ID == "" {
		return errors.ErrMissingActorID
	}
	if p.Subject.ID == "" {
		return errors.ErrMissingSubjectID
	}
	if !p.Action.IsValid() {
		return errors.ErrInvalidActionType
	}
	if p.Resource.ID == "" || p.Resource.Name == "" {
		return errors.ErrInvalidResource
	}
	if !p.Result.Status.IsValid() {
		return errors.ErrInvalidResultStatus
	}
	return nil
}

// NewAuditEvent validates the input parameters and constructs a new AuditEvent.
func NewAuditEvent(params CreateAuditEventParams) (*AuditEvent, error) {
	if err := validateBaseEvent(params.BaseEventParams); err != nil {
		return nil, err
	}

	return &AuditEvent{
		baseEvent: baseEvent{
			ID:          params.ID,
			Timestamp:   params.Timestamp,
			Environment: params.Environment,
			Actor:       params.Actor,
			Subject:     params.Subject,
			Action:      params.Action,
			Resource:    params.Resource,
			Result:      params.Result,
		},
		Metadata: params.Metadata,
	}, nil
}

// IngestAuditEventResult represents the result of ingesting an audit event, including the assigned EventID, updated ledger size, and ingestion timestamp.
type IngestAuditEventResult struct {
	EventID    uuid.UUID
	LedgerSize int64
	IngestedAt time.Time
}
