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
)
