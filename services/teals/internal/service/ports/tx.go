package ports

import "context"

type Repositories struct {
	AuditLog AuditLogRepository
	//Ledger   LedgerRepository
}

type TransactionProvider interface {
	Transact(ctx context.Context, fn func(repos Repositories) error) error
}
