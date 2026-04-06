package repository

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql/query"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// ProducerKeyRepository provides methods to manage producer keys in the database. It implements the KeyRegistry interface defined in the service layer, allowing the service to add, retrieve, and revoke producer keys as needed for JWS signing and verification operations.
type ProducerKeyRepository struct {
	db sql.Db
}

// NewProducerKeyRepository creates a new instance of ProducerKeyRepository with the provided database connection.
func NewProducerKeyRepository(db sql.Db) *ProducerKeyRepository {
	return &ProducerKeyRepository{db: db}
}

// AddPublicKey adds a new producer key to the database. It executes an SQL query to insert the key details, and handles any errors that may occur during the operation. If a key with the same ID already exists, it returns a specific error indicating a duplicate key.
func (r *ProducerKeyRepository) AddPublicKey(ctx context.Context, key *svcmodel.ProducerKey) error {
	_, err := r.db.Exec(ctx, query.AddProducerPublicKey,
		key.ID, key.ProducerID, key.KeyID, []byte(key.PublicKey), string(key.Status), key.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return svcerrors.ErrDuplicateProducerKey
			case pgerrcode.ForeignKeyViolation:
				return svcerrors.ErrProducerNotFound
			}
		}
		return fmt.Errorf("store producer key: %w", err)
	}
	return nil
}

// PublicKey retrieves the public key for a given key ID (kid) from the database. It executes an SQL query to fetch the key details, and handles any errors that may occur during the operation. If no active key is found for the provided kid, it returns an error indicating that the public key was not found.
func (r *ProducerKeyRepository) PublicKey(ctx context.Context, kid string) (ed25519.PublicKey, error) {
	var record model.ProducerKeyRecord
	err := pgxscan.Get(ctx, r.db, &record, query.SelectProducerPublicKey, kid, svcmodel.KeyStatusActive)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, fmt.Errorf("public key not found for kid %q", kid)
		}
		return nil, fmt.Errorf("select public key: %w", err)
	}

	return record.PublicKey, nil
}

// RevokeKey revokes a producer key by updating its status in the database. It executes an SQL query to mark the key as revoked, and handles any errors that may occur during the operation.
func (r *ProducerKeyRepository) RevokeKey(ctx context.Context, kid string) error {
	tag, err := r.db.Exec(ctx, query.RevokeProducerPublicKey, kid)
	if err != nil {
		return fmt.Errorf("revoke producer key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return svcerrors.ErrKeyNotFound
	}
	return nil
}
