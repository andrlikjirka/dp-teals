package ports

import "context"

type Repositories struct {
	AuditLog     AuditLog
	ProducerKeys ProducerKeyRegistry
	//Ledger   Ledger
}

type TransactionProvider interface {
	Transact(ctx context.Context, fn func(repos Repositories) error) error
}
