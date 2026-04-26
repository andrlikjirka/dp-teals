package canonical

// AuditEventPayload is the canonical, transport-independent representation of an audit event used as the JWS signing payload.
type AuditEventPayload struct {
	ID          string              `json:"id"`
	Timestamp   string              `json:"timestamp"`
	Environment *EnvironmentPayload `json:"environment,omitempty"`
	Actor       ActorPayload        `json:"actor"`
	Subject     SubjectPayload      `json:"subject"`
	Action      string              `json:"action"`
	Resource    ResourcePayload     `json:"resource"`
	Result      ResultPayload       `json:"result"`
	Metadata    map[string]any      `json:"metadata,omitempty"`
}

// EnvironmentPayload is the canonical form of service/trace context.
type EnvironmentPayload struct {
	Service string `json:"service"`
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
}

// ActorPayload is the canonical form of the actor.
type ActorPayload struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// SubjectPayload is the canonical form of the data subject.
type SubjectPayload struct {
	ID string `json:"id"`
}

// ResourcePayload is the canonical form of the affected resource.
type ResourcePayload struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Fields []string `json:"fields,omitempty"`
}

// ResultPayload is the canonical form of the action outcome.
type ResultPayload struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}
