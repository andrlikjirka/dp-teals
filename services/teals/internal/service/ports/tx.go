package ports

import "context"

type Repositories struct {
	AuditLog AuditLog
	// ProducerKey KeyRegistry
	//Ledger   Ledger
}

type TransactionProvider interface {
	Transact(ctx context.Context, fn func(repos Repositories) error) error
}
