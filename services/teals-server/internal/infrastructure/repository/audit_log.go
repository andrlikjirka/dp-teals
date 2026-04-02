package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/andrlijirka/dp-teals/services/teals-server/internal/infrastructure/repository/sql/query"
	svcerrors "github.com/andrlijirka/dp-teals/services/teals-server/internal/service/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewAuditLogRepository(pool *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{pool: pool}
}

func (r *AuditLogRepository) StoreAuditLogEntry(ctx context.Context, eventId uuid.UUID, payload json.RawMessage) error {
	_, err := r.pool.Exec(ctx, query.InsertAuditEvent, eventId, payload)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return svcerrors.ErrDuplicateEventID
			}
		}
		return err
	}

	return nil
}
