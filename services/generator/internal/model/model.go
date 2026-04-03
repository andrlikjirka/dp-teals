package model

import (
	"time"

	ingestionv1 "github.com/andrlijirka/dp-teals/gen/audit/v1"
)

type AuditEvent struct {
	ID          string
	Timestamp   time.Time
	Environment *Environment
	Actor       Actor
	Subject     Subject
	Action      ingestionv1.Action // proto enum — no internal equivalent needed
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
	Type ingestionv1.Actor_Type // proto enum — used directly
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
	Status ingestionv1.Result_Status // proto enum — used directly
	Reason string
}
