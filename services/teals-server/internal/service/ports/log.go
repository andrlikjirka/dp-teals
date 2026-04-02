package ports

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// AuditLogRepository defines the interface for storing audit log entries.
type AuditLogRepository interface {
	StoreAuditLogEntry(ctx context.Context, eventId uuid.UUID, payload json.RawMessage) error
}
