package ports

import (
	"context"
	"crypto/ed25519"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
)

// KeyRegistry defines the interface for managing producer keys. It is used by the JWS signing and verification components to access the necessary public keys for their operations.
type KeyRegistry interface {
	AddPublicKey(ctx context.Context, key *model.ProducerKey) error
	PublicKey(ctx context.Context, kid string) (ed25519.PublicKey, error)
	RevokeKey(ctx context.Context, kid string) error
}
