package audit

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	id          uuid.UUID
	timestamp   time.Time
	environment *Environment
	actor       Actor
	subject     Subject
	action      string
	resource    Resource
	result      string
	metadata    map[string]any
}

// Environment is the service and trace context where the activity occurred.
type Environment struct {
	Service string
	TraceID string
	SpanID  string
}

// Actor is who performed the action (user or system).
type Actor struct {
	Type string // "user" | "system"
	ID   string
}

// Subject is the data subject whose personal data was processed.
type Subject struct {
	ID string
}

// Resource is the affected resource (record, dataset, or other).
type Resource struct {
	ID   uuid.UUID
	Type string // "record" | "dataset" | "other"
	Name string
}
