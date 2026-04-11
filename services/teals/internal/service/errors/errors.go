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

	ErrInvalidSignature = errors.New("invalid event signature")

	ErrDuplicateProducerKey       = errors.New("duplicate producer key")
	ErrKeyRegistrationFailed      = errors.New("failed to register producer key")
	ErrInvalidPublicKey           = errors.New("invalid public key")
	ErrProducerKeyNotFound        = errors.New("producer key not found")
	ErrProducerNotFound           = errors.New("producer not found")
	ErrKeyNotFound                = errors.New("producer key not found")
	ErrProducerKeyRetrievalFailed = errors.New("failed to retrieve producer key by kid")

	ErrLedgerAppendFailed = errors.New("failed to append audit event to ledger")
	ErrEmptyLeafData      = errors.New("empty leaf data not allowed")
	ErrInsertNodeFailed   = errors.New("failed to insert node into ledger")

	ErrLedgerSizeFailed                = errors.New("failed to get ledger size")
	ErrAuditLogEntryNotFound           = errors.New("audit log entry not found")
	ErrInclusionProofFailed            = errors.New("failed to generate inclusion proof")
	ErrRootHashFailed                  = errors.New("failed to get ledger root hash")
	ErrInvalidConsistencyProofRange    = errors.New("invalid consistency proof range: from_size must be less than to_size and both must be less than or equal to the current ledger size")
	ErrConsistencyProofFailed          = errors.New("failed to generate consistency proof")
	ErrInvalidInclusionProofLedgerSize = errors.New("invalid inclusion proof ledger size: tree_size must be gte leaf position and lte current ledger size")
)
