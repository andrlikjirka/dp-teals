package model

import (
	"time"

	"github.com/google/uuid"
)

type AuditEvent struct {
	ID          uuid.UUID
	Timestamp   time.Time
	Environment *Environment
	Actor       Actor
	Subject     Subject
	Action      ActionType
	Resource    Resource
	Result      Result
	Metadata    map[string]any
}

type Environment struct {
	Service string
	TraceID string
	SpanID  string
}

type Actor struct {
	Type ActorType
	ID   string
}

type Subject struct {
	ID string
}

type Resource struct {
	ID     string
	Name   string
	Fields []string
}

type Result struct {
	Status ResultStatus
	Reason string
}

// ActionType represents the type of action performed on a resource.
type ActionType string

const (
	ActionAccess ActionType = "ACCESS"
	ActionCreate ActionType = "CREATE"
	ActionUpdate ActionType = "UPDATE"
	ActionDelete ActionType = "DELETE"
	ActionShare  ActionType = "SHARE"
	ActionExport ActionType = "EXPORT"
	ActionLogin  ActionType = "LOGIN"
	ActionLogout ActionType = "LOGOUT"
)

// ActorType represents whether the actor is a human user or an automated system.
type ActorType string

const (
	ActorTypeUser   ActorType = "USER"
	ActorTypeSystem ActorType = "SYSTEM"
)

// ResultStatus represents the outcome of the action.
type ResultStatus string

const (
	ResultStatusSuccess ResultStatus = "SUCCESS"
	ResultStatusFailure ResultStatus = "FAILURE"
)
