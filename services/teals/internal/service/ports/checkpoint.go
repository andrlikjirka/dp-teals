package ports

import (
	"context"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
)

// CheckpointStore defines the interface for storing and retrieving ledger checkpoints. It abstracts the underlying storage mechanism, allowing for different implementations (e.g., database, file system) without affecting the service logic.
type CheckpointStore interface {
	// StoreCheckpoint persists a signed ledger checkpoint.
	StoreCheckpoint(ctx context.Context, checkpoint *model.SignedCheckpoint) error
	// GetLatestSignedCheckpoint returns the most recently anchored checkpoint.
	GetLatestSignedCheckpoint(ctx context.Context) (*model.SignedCheckpoint, error)
}

// CheckpointSigner defines the interface for signing checkpoint data. It abstracts the signing mechanism, allowing for different implementations without affecting the service logic.
type CheckpointSigner interface {
	Sign(payload []byte) (string, error)
	Kid() string
	PublicKey() []byte
}
