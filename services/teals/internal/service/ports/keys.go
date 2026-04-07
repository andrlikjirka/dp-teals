package ports

import (
	"context"
	"crypto/ed25519"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
)

// ProducerKeyRegistry defines the interface for managing producer keys, including adding, retrieving, and revoking public keys. Implementations of this interface are responsible for the underlying storage and retrieval logic, allowing the KeyService to interact with the key data without needing to know the details of how it is stored.
type ProducerKeyRegistry interface {
	// AddPublicKey adds a new producer key to the registry. It takes a context for managing request-scoped values and cancellation, and a ProducerKey model containing the key details. The method returns an error if the operation fails, allowing the caller to handle any issues that arise during the key addition process.
	AddPublicKey(ctx context.Context, key *model.ProducerKey) error
	// PublicKey retrieves the public key associated with the given key ID (kid). It takes a context for managing request-scoped values and cancellation, and the key ID as input. The method returns the corresponding ed25519.PublicKey if found, or an error if the key cannot be retrieved or does not exist in the registry.
	PublicKey(ctx context.Context, kid string) (ed25519.PublicKey, error)
	// RevokeKey revokes the producer key associated with the given key ID (kid). It takes a context for managing request-scoped values and cancellation, and the key ID as input. The method returns an error if the revocation process fails, allowing the caller to handle any issues that arise during the key revocation.
	RevokeKey(ctx context.Context, kid string) error
	// GetProducerKeyByKid retrieves the ProducerKey object associated with the given key ID (kid). It takes a context for managing request-scoped values and cancellation, and the key ID as input. The method returns the corresponding ProducerKey if found, or an error if the key cannot be retrieved or does not exist in the registry.
	GetProducerKeyByKid(ctx context.Context, kid string) (*model.ProducerKey, error)
}
