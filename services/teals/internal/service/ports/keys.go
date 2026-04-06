package ports

import (
	"context"
	"crypto/ed25519"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
)

// ProducerKeyRegistry defines the interface for managing producer keys, including adding, retrieving, and revoking public keys. Implementations of this interface are responsible for the underlying storage and retrieval logic, allowing the KeyService to interact with the key data without needing to know the details of how it is stored.
type ProducerKeyRegistry interface {
	AddPublicKey(ctx context.Context, key *model.ProducerKey) error
	PublicKey(ctx context.Context, kid string) (ed25519.PublicKey, error)
	RevokeKey(ctx context.Context, kid string) error
	GetProducerKeyByKid(ctx context.Context, kid string) (*model.ProducerKey, error)
}
