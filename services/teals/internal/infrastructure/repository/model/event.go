package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLogEntryRecord defines the database schema for the audit event record.
type AuditLogEntryRecord struct {
	ID             int64     `db:"id"`
	EventID        uuid.UUID `db:"event_id"`
	SignatureToken string    `db:"signature_token"`
	//LeafIndex int64           `db:"leaf_index"`
	CreatedAt time.Time       `db:"created_at"`
	Payload   json.RawMessage `db:"payload"`
}
