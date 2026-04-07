package ports

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// AuditLog defines the interface for storing audit log entries.
type AuditLog interface {
	// StoreAuditLogEntry stores an audit log entry with the given event ID, payload, signature token, and producer key ID. It returns an error if the operation fails.
	StoreAuditLogEntry(ctx context.Context, eventId uuid.UUID, payload json.RawMessage, sigToken string, producerKeyId uuid.UUID) error
}
