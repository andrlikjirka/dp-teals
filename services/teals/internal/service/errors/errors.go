package errors

import "errors"

// Common errors for event validation and processing
var (
	ErrInvalidEventID      = errors.New("invalid event ID")
	ErrInvalidActorType    = errors.New("invalid or unsupported actor type")
	ErrMissingActorID      = errors.New("invalid actor: missing ID")
	ErrInvalidActionType   = errors.New("invalid or unsupported action type")
	ErrInvalidResultStatus = errors.New("invalid or unsupported result status")
	ErrMissingTimestamp    = errors.New("missing or invalid timestamp")
	ErrInvalidResource     = errors.New("invalid resource: missing ID or name")
	ErrMissingSubjectID    = errors.New("invalid subject: missing ID")

	ErrEventSerializationFailed = errors.New("failed to serialize audit event")
	ErrDuplicateEventID         = errors.New("duplicate event ID")
	ErrEventAppendFailed        = errors.New("failed to append audit event")

	ErrDuplicateProducerKey  = errors.New("duplicate producer key")
	ErrKeyRegistrationFailed = errors.New("failed to register producer key")
	ErrInvalidProducerID     = errors.New("invalid producer ID")
	ErrInvalidPublicKey      = errors.New("invalid public key")
	ErrProducerNotFound      = errors.New("producer not found")
)
