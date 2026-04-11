package repository

import (
	"context"
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

// CheckpointRepository provides methods to manage checkpoint data in the database. It implements the CheckpointStore interface defined in the service layer, allowing the service to store and retrieve signed checkpoints as needed for ledger anchoring and verification processes.
type CheckpointRepository struct {
	db sql.Db
}

// NewCheckpointRepository creates a new instance of CheckpointRepository with the provided database connection. This repository is responsible for managing checkpoint data in the database, allowing the service layer to store and retrieve signed checkpoints as needed for ledger anchoring and verification processes.
func NewCheckpointRepository(db sql.Db) *CheckpointRepository {
	return &CheckpointRepository{db: db}
}

// StoreCheckpoint persists a signed checkpoint in the database. It executes an SQL query to insert the checkpoint details, and handles any errors that may occur during the operation. If the insertion is successful, it returns nil; otherwise, it returns an error indicating the failure reason.
func (r *CheckpointRepository) StoreCheckpoint(ctx context.Context, sc *svcmodel.SignedCheckpoint) error {
	_, err := r.db.Exec(ctx, query.InsertCheckpoint,
		sc.ID,
		sc.Checkpoint.Size,
		sc.Checkpoint.RootHash,
		sc.Checkpoint.AnchoredAt,
		sc.Kid,
		sc.SignatureToken)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return svcerrors.ErrCheckpointAlreadyExists
			}
		}
		return fmt.Errorf("store checkpoint: %w", err)
	}
	return nil
}

// GetLatestSignedCheckpoint retrieves the most recently anchored checkpoint from the database. It executes an SQL query to fetch the checkpoint details, and handles any errors that may occur during the operation. If no checkpoint is found, it returns an error indicating that the checkpoint was not found.
func (r *CheckpointRepository) GetLatestSignedCheckpoint(ctx context.Context) (*svcmodel.SignedCheckpoint, error) {
	var record model.CheckpointRecord
	err := pgxscan.Get(ctx, r.db, &record, query.GetLatestCheckpoint)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, svcerrors.ErrCheckpointNotFound
		}
		return nil, fmt.Errorf("get latest checkpoint: %w", err)
	}

	return &svcmodel.SignedCheckpoint{
		ID: record.ID,
		Checkpoint: svcmodel.Checkpoint{
			Size:       record.Size,
			RootHash:   record.RootHash,
			AnchoredAt: record.AnchoredAt,
		},
		Kid:            record.Kid,
		SignatureToken: record.SignatureToken,
	}, nil
}
