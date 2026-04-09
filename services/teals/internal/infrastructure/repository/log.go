package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql/query"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/georgysavva/scany/v2/pgxscan"
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

// GetAuditLogEntryByEventID retrieves the audit log entry and its MMR leaf index for a given event ID.
func (r *AuditLogRepository) GetAuditLogEntryByEventID(ctx context.Context, eventID uuid.UUID) (*svcmodel.AuditLogEntry, error) {
	var record model.AuditLogEntryRecord
	err := pgxscan.Get(ctx, r.db, &record, query.GetAuditLogEntryByEventID, eventID)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, svcerrors.ErrAuditLogEntryNotFound
		}
		return nil, fmt.Errorf("get audit log entry by event id: %w", err)
	}

	return &svcmodel.AuditLogEntry{
		ID:             &record.ID,
		EventID:        record.EventID,
		ProducerKeyID:  record.ProducerKeyID,
		SignatureToken: record.SignatureToken,
		LeafIndex:      record.LeafIndex,
		CreatedAt:      record.CreatedAt,
		// TODO: Payload deserialization
	}, nil
}
