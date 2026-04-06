package model

import (
	"time"

	"github.com/google/uuid"
)

// ProducerKeyRecord defines the database schema for the producer key record.
type ProducerKeyRecord struct {
	ID         uuid.UUID `db:"id"`
	ProducerID uuid.UUID `db:"producer_id"`
	KeyID      string    `db:"kid"`
	PublicKey  []byte    `db:"public_key"`
	Status     string    `db:"status"`
	CreatedAt  time.Time `db:"created_at"`
}
