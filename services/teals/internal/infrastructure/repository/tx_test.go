package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
)

func TestTransactionProvider_Transact(t *testing.T) {
	ctx := context.Background()
	tp := repository.NewTransactionProvider(testPool)

	// Create a standalone repo to verify data outside the transaction context
	verifierRepo := repository.NewSubjectSecretRepository(testPool)

	t.Run("CommitsSuccessfully", func(t *testing.T) {
		truncateTables(t)
		subjectID := "subject-commit"

		// Act: Run operations inside the transaction
		err := tp.Transact(ctx, func(r ports.Repositories) error {
			_, err := r.SubjectSecretStore.GetOrCreateSecret(ctx, subjectID)
			return err // Returning nil triggers the COMMIT
		})

		// Assert: No error from the transaction
		require.NoError(t, err)

		// Verify: Check the database directly using our standalone connection
		// to ensure the commit actually persisted to disk.
		_, err = verifierRepo.GetSecretBySubjectId(ctx, subjectID)
		assert.NoError(t, err, "secret should be persisted to the database")
	})

	t.Run("RollsBackOnError", func(t *testing.T) {
		truncateTables(t)
		subjectID := "subject-rollback"
		expectedErr := errors.New("simulated domain error")

		// Act: Run operations inside the transaction, but fail at the end
		err := tp.Transact(ctx, func(r ports.Repositories) error {
			// 1. Create the secret inside the transaction
			_, err := r.SubjectSecretStore.GetOrCreateSecret(ctx, subjectID)
			require.NoError(t, err)

			// 2. Return an error to trigger the ROLLBACK
			return expectedErr
		})

		// Assert: The provider should bubble up our exact error
		require.ErrorIs(t, err, expectedErr)

		// Verify: Because of the rollback, the secret should NOT exist in the database
		_, err = verifierRepo.GetSecretBySubjectId(ctx, subjectID)
		assert.ErrorIs(t, err, svcerrors.ErrSubjectSecretNotFound, "secret should NOT exist due to rollback")
	})
}
