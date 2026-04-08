package ports

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// AuditLog defines the interface for storing audit log entries.
type AuditLog interface {
	// StoreAuditLogEntry stores an audit log entry with the given event ID, payload, signature token, producer key ID, and node ID.
	StoreAuditLogEntry(ctx context.Context, eventId uuid.UUID, payload json.RawMessage, sigToken string, producerKeyId uuid.UUID, nodeID int64) error
}
