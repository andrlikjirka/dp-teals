package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLogEntryRecord defines the database schema for the audit event record.
type AuditLogEntryRecord struct {
	ID      int64     `db:"id"`
	EventID uuid.UUID `db:"event_id"`
	//LeafIndex int64           `db:"leaf_index"`
	//Signature []byte          `db:"signature"`
	CreatedAt time.Time       `db:"created_at"`
	Payload   json.RawMessage `db:"payload"`
}
