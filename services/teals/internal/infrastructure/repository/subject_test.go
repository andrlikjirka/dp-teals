package repository_test

import (
	"context"
	"testing"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubjectSecretRepository_GetOrCreateSecret(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewSubjectSecretRepository(testPool)

	t.Run("CreatesNewSecret", func(t *testing.T) {
		truncateTables(t)

		secret, err := repo.GetOrCreateSecret(ctx, "subject-1")

		require.NoError(t, err)
		assert.Len(t, secret, 32)
	})

	t.Run("ReturnsSameSecretOnSubsequentCalls", func(t *testing.T) {
		truncateTables(t)

		first, err := repo.GetOrCreateSecret(ctx, "subject-1")
		require.NoError(t, err)

		second, err := repo.GetOrCreateSecret(ctx, "subject-1")
		require.NoError(t, err)

		assert.Equal(t, first, second)
	})

	t.Run("DifferentSubjectsGetDifferentSecrets", func(t *testing.T) {
		truncateTables(t)

		first, err := repo.GetOrCreateSecret(ctx, "subject-1")
		require.NoError(t, err)

		second, err := repo.GetOrCreateSecret(ctx, "subject-2")
		require.NoError(t, err)

		assert.NotEqual(t, first, second)
	})
}

func TestSubjectSecretRepository_GetSecretBySubjectId(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewSubjectSecretRepository(testPool)

	t.Run("ReturnsSecret", func(t *testing.T) {
		truncateTables(t)

		created, err := repo.GetOrCreateSecret(ctx, "subject-1")
		require.NoError(t, err)

		got, err := repo.GetSecretBySubjectId(ctx, "subject-1")

		require.NoError(t, err)
		assert.Equal(t, created, got)
	})

	t.Run("NotFound", func(t *testing.T) {
		truncateTables(t)

		_, err := repo.GetSecretBySubjectId(ctx, "nonexistent")

		assert.ErrorIs(t, err, svcerrors.ErrSubjectSecretNotFound)
	})
}

func TestSubjectSecretRepository_DeleteSecretBySubjectId(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewSubjectSecretRepository(testPool)

	t.Run("DeletesExistingSecret", func(t *testing.T) {
		truncateTables(t)

		_, err := repo.GetOrCreateSecret(ctx, "subject-1")
		require.NoError(t, err)

		err = repo.DeleteSecretBySubjectId(ctx, "subject-1")

		require.NoError(t, err)
	})

	t.Run("SecretInaccessibleAfterDeletion", func(t *testing.T) {
		truncateTables(t)

		_, err := repo.GetOrCreateSecret(ctx, "subject-1")
		require.NoError(t, err)

		require.NoError(t, repo.DeleteSecretBySubjectId(ctx, "subject-1"))

		_, err = repo.GetSecretBySubjectId(ctx, "subject-1")
		assert.ErrorIs(t, err, svcerrors.ErrSubjectSecretNotFound)
	})

	t.Run("NotFound", func(t *testing.T) {
		truncateTables(t)

		err := repo.DeleteSecretBySubjectId(ctx, "nonexistent")

		assert.ErrorIs(t, err, svcerrors.ErrSubjectSecretNotFound)
	})
}
