package service

import (
	"context"
	"errors"

	"github.com/andrlikjirka/dp-teals/pkg/logger"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/google/uuid"
)

// LedgerService provides methods to interact with the MMR ledger, such as generating inclusion proofs and retrieving the root hash.
type LedgerService struct {
	tx     ports.TransactionProvider
	logger *logger.Logger
}

// NewLedgerService creates a new instance of LedgerService with the provided TransactionProvider and Logger. This allows the service to manage database transactions and log important information and errors during ledger operations.
func NewLedgerService(tx ports.TransactionProvider, l *logger.Logger) *LedgerService {
	return &LedgerService{
		tx:     tx,
		logger: l,
	}
}

// GetInclusionProof retrieves the audit log entry for the given event ID and generates an inclusion proof for that entry in the MMR ledger. It returns the inclusion proof if successful, or an appropriate error if the audit log entry is not found or if there was an error generating the inclusion proof.
func (s *LedgerService) GetInclusionProof(ctx context.Context, eventID uuid.UUID) (*model.InclusionProofResult, error) {
	var result *model.InclusionProofResult

	err := s.tx.Transact(ctx, func(r ports.Repositories) error {
		entry, err := r.AuditLog.GetAuditLogEntryByEventID(ctx, eventID)
		if err != nil {
			s.logger.Error("failed to get audit log entry", "event_id", eventID, "error", err)
			return svcerrors.ErrAuditLogEntryNotFound
		}
		proof, err := r.LedgerProver.GenerateInclusionProof(ctx, entry.LeafIndex)
		if err != nil {
			s.logger.Error("failed to generate inclusion proof", "leaf_index", entry.LeafIndex, "error", err)
			return svcerrors.ErrInclusionProofFailed
		}

		result = &model.InclusionProofResult{
			EventID:       entry.EventID,
			LeafIndex:     entry.LeafIndex,
			LeafEventHash: proof.LeafHash,
			RootHash:      proof.RootHash,
			LedgerSize:    proof.LedgerSize,
			Proof:         proof.Proof,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.Info("inclusion proof generated successfully", "event_id", eventID)
	return result, nil
}

// GetRootHash retrieves the current root hash of the MMR ledger. It returns the root hash if successful, or an appropriate error if there was an error retrieving the root hash.
func (s *LedgerService) GetRootHash(ctx context.Context) ([]byte, error) {
	var root []byte
	err := s.tx.Transact(ctx, func(r ports.Repositories) error {
		var err error
		root, err = r.LedgerProver.RootHash(ctx)
		if err != nil {
			s.logger.Error("failed to get root hash", "error", err)
			return svcerrors.ErrRootHashFailed
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.Info("root hash calculated successfully")
	return root, nil
}

// GetConsistencyProof generates a consistency proof between two ledger sizes. It returns the consistency proof if successful, or an appropriate error if there was an error generating the consistency proof.
func (s *LedgerService) GetConsistencyProof(ctx context.Context, fromSize int64, toSize int64) (*model.ConsistencyProofResult, error) {
	if fromSize < 0 || fromSize > toSize {
		return nil, svcerrors.ErrInvalidConsistencyProofRange
	}

	var result *model.ConsistencyProofResult

	err := s.tx.Transact(ctx, func(r ports.Repositories) error {
		proof, err := r.LedgerProver.GenerateConsistencyProof(ctx, fromSize, toSize)
		if err != nil {
			if errors.Is(err, svcerrors.ErrInvalidConsistencyProofRange) {
				s.logger.Warn("invalid consistency proof range", "from_size", fromSize, "to_size", toSize)
				return svcerrors.ErrInvalidConsistencyProofRange
			}

			s.logger.Error("failed to generate consistency proof", "from_size", fromSize, "to_size", toSize, "error", err)
			return svcerrors.ErrConsistencyProofFailed
		}

		result = &model.ConsistencyProofResult{
			Proof: proof,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.Info("consistency proof generated successfully", "from_size", fromSize, "to_size", toSize)
	return result, nil
}
