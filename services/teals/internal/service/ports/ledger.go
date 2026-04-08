package ports

import (
	"context"
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
	GenerateInclusionProof(ctx context.Context, leafIndex int64) (proof any, err error) // TODO proof
	// GenerateConsistencyProof generates a consistency proof between two specified indices in the MMR ledger, demonstrating that the ledger has evolved correctly from the earlier state to the later state.
	GenerateConsistencyProof(ctx context.Context, fromIndex int64, toIndex int64) (proof any, err error) // TODO proof
}
