package model

import (
	"time"

	"github.com/google/uuid"
)

type AuditLogEntry struct {
	ID      *int64
	EventID uuid.UUID
	//LeafIndex int64
	//Signature []byte
	CreatedAt time.Time
	Payload   AuditEvent
}
