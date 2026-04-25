package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql/query"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	model3 "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
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

// GetAuditLogEntryByEventID retrieves an audit log entry from the database by its event ID. It executes an SQL query to fetch the entry details, and handles any errors that may occur during the operation. If no entry is found with the given event ID, it returns a specific error indicating that the audit log entry was not found.
func (r *AuditLogRepository) GetAuditLogEntryByEventID(ctx context.Context, eventID uuid.UUID) (*model3.AuditLogEntryRaw, error) {
	var record model.AuditLogEntryRecord
	err := pgxscan.Get(ctx, r.db, &record, query.GetAuditLogEntryByEventID, eventID)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, svcerrors.ErrAuditLogEntryNotFound
		}
		return nil, fmt.Errorf("get audit log entry by event id: %w", err)
	}

	return &model3.AuditLogEntryRaw{
		ID:             &record.ID,
		EventID:        record.EventID,
		ProducerKeyID:  record.ProducerKeyID,
		SignatureToken: record.SignatureToken,
		LeafIndex:      record.LeafIndex,
		CreatedAt:      record.CreatedAt,
		Payload:        record.Payload,
	}, nil
}

// ListAuditLogEntries retrieves a list of audit log entries from the database based on the provided filter criteria and pagination cursor. It constructs an SQL query dynamically using the squirrel library, executes the query to fetch matching entries, and handles any errors that may occur during the operation. The results are returned as a slice of AuditLogEntryRaw objects.
func (r *AuditLogRepository) ListAuditLogEntries(ctx context.Context, filter *model3.AuditEventFilter, cursor *int64, size int) ([]*model3.AuditLogEntryRaw, error) {
	sqlSelect, args, err := buildListAuditLogEntriesQuery(filter, cursor, size)
	if err != nil {
		return nil, fmt.Errorf("build list audit log entries query: %w", err)
	}

	var records []model.AuditLogEntryRecord
	err = pgxscan.Select(ctx, r.db, &records, sqlSelect, args...)
	if err != nil {
		return nil, fmt.Errorf("list audit log entries: %w", err)
	}

	results := make([]*model3.AuditLogEntryRaw, len(records))
	for i, rec := range records {
		id := rec.ID
		results[i] = &model3.AuditLogEntryRaw{
			ID:             &id,
			EventID:        rec.EventID,
			ProducerKeyID:  rec.ProducerKeyID,
			SignatureToken: rec.SignatureToken,
			LeafIndex:      rec.LeafIndex,
			CreatedAt:      rec.CreatedAt,
			Payload:        rec.Payload,
		}
	}
	return results, nil
}

// toStringSlice is a helper function that converts a slice of any type that is based on string (e.g., string, custom string types) to a slice of strings. This is useful for preparing arguments for SQL queries that expect text arrays.
func toStringSlice[T ~string](vals []T) []string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = string(v)
	}
	return out
}

// buildListAuditLogEntriesQuery constructs an SQL query for listing audit log entries based on the provided filter criteria. It uses the squirrel library to build the query dynamically, applying the appropriate WHERE clauses for each filter parameter. The function returns the final SQL query string, a slice of arguments for the query, and any error that may occur during query construction.
func buildListAuditLogEntriesQuery(filter *model3.AuditEventFilter, cursor *int64, size int) (string, []any, error) {
	q := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Select("le.id", "le.event_id", "le.mmr_node_id", "le.producer_key_id",
			"le.signature_token", "le.created_at", "le.payload", "mn.leaf_index").
		From("teals.log_entry le").
		Join("teals.mmr_node mn ON mn.id = le.mmr_node_id").
		OrderBy("le.id ASC").
		Limit(uint64(size))

	if len(filter.Actions) > 0 {
		q = q.Where("payload->>'action' = ANY(?::text[])", toStringSlice(filter.Actions))
	}
	if len(filter.ActorTypes) > 0 {
		q = q.Where("payload->'actor'->>'type' = ANY(?::text[])", toStringSlice(filter.ActorTypes))
	}
	if filter.ActorID != "" {
		q = q.Where("payload @> jsonb_build_object('actor', jsonb_build_object('id', ?::text))", filter.ActorID)
	}
	if filter.SubjectID != "" {
		q = q.Where("payload @> jsonb_build_object('subject', jsonb_build_object('id', ?::text))", filter.SubjectID)
	}
	if filter.ResourceID != "" {
		q = q.Where("payload @> jsonb_build_object('resource', jsonb_build_object('id', ?::text))", filter.ResourceID)
	}
	if filter.ResourceName != "" {
		q = q.Where("payload @> jsonb_build_object('resource', jsonb_build_object('name', ?::text))", filter.ResourceName)
	}
	if len(filter.ResultStatuses) > 0 {
		q = q.Where("payload->'result'->>'status' = ANY(?::text[])", toStringSlice(filter.ResultStatuses))
	}
	if filter.TimestampFrom != nil {
		q = q.Where("(payload->>'timestamp') >= ?", filter.TimestampFrom.Format("2006-01-02T15:04:05.000000Z"))
	}
	if filter.TimestampTo != nil {
		q = q.Where("(payload->>'timestamp') <= ?", filter.TimestampTo.Format("2006-01-02T15:04:05.000000Z"))
	}
	if cursor != nil {
		q = q.Where("le.id > ?", *cursor)
	}

	return q.ToSql()
}
