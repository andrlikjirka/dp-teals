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
)

type AuditLogRepository struct {
	db db
}

func NewAuditLogRepository(db db) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) StoreAuditLogEntry(ctx context.Context, eventId uuid.UUID, payload json.RawMessage) error {
	_, err := r.db.Exec(ctx, query.InsertAuditEvent, eventId, payload)
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
