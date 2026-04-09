package service

import (
	"context"

	"github.com/andrlikjirka/dp-teals/pkg/logger"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/google/uuid"
)

// LedgerService provides methods to interact with the MMR ledger, such as generating inclusion proofs and retrieving the root hash.
type LedgerService struct {
	auditLog     ports.AuditLog
	ledgerProver ports.LedgerProver
	logger       *logger.Logger
}

// NewLedgerService creates a new instance of LedgerService with the provided AuditLog, LedgerProver, and Logger. This allows the service to generate inclusion proofs for audit log entries and retrieve the current root hash of the MMR ledger, while logging important information and errors during these operations.
func NewLedgerService(auditLog ports.AuditLog, ledgerProver ports.LedgerProver, l *logger.Logger) *LedgerService {
	return &LedgerService{
		auditLog:     auditLog,
		ledgerProver: ledgerProver,
		logger:       l,
	}
}

// GetInclusionProof retrieves the audit log entry for the given event ID and generates an inclusion proof for that entry in the MMR ledger. It returns the inclusion proof if successful, or an appropriate error if the audit log entry is not found or if there was an error generating the inclusion proof.
func (s *LedgerService) GetInclusionProof(ctx context.Context, eventID uuid.UUID) (*model.InclusionProofResult, error) {
	entry, err := s.auditLog.GetAuditLogEntryByEventID(ctx, eventID)
	if err != nil {
		s.logger.Error("failed to get audit log entry", "event_id", eventID, "error", err)
		return nil, svcerrors.ErrAuditLogEntryNotFound
	}
	proof, err := s.ledgerProver.GenerateInclusionProof(ctx, entry.LeafIndex)
	if err != nil {
		s.logger.Error("failed to generate inclusion proof", "leaf_index", entry.LeafIndex, "error", err)
		return nil, svcerrors.ErrInclusionProofFailed
	}

	s.logger.Info("inclusion proof generated successfully", "event_id", eventID, "leaf_index", entry.LeafIndex)

	return &model.InclusionProofResult{
		EventID:       entry.EventID,
		LeafIndex:     entry.LeafIndex,
		LeafEventHash: proof.LeafHash,
		RootHash:      proof.RootHash,
		LedgerSize:    proof.LedgerSize,
		Proof:         proof.Proof,
	}, nil
}

// GetRootHash retrieves the current root hash of the MMR ledger. It returns the root hash if successful, or an appropriate error if there was an error retrieving the root hash.
func (s *LedgerService) GetRootHash(ctx context.Context) ([]byte, error) {
	root, err := s.ledgerProver.RootHash(ctx)
	if err != nil {
		s.logger.Error("failed to get root hash", "error", err)
		return nil, svcerrors.ErrRootHashFailed
	}
	return root, nil
}
