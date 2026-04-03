package repository

import (
	"context"
)

type LedgerRepository interface {
	AppendLeaf(ctx context.Context, leafHash []byte) (position int64, err error)
	GetInclusionProof(ctx context.Context, leafIndex int64) (any, error)
	SaveSignedRoot(ctx context.Context, root any) error
	GetLatestRoot(ctx context.Context) (any, error)
}

// TODO: change the interface to struct implementation
