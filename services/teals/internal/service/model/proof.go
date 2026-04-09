package model

import (
	"github.com/andrlikjirka/dp-teals/pkg/mmr"
	"github.com/google/uuid"
)

type InclusionProofData struct {
	LeafIndex  int64
	LedgerSize int64
	LeafHash   []byte
	RootHash   []byte
	Proof      *mmr.InclusionProof
}

type InclusionProofResult struct {
	EventID       uuid.UUID
	LeafIndex     int64
	LeafEventHash []byte
	RootHash      []byte
	LedgerSize    int64
	Proof         *mmr.InclusionProof
}

type ConsistencyProofResult struct {
}
