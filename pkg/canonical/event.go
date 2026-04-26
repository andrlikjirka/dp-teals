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

// ProtectedAuditEventPayload is the canonical form of an audit event with protected PII fields. It includes the same core fields as AuditEventPayload, but instead of a generic Metadata map, it has a structured ProtectedMetadata field that contains the encrypted PII data and associated cryptographic metadata for decryption and integrity verification.
type ProtectedAuditEventPayload struct {
	ID                string                    `json:"id"`
	Timestamp         string                    `json:"timestamp"`
	Environment       *EnvironmentPayload       `json:"environment,omitempty"`
	Actor             ActorPayload              `json:"actor"`
	Subject           SubjectPayload            `json:"subject"`
	Action            string                    `json:"action"`
	Resource          ResourcePayload           `json:"resource"`
	Result            ResultPayload             `json:"result"`
	ProtectedMetadata *ProtectedMetadataPayload `json:"protected_metadata,omitempty"`
}

// ProtectedMetadataPayload is the canonical form of the protected PII metadata. It contains the ciphertext, wrapped data encryption key (DEK), and commitment for integrity verification.
type ProtectedMetadataPayload struct {
	Ciphertext string `json:"ciphertext"`
	WrappedDEK string `json:"wrapped_dek"`
	Commitment string `json:"commitment"`
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
