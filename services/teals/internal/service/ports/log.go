package ports

import (
	"context"
	"encoding/json"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/google/uuid"
)

// AuditLog defines the interface for storing audit log entries.
type AuditLog interface {
	// StoreAuditLogEntry stores an audit log entry with the given event ID, payload, signature token, producer key ID, and node ID.
	StoreAuditLogEntry(ctx context.Context, eventId uuid.UUID, payload json.RawMessage, sigToken string, producerKeyId uuid.UUID, nodeID int64) error
	// GetAuditLogEntryByEventID retrieves an audit log entry by its event ID. It returns the audit log entry if found, or an error if the entry is not found or if there was an error during retrieval.
	GetAuditLogEntryByEventID(ctx context.Context, eventID uuid.UUID) (*model.AuditLogEntryRaw, error)
	// ListAuditLogEntries retrieves a list of audit log entries based on the provided filter and cursor. It returns a slice of audit log entries that match the filter criteria, or an error if there was an issue during retrieval.
	ListAuditLogEntries(ctx context.Context, filter *model.AuditEventFilter, cursor *int64, size int) ([]*model.AuditLogEntryRaw, error)
}
