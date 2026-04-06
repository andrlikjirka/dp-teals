package service

import (
	"context"
	"crypto/ed25519"
	"errors"
	"time"

	"github.com/andrlikjirka/dp-teals/pkg/jws"
	"github.com/andrlikjirka/dp-teals/pkg/logger"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/google/uuid"
)

// KeyService provides functionality for managing producer keys, including registration and retrieval.
type KeyService struct {
	registry ports.KeyRegistry
	logger   *logger.Logger
}

// NewKeyService creates a new instance of KeyService with the provided KeyRegistry and Logger.
func NewKeyService(r ports.KeyRegistry, l *logger.Logger) *KeyService {
	return &KeyService{
		registry: r,
		logger:   l,
	}
}

// RegisterProducerKey registers a new public key for a producer. It validates the key, computes its thumbprint (kid), and stores it in the registry. If the key is invalid or already exists, it returns an appropriate error.
func (s *KeyService) RegisterProducerKey(ctx context.Context, producerID uuid.UUID, key []byte) (string, error) {
	if len(key) != ed25519.PublicKeySize {
		return "", svcerrors.ErrInvalidPublicKey
	}
	pub := ed25519.PublicKey(key) // Ensure the key is of the correct type

	kid, err := jws.Thumbprint(pub)
	if err != nil {
		s.logger.Error("failed to compute key thumbprint (kid)", "error", err)
		return "", svcerrors.ErrKeyRegistrationFailed
	}

	pk := &model.ProducerKey{
		ID:         uuid.New(),
		ProducerID: producerID,
		KeyID:      kid,
		PublicKey:  pub,
		Status:     model.KeyStatusActive,
		CreatedAt:  time.Now().UTC(),
	}
	err = s.registry.AddPublicKey(ctx, pk)
	if err != nil {
		if errors.Is(err, svcerrors.ErrDuplicateProducerKey) {
			s.logger.Warn("duplicate key rejected", "kid", kid, "producer_id", producerID)
			return "", svcerrors.ErrDuplicateProducerKey
		}
		if errors.Is(err, svcerrors.ErrProducerNotFound) {
			s.logger.Warn("key registration rejected: producer not found", "producer_id", producerID)
			return "", svcerrors.ErrProducerNotFound
		}
		s.logger.Error("failed to store public_key", "kid", kid, "producer_id", producerID, "error", err)
		return "", svcerrors.ErrKeyRegistrationFailed
	}

	s.logger.Info("producer public_key registered", "kid", kid, "producer_id", producerID)
	return kid, nil
}
