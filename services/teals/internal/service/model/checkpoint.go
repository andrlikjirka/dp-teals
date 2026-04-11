package model

import (
	"time"

	"github.com/google/uuid"
)

// Checkpoint represents a ledger checkpoint with size, root hash, and the time it was anchored.
type Checkpoint struct {
	Size       int64
	RootHash   []byte
	AnchoredAt time.Time
}

// SignedCheckpoint represents a checkpoint along with its signature and the key identifier used for signing.
type SignedCheckpoint struct {
	ID             uuid.UUID
	Checkpoint     Checkpoint
	Kid            string
	SignatureToken string
}
