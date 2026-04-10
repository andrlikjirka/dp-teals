package ports

import (
	"context"

	"github.com/andrlikjirka/dp-teals/pkg/mmr"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
)

// Ledger defines the interface for appending leaves to the MMR ledger.
type Ledger interface {
	// AppendLeaf appends a new leaf with the given payload to the MMR ledger and returns the node ID of the newly added leaf.
	AppendLeaf(ctx context.Context, payload []byte) (nodeID int64, err error)
}

// LedgerProver defines the interface for generating proofs related to the MMR ledger, such as inclusion and consistency proofs.
type LedgerProver interface {
	// Size returns the current number of leaves in the MMR ledger.
	Size(ctx context.Context) (size int64, err error)
	// RootHash returns the current root hash of the MMR ledger.
	RootHash(ctx context.Context) (rootHash []byte, err error)
	// GenerateInclusionProof generates an inclusion proof for the leaf at the specified index in the MMR ledger.
	GenerateInclusionProof(ctx context.Context, leafIndex int64) (proof *model.InclusionProofData, err error)
	// GenerateConsistencyProof generates a consistency proof between two sizes of the MMR ledger, fromSize and toSize, where fromSize is less than or equal to toSize. This proof can be used to verify that the ledger has been extended correctly without any tampering.
	GenerateConsistencyProof(ctx context.Context, fromSize int64, toSize int64) (proof *mmr.ConsistencyProof, err error)
}
