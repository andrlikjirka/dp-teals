package repository_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// insertProducer inserts a producer row directly (as there is no ProducerRepository implemented)
func insertProducer(t *testing.T, id uuid.UUID) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		`INSERT INTO teals.producer (id, name) VALUES ($1, $2)`,
		id, id.String(),
	)
	require.NoError(t, err)
}

func newProducerKey(producerID uuid.UUID) *svcmodel.ProducerKey {
	_, pub, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("generate ed25519 key: " + err.Error())
	}
	return &svcmodel.ProducerKey{
		ID:         uuid.New(),
		ProducerID: producerID,
		KeyID:      uuid.NewString(),
		PublicKey:  ed25519.PublicKey(pub),
		Status:     svcmodel.KeyStatusActive,
		CreatedAt:  time.Now().UTC().Truncate(time.Microsecond),
	}
}

func TestProducerKeyRepository_AddPublicKey(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewProducerKeyRepository(testPool)

	t.Run("AddsKey", func(t *testing.T) {
		truncateTables(t)
		producerID := uuid.New()
		insertProducer(t, producerID)

		err := repo.AddPublicKey(ctx, newProducerKey(producerID))

		require.NoError(t, err)
	})

	t.Run("DuplicateKidReturnsError", func(t *testing.T) {
		truncateTables(t)
		producerID := uuid.New()
		insertProducer(t, producerID)
		key := newProducerKey(producerID)
		require.NoError(t, repo.AddPublicKey(ctx, key))

		// second key reuses the same kid
		duplicate := newProducerKey(producerID)
		duplicate.KeyID = key.KeyID

		err := repo.AddPublicKey(ctx, duplicate)

		assert.ErrorIs(t, err, svcerrors.ErrDuplicateProducerKey)
	})

	t.Run("NonExistentProducerReturnsError", func(t *testing.T) {
		truncateTables(t)

		err := repo.AddPublicKey(ctx, newProducerKey(uuid.New()))

		assert.ErrorIs(t, err, svcerrors.ErrProducerNotFound)
	})
}

func TestProducerKeyRepository_PublicKey(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewProducerKeyRepository(testPool)

	t.Run("ReturnsPublicKey", func(t *testing.T) {
		truncateTables(t)
		producerID := uuid.New()
		insertProducer(t, producerID)
		key := newProducerKey(producerID)
		require.NoError(t, repo.AddPublicKey(ctx, key))

		got, err := repo.PublicKey(ctx, key.KeyID)

		require.NoError(t, err)
		assert.Equal(t, []byte(key.PublicKey), []byte(got))
	})

	t.Run("NonExistentKidReturnsError", func(t *testing.T) {
		truncateTables(t)

		_, err := repo.PublicKey(ctx, "nonexistent-kid")

		assert.Error(t, err)
	})

	t.Run("RevokedKeyReturnsError", func(t *testing.T) {
		truncateTables(t)
		producerID := uuid.New()
		insertProducer(t, producerID)
		key := newProducerKey(producerID)
		require.NoError(t, repo.AddPublicKey(ctx, key))
		require.NoError(t, repo.RevokeKey(ctx, key.KeyID))

		_, err := repo.PublicKey(ctx, key.KeyID)

		assert.Error(t, err)
	})
}

func TestProducerKeyRepository_GetProducerKeyByKid(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewProducerKeyRepository(testPool)

	t.Run("ReturnsProducerKey", func(t *testing.T) {
		truncateTables(t)
		producerID := uuid.New()
		insertProducer(t, producerID)
		key := newProducerKey(producerID)
		require.NoError(t, repo.AddPublicKey(ctx, key))

		got, err := repo.GetProducerKeyByKid(ctx, key.KeyID)

		require.NoError(t, err)
		assert.Equal(t, key.ID, got.ID)
		assert.Equal(t, key.ProducerID, got.ProducerID)
		assert.Equal(t, key.KeyID, got.KeyID)
		assert.Equal(t, []byte(key.PublicKey), []byte(got.PublicKey))
		assert.Equal(t, key.Status, got.Status)
		assert.True(t, key.CreatedAt.Equal(got.CreatedAt))
	})

	t.Run("NonExistentKidReturnsError", func(t *testing.T) {
		truncateTables(t)

		_, err := repo.GetProducerKeyByKid(ctx, "nonexistent-kid")

		assert.Error(t, err)
	})

	t.Run("RevokedKeyReturnsError", func(t *testing.T) {
		truncateTables(t)
		producerID := uuid.New()
		insertProducer(t, producerID)
		key := newProducerKey(producerID)
		require.NoError(t, repo.AddPublicKey(ctx, key))
		require.NoError(t, repo.RevokeKey(ctx, key.KeyID))

		_, err := repo.GetProducerKeyByKid(ctx, key.KeyID)

		assert.Error(t, err)
	})
}

func TestProducerKeyRepository_RevokeKey(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewProducerKeyRepository(testPool)

	t.Run("RevokesKey", func(t *testing.T) {
		truncateTables(t)
		producerID := uuid.New()
		insertProducer(t, producerID)
		key := newProducerKey(producerID)
		require.NoError(t, repo.AddPublicKey(ctx, key))

		err := repo.RevokeKey(ctx, key.KeyID)

		require.NoError(t, err)
	})

	t.Run("RevokedKeyNotAccessibleViaPublicKey", func(t *testing.T) {
		truncateTables(t)
		producerID := uuid.New()
		insertProducer(t, producerID)
		key := newProducerKey(producerID)
		require.NoError(t, repo.AddPublicKey(ctx, key))

		require.NoError(t, repo.RevokeKey(ctx, key.KeyID))

		_, err := repo.PublicKey(ctx, key.KeyID)
		assert.Error(t, err)
	})

	t.Run("NonExistentKidReturnsError", func(t *testing.T) {
		truncateTables(t)

		err := repo.RevokeKey(ctx, "nonexistent-kid")

		assert.ErrorIs(t, err, svcerrors.ErrKeyNotFound)
	})
}
