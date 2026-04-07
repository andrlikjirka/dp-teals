package ports

import "context"

// Ledger defines the interface for interacting with the ledger, including appending leaves and managing signed roots.
type Ledger interface {
	AppendLeaf(ctx context.Context, leafHash []byte) (position int64, err error)
	SaveSignedRoot(ctx context.Context, root any) error
	GetLatestSignedRoot(ctx context.Context) (any, error)
}
