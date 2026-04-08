package model

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogEntry represents a complete audit log entry with all relevant details.
type AuditLogEntry struct {
	ID             *int64
	EventID        uuid.UUID
	ProducerKeyID  uuid.UUID
	SignatureToken string
	LeafIndex      int64
	CreatedAt      time.Time
	Payload        AuditEvent
}
