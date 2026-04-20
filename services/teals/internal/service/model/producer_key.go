package model

import (
	"crypto/ed25519"
	"time"

	"github.com/google/uuid"
)

// KeyStatus defines the possible statuses of a producer key, indicating whether it is active or revoked. This status is crucial for determining if a key can be used for signing JWS tokens or if it should be considered invalid due to revocation.
type KeyStatus string

// KeyStatus represents the status of a producer key, indicating whether it is active and can be used for signing JWS tokens or if it has been revoked and should no longer be used.
const (
	KeyStatusActive  KeyStatus = "active"
	KeyStatusRevoked KeyStatus = "revoked"
)

// ProducerKey represents a public key associated with a producer, including its metadata and status. It is used for signing and verifying JWS tokens in the Teals system.
type ProducerKey struct {
	ID         uuid.UUID
	ProducerID uuid.UUID
	KeyID      string
	PublicKey  ed25519.PublicKey
	Status     KeyStatus
	CreatedAt  time.Time
}

// Producer represents an entity that produces audit events, identified by a unique ID and associated with one or more producer keys. It includes metadata such as the producer's name and the timestamp of when it was created.
type Producer struct {
	ID        uuid.UUID
	Name      string
	Keys      []ProducerKey
	CreatedAt time.Time
}
