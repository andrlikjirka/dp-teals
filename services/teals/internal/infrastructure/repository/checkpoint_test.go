package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSignedCheckpoint(anchoredAt time.Time) *svcmodel.SignedCheckpoint {
	return &svcmodel.SignedCheckpoint{
		ID: uuid.New(),
		Checkpoint: svcmodel.Checkpoint{
			Size:       10,
			RootHash:   []byte("roothash"),
			AnchoredAt: anchoredAt.UTC(),
		},
		Kid:            "key-1",
		SignatureToken: "token-abc",
	}
}

func TestCheckpointRepository_StoreCheckpoint(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewCheckpointRepository(testPool)

	t.Run("StoresCheckpoint", func(t *testing.T) {
		truncateTables(t)

		sc := newSignedCheckpoint(time.Now())

		err := repo.StoreCheckpoint(ctx, sc)

		require.NoError(t, err)
	})

	t.Run("DuplicateIDReturnsError", func(t *testing.T) {
		truncateTables(t)

		sc := newSignedCheckpoint(time.Now())
		require.NoError(t, repo.StoreCheckpoint(ctx, sc))

		err := repo.StoreCheckpoint(ctx, sc)

		assert.ErrorIs(t, err, svcerrors.ErrCheckpointAlreadyExists)
	})
}

func TestCheckpointRepository_GetLatestSignedCheckpoint(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewCheckpointRepository(testPool)

	t.Run("ReturnsStoredCheckpoint", func(t *testing.T) {
		truncateTables(t)

		sc := newSignedCheckpoint(time.Now())
		require.NoError(t, repo.StoreCheckpoint(ctx, sc))

		got, err := repo.GetLatestSignedCheckpoint(ctx)

		require.NoError(t, err)
		assert.Equal(t, sc.ID, got.ID)
		assert.Equal(t, sc.Checkpoint.Size, got.Checkpoint.Size)
		assert.Equal(t, sc.Checkpoint.RootHash, got.Checkpoint.RootHash)
		assert.True(t, sc.Checkpoint.AnchoredAt.Equal(got.Checkpoint.AnchoredAt))
		assert.Equal(t, sc.Kid, got.Kid)
		assert.Equal(t, sc.SignatureToken, got.SignatureToken)
	})

	t.Run("ReturnsCheckpointWithLatestAnchoredAt", func(t *testing.T) {
		truncateTables(t)

		older := newSignedCheckpoint(time.Now().Add(-1 * time.Hour))
		newer := newSignedCheckpoint(time.Now())
		require.NoError(t, repo.StoreCheckpoint(ctx, older))
		require.NoError(t, repo.StoreCheckpoint(ctx, newer))

		got, err := repo.GetLatestSignedCheckpoint(ctx)

		require.NoError(t, err)
		assert.Equal(t, newer.ID, got.ID)
	})

	t.Run("EmptyReturnsNotFound", func(t *testing.T) {
		truncateTables(t)

		_, err := repo.GetLatestSignedCheckpoint(ctx)

		assert.ErrorIs(t, err, svcerrors.ErrCheckpointNotFound)
	})
}
