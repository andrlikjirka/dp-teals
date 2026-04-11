package service

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	pkgcannon "github.com/andrlikjirka/dp-teals/pkg/canonicalizer"
	"github.com/andrlikjirka/dp-teals/pkg/logger"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/google/uuid"
)

// CheckpointService provides methods for creating and retrieving ledger checkpoints. It interacts with the CheckpointStore to persist and access checkpoint data, and uses the CheckpointSigner to sign checkpoint payloads. The service ensures that all operations are executed within a transaction context to maintain data consistency.
type CheckpointService struct {
	tx     ports.TransactionProvider
	signer ports.CheckpointSigner
	logger *logger.Logger
}

// NewCheckpointService creates a new instance of CheckpointService with the provided TransactionProvider, CheckpointSigner, and Logger. This service is responsible for managing ledger checkpoints, including creating new checkpoints and retrieving the latest anchored checkpoint from the storage.
func NewCheckpointService(tx ports.TransactionProvider, signer ports.CheckpointSigner, logger *logger.Logger) *CheckpointService {
	return &CheckpointService{
		tx:     tx,
		signer: signer,
		logger: logger,
	}
}

// CreateCheckpoint creates a new checkpoint for the current state of the ledger. It retrieves the ledger size and root hash, constructs a checkpoint payload, signs it using the CheckpointSigner, and stores the signed checkpoint in the CheckpointStore. The operation is executed within a transaction to ensure data consistency. If successful, it returns the created signed checkpoint; otherwise, it logs the error and returns it.
func (s *CheckpointService) CreateCheckpoint(ctx context.Context) (*model.SignedCheckpoint, error) {
	var sc *model.SignedCheckpoint

	err := s.tx.Transact(ctx, func(repos ports.Repositories) error {
		size, err := repos.LedgerProver.Size(ctx)
		if err != nil {
			s.logger.Error("failed to get ledger size", "error", err)
			return svcerrors.ErrLedgerSizeFailed
		}
		if size == 0 {
			return svcerrors.ErrCheckpointEmptyLedger
		}

		latest, err := repos.CheckpointStore.GetLatestSignedCheckpoint(ctx)
		if err != nil && !errors.Is(err, svcerrors.ErrCheckpointNotFound) {
			s.logger.Error("failed to get latest checkpoint", "error", err)
			return svcerrors.ErrGetCheckpointFailed
		}
		if latest != nil && latest.Checkpoint.Size == size {
			s.logger.Info("skipping checkpoint: ledger unchanged since last checkpoint", "size", size)
			return svcerrors.ErrCheckpointEmptyLedger
		}

		rootHash, err := repos.LedgerProver.RootHash(ctx)
		if err != nil {
			s.logger.Error("failed to get ledger root hash", "error", err)
			return svcerrors.ErrGetCheckpointFailed
		}

		anchoredAt := time.Now().UTC()
		canonical, err := pkgcannon.CanonicalizeCheckpoint(&pkgcannon.CheckpointPayload{
			RootHash:   hex.EncodeToString(rootHash),
			Size:       size,
			AnchoredAt: anchoredAt.Format(time.RFC3339Nano),
		})
		if err != nil {
			return svcerrors.ErrCheckpointCanonicalizationFailed
		}

		sigToken, err := s.signer.Sign(canonical)
		if err != nil {
			s.logger.Error("failed to sign checkpoint payload", "error", err)
			return svcerrors.ErrSignCheckpointFailed
		}

		sc = &model.SignedCheckpoint{
			ID: uuid.New(),
			Checkpoint: model.Checkpoint{
				Size:       size,
				RootHash:   rootHash,
				AnchoredAt: anchoredAt,
			},
			Kid:            s.signer.Kid(),
			SignatureToken: sigToken,
		}

		if err := repos.CheckpointStore.StoreCheckpoint(ctx, sc); err != nil {
			s.logger.Error("failed to store checkpoint", "error", err)
			return err
		}

		return nil

	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("signed checkpoint created and stored", "id", sc.ID, "size", sc.Checkpoint.Size)
	return sc, nil
}

// GetLatestCheckpoint retrieves the most recently anchored checkpoint from the CheckpointStore. It executes the retrieval within a transaction to ensure data consistency. If successful, it returns the signed checkpoint; otherwise, it logs the error and returns it.
func (s *CheckpointService) GetLatestCheckpoint(ctx context.Context) (*model.SignedCheckpoint, error) {
	var sc *model.SignedCheckpoint

	err := s.tx.Transact(ctx, func(r ports.Repositories) error {
		var err error
		sc, err = r.CheckpointStore.GetLatestSignedCheckpoint(ctx)
		return err
	})
	if err != nil {
		if errors.Is(err, svcerrors.ErrCheckpointNotFound) {
			s.logger.Info("no checkpoint found in storage")
			return nil, err
		}
		s.logger.Error("failed to retrieve latest signed checkpoint", "error", err)
		return nil, err
	}

	s.logger.Info("latest signed checkpoint retrieved", "id", sc.ID, "size", sc.Checkpoint.Size)
	return sc, nil
}

// ServerKid returns the key identifier (KID) used by the CheckpointSigner for signing checkpoint payloads. This KID can be used by clients to verify the authenticity of the checkpoint signatures by retrieving the corresponding public key.
func (s *CheckpointService) ServerKid() string {
	return s.signer.Kid()
}

// ServerPublicKey returns the public key used by the CheckpointSigner for signing checkpoint payloads. This public key can be used by clients to verify the authenticity of the checkpoint signatures.
func (s *CheckpointService) ServerPublicKey() []byte {
	return s.signer.PublicKey()
}
