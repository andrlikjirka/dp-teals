package repository

import (
	"context"
	"errors"

	"github.com/andrlikjirka/dp-teals/pkg/hash"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TransactionProvider provides a way to execute multiple repository operations within a single database transaction.
type TransactionProvider struct {
	pool *pgxpool.Pool
}

// NewTransactionProvider creates a new TransactionProvider with the given database connection pool.
func NewTransactionProvider(pool *pgxpool.Pool) *TransactionProvider {
	return &TransactionProvider{
		pool: pool,
	}
}

// Transact executes the given function within a database transaction. It provides a set of repositories that use the same transaction context. If the function returns an error, the transaction is rolled back; otherwise, it is committed.
func (tp *TransactionProvider) Transact(ctx context.Context, txFunc func(ports.Repositories) error) error {
	return runInTransaction(ctx, tp.pool, func(tx pgx.Tx) error {
		ledgerRepo := NewLedgerRepository(tx, hash.SHA3HashFunc)
		subjectSecretRepo := NewSubjectSecretRepository(tx)

		r := ports.Repositories{
			AuditLog:           NewAuditLogRepository(tx),
			ProducerKeys:       NewProducerKeyRepository(tx),
			Ledger:             ledgerRepo,
			CheckpointStore:    NewCheckpointRepository(tx),
			SubjectSecretStore: subjectSecretRepo,
		}

		return txFunc(r)
	})
}

// runInTransaction is a helper function that executes the given function within a database transaction. It handles committing the transaction if the function succeeds, or rolling back if it returns an error.
func runInTransaction(ctx context.Context, db *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err == nil {
		return tx.Commit(ctx)
	}

	rollbackErr := tx.Rollback(ctx)
	if rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}
