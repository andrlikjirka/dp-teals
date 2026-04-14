package model

import (
	"time"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model/enum"
	"github.com/google/uuid"
)

// AuditEventFilter defines the criteria for filtering audit events when querying.
type AuditEventFilter struct {
	Actions        []enum.ActionType
	ActorTypes     []enum.ActorType
	ActorID        string
	SubjectID      string
	ResourceID     string
	ResourceName   string
	ResultStatuses []enum.ResultStatusType
	TimestampFrom  *time.Time
	TimestampTo    *time.Time
	AtLedgerSize   int64
}

// GetAuditEventResult encapsulates the result of retrieving a single audit event, including the event details and associated metadata.
type GetAuditEventResult struct {
	Event          *AuditEvent
	LeafIndex      int64
	SignatureToken string
}

// ListAuditEventsResult encapsulates the result of listing audit events based on a filter, including the list of events and the ledger size at the time of retrieval.
type ListAuditEventsResult struct {
	Items      []*AuditEventListItem
	LedgerSize int64
}

// AuditEventListItem represents an individual audit event in the list of events returned by a query, including the event details and associated metadata such as the leaf index and signature token.
type AuditEventListItem struct {
	EventID        uuid.UUID
	Event          *AuditEvent
	SignatureToken string
	LeafIndex      int64
}
