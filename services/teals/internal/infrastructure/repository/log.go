package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql/query"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// AuditLogRepository provides methods to store audit log entries in the database. It interacts with the database using SQL queries defined in the query package, and handles any errors that may occur during the operation, including duplicate event IDs.
type AuditLogRepository struct {
	db sql.Db
}

// NewAuditLogRepository creates a new instance of AuditLogRepository with the provided database connection.
func NewAuditLogRepository(db sql.Db) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// StoreAuditLogEntry stores a new audit log entry in the database. It executes an SQL query to insert the event details, and handles any errors that may occur during the operation. If an entry with the same event ID already exists, it returns a specific error indicating a duplicate event ID.
func (r *AuditLogRepository) StoreAuditLogEntry(ctx context.Context, eventId uuid.UUID, payload json.RawMessage, sigToken string, producerKeyId uuid.UUID, nodeID int64) error {
	_, err := r.db.Exec(ctx, query.InsertAuditEvent, eventId, payload, sigToken, producerKeyId, nodeID)
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
