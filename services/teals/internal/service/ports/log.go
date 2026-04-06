package ports

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// AuditLog defines the interface for storing audit log entries.
type AuditLog interface {
	StoreAuditLogEntry(ctx context.Context, eventId uuid.UUID, payload json.RawMessage) error
}
