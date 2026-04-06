package model

import (
	"crypto/ed25519"
	"time"

	"github.com/google/uuid"
)

type KeyStatus string

const (
	KeyStatusActive  KeyStatus = "active"
	KeyStatusRevoked KeyStatus = "revoked"
)

type ProducerKey struct {
	ID         uuid.UUID
	ProducerID uuid.UUID
	KeyId      string
	PublicKey  ed25519.PublicKey
	Status     KeyStatus
	CreatedAt  time.Time
}
