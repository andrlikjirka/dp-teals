package model

import (
	"time"

	"github.com/google/uuid"
)

// CheckpointRecord defines the database schema for storing checkpoint information.
type CheckpointRecord struct {
	ID             uuid.UUID `db:"id"`
	Size           int64     `db:"size"`
	RootHash       []byte    `db:"root_hash"`
	AnchoredAt     time.Time `db:"anchored_at"`
	Kid            string    `db:"kid"`
	SignatureToken string    `db:"signature_token"`
}
