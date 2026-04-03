package model

import (
	"time"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model/enum"
	"github.com/google/uuid"
)

// AuditEvent represents an audit event with all relevant details.
type AuditEvent struct {
	ID          uuid.UUID
	Timestamp   time.Time
	Environment *Environment
	Actor       Actor
	Subject     Subject
	Action      enum.ActionType
	Resource    Resource
	Result      Result
	Metadata    map[string]any
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

// CreateAuditEventParams encapsulates the parameters needed to create an AuditEvent.
type CreateAuditEventParams struct {
	ID          uuid.UUID
	Timestamp   time.Time
	Environment *Environment // Pointer because it's optional
	Actor       Actor
	Subject     Subject
	Action      enum.ActionType
	Resource    Resource
	Result      Result
	Metadata    map[string]any
}

// NewAuditEvent validates the input parameters and constructs a new AuditEvent.
func NewAuditEvent(params CreateAuditEventParams) (*AuditEvent, error) {
	if params.ID == uuid.Nil {
		return nil, errors.ErrInvalidEventID
	}
	if params.Timestamp.IsZero() {
		return nil, errors.ErrMissingTimestamp
	}

	if !params.Actor.Type.IsValid() {
		return nil, errors.ErrInvalidActorType
	}
	if params.Actor.ID == "" {
		return nil, errors.ErrMissingActorID
	}

	if params.Subject.ID == "" {
		return nil, errors.ErrMissingSubjectID
	}

	if !params.Action.IsValid() {
		return nil, errors.ErrInvalidActionType
	}

	if params.Resource.ID == "" || params.Resource.Name == "" {
		return nil, errors.ErrInvalidResource
	}

	if !params.Result.Status.IsValid() {
		return nil, errors.ErrInvalidResultStatus
	}

	return &AuditEvent{
		ID:          params.ID,
		Timestamp:   params.Timestamp,
		Environment: params.Environment,
		Actor:       params.Actor,
		Subject:     params.Subject,
		Action:      params.Action,
		Resource:    params.Resource,
		Result:      params.Result,
		Metadata:    params.Metadata,
	}, nil
}
