package model

// ProtectedAuditEvent represents an audit event with sensitive metadata protected by encryption. It contains all the standard fields of an audit event, along with an optional ProtectedMetadata field that holds the encrypted metadata if it was present in the original event.
type ProtectedAuditEvent struct {
	baseEvent
	ProtectedMetadata *ProtectedMetadata
}

// ProtectedMetadata encapsulates the encrypted PII metadata for a ProtectedAuditEvent, including the ciphertext, the wrapped data encryption key (DEK), and the commitment used for PII integrity verification.
type ProtectedMetadata struct {
	Ciphertext []byte
	WrappedDEK []byte
	Commitment []byte
}

// CreateProtectedAuditEventParams encapsulates the parameters needed to create a ProtectedAuditEvent, including all standard event fields and the optional protected metadata.
type CreateProtectedAuditEventParams struct {
	BaseEventParams
	ProtectedMetadata *ProtectedMetadata
}

// NewProtectedAuditEvent validates the input parameters and constructs a new ProtectedAuditEvent. It returns an error if any required fields are missing or invalid.
func NewProtectedAuditEvent(params CreateProtectedAuditEventParams) (*ProtectedAuditEvent, error) {
	if err := validateBaseEvent(params.BaseEventParams); err != nil {
		return nil, err
	}
	return &ProtectedAuditEvent{
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
		ProtectedMetadata: params.ProtectedMetadata,
	}, nil

}
