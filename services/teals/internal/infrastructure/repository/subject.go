package repository

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql/query"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/georgysavva/scany/v2/pgxscan"
)

const secretSizeBytes = 32

// SubjectSecretRepository manages per-subject cryptographic secrets in the database.
type SubjectSecretRepository struct {
	db sql.Db
}

// NewSubjectSecretRepository creates a new instance of SubjectSecretRepository with the provided database connection. This repository is responsible for managing per-subject cryptographic secrets in the database, allowing the service layer to store and retrieve secrets as needed for metadata protection and other cryptographic operations.
func NewSubjectSecretRepository(db sql.Db) *SubjectSecretRepository {
	return &SubjectSecretRepository{db: db}
}

// GetOrCreateSecret retrieves the cryptographic secret associated with the given subject ID from the database, or creates a new one if it does not exist. It generates a random secret candidate and attempts to insert it into the database. If the insertion fails due to a conflict (i.e., the subject secret already exists), it fetches the existing secret instead. The method returns the retrieved or newly created secret, or an error if any operation fails during the process.
func (r *SubjectSecretRepository) GetOrCreateSecret(ctx context.Context, subjectID string) ([]byte, error) {
	candidate := make([]byte, secretSizeBytes)
	if _, err := rand.Read(candidate); err != nil {
		return nil, fmt.Errorf("generate subject secret: %w", err)
	}

	// try to insert the new secret, if it already exists, fetch the existing one
	var secret []byte
	err := r.db.QueryRow(ctx, query.GetOrCreateSubjectSecret, subjectID, candidate).Scan(&secret)
	if err == nil {
		return secret, nil
	}

	// subject secret already exists, fetch it
	err = pgxscan.Get(ctx, r.db, &secret, query.GetSubjectSecret, subjectID)
	if err != nil {
		return nil, fmt.Errorf("get subject secret: %w", err)
	}
	return secret, nil
}

// GetSecretBySubjectId retrieves the cryptographic secret associated with the given subject ID from the database. It executes an SQL query to fetch the secret, and handles any errors that may occur during the operation. If no secret is found for the specified subject ID, it returns an error indicating that the subject secret was not found.
func (r *SubjectSecretRepository) GetSecretBySubjectId(ctx context.Context, subjectID string) ([]byte, error) {
	var secret []byte
	err := pgxscan.Get(ctx, r.db, &secret, query.GetSubjectSecret, subjectID)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, svcerrors.ErrSubjectSecretNotFound
		}
		return nil, fmt.Errorf("get subject secret: %w", err)
	}
	return secret, nil
}

// DeleteSecretBySubjectId deletes the cryptographic secret associated with the given subject ID from the database. It executes an SQL query to delete the secret, and handles any errors that may occur during the operation. If no secret is found for the specified subject ID, it returns an error indicating that the subject secret was not found.
func (r *SubjectSecretRepository) DeleteSecretBySubjectId(ctx context.Context, subjectID string) error {
	tag, err := r.db.Exec(ctx, query.DeleteSubjectSecret, subjectID)
	if err != nil {
		return fmt.Errorf("delete subject secret: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return svcerrors.ErrSubjectSecretNotFound
	}
	return nil
}
