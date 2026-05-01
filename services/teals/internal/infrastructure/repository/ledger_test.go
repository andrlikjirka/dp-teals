package repository_test

import (
	"context"
	"testing"

	"github.com/andrlikjirka/dp-teals/pkg/hash"
	pkgmmr "github.com/andrlikjirka/dp-teals/pkg/mmr"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newLedgerRepo creates a LedgerRepository using the same hash function as
func newLedgerRepo() *repository.LedgerRepository {
	return repository.NewLedgerRepository(testPool, hash.SHA3HashFunc)
}

func TestLedgerRepository_Size(t *testing.T) {
	ctx := context.Background()

	t.Run("EmptyLedgerReturnsZero", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()

		size, err := repo.Size(ctx)

		require.NoError(t, err)
		assert.Equal(t, int64(0), size)
	})

	t.Run("ReturnsCorrectSizeAfterAppending", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()
		require.NoError(t, appendLeaves(t, repo, []byte("a"), []byte("b"), []byte("c")))

		size, err := repo.Size(ctx)

		require.NoError(t, err)
		assert.Equal(t, int64(3), size)
	})
}

func TestLedgerRepository_RootHash(t *testing.T) {
	ctx := context.Background()

	t.Run("EmptyLedgerReturnsNil", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()

		rootHash, err := repo.RootHash(ctx)

		require.NoError(t, err)
		assert.Nil(t, rootHash)
	})

	t.Run("SingleLeafReturnsLeafHash", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()
		payload := []byte("leaf-payload")
		_, _, err := repo.AppendLeaf(ctx, payload)
		require.NoError(t, err)

		rootHash, err := repo.RootHash(ctx)

		require.NoError(t, err)
		assert.Equal(t, pkgmmr.HashLeafData(payload, hash.SHA3HashFunc), rootHash)
	})

	t.Run("TwoLeavesRootIsInternalNodeHash", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()
		p1, p2 := []byte("leaf-1"), []byte("leaf-2")
		require.NoError(t, appendLeaves(t, repo, p1, p2))

		rootHash, err := repo.RootHash(ctx)

		require.NoError(t, err)
		h1 := pkgmmr.HashLeafData(p1, hash.SHA3HashFunc)
		h2 := pkgmmr.HashLeafData(p2, hash.SHA3HashFunc)
		assert.Equal(t, pkgmmr.HashInternalNodes(h1, h2, hash.SHA3HashFunc), rootHash)
	})

	t.Run("RootChangesAfterEachAppend", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()

		_, _, err := repo.AppendLeaf(ctx, []byte("leaf-1"))
		require.NoError(t, err)
		root1, err := repo.RootHash(ctx)
		require.NoError(t, err)

		_, _, err = repo.AppendLeaf(ctx, []byte("leaf-2"))
		require.NoError(t, err)
		root2, err := repo.RootHash(ctx)
		require.NoError(t, err)

		assert.NotEqual(t, root1, root2)
	})
}

func TestLedgerRepository_AppendLeaf(t *testing.T) {
	ctx := context.Background()

	t.Run("EmptyPayloadReturnsError", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()

		_, _, err := repo.AppendLeaf(ctx, []byte{})

		assert.Error(t, err)
	})

	t.Run("FirstLeafReturnsLeafIndexZero", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()

		_, leafIndex, err := repo.AppendLeaf(ctx, []byte("leaf-1"))

		require.NoError(t, err)
		assert.Equal(t, int64(0), leafIndex)
	})

	t.Run("LeafIndexIncrementsWithEachAppend", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()

		_, idx0, err := repo.AppendLeaf(ctx, []byte("leaf-0"))
		require.NoError(t, err)
		_, idx1, err := repo.AppendLeaf(ctx, []byte("leaf-1"))
		require.NoError(t, err)
		_, idx2, err := repo.AppendLeaf(ctx, []byte("leaf-2"))
		require.NoError(t, err)

		assert.Equal(t, int64(0), idx0)
		assert.Equal(t, int64(1), idx1)
		assert.Equal(t, int64(2), idx2)
	})
}

func TestLedgerRepository_GenerateInclusionProof(t *testing.T) {
	ctx := context.Background()

	t.Run("NonExistentLeafIndexReturnsError", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()
		_, _, err := repo.AppendLeaf(ctx, []byte("leaf-0"))
		require.NoError(t, err)

		_, err = repo.GenerateInclusionProof(ctx, 99, 1)

		assert.Error(t, err)
	})

	t.Run("SingleLeafProofVerifies", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()
		payload := []byte("leaf-0")
		_, _, err := repo.AppendLeaf(ctx, payload)
		require.NoError(t, err)
		rootHash, err := repo.RootHash(ctx)
		require.NoError(t, err)

		proof, err := repo.GenerateInclusionProof(ctx, 0, 1)

		require.NoError(t, err)
		assert.True(t, pkgmmr.VerifyInclusionProof(payload, proof.Proof, rootHash, hash.SHA3HashFunc))
	})

	t.Run("ProofVerifiesForAllLeavesInThreeLeafTree", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()
		payloads := [][]byte{[]byte("leaf-0"), []byte("leaf-1"), []byte("leaf-2")}
		require.NoError(t, appendLeaves(t, repo, payloads...))
		rootHash, err := repo.RootHash(ctx)
		require.NoError(t, err)

		for i, payload := range payloads {
			proof, err := repo.GenerateInclusionProof(ctx, int64(i), int64(len(payloads)))
			require.NoError(t, err, "leaf index %d", i)
			assert.True(t,
				pkgmmr.VerifyInclusionProof(payload, proof.Proof, rootHash, hash.SHA3HashFunc),
				"inclusion proof failed for leaf index %d", i,
			)
		}
	})
}

func TestLedgerRepository_GenerateConsistencyProof(t *testing.T) {
	ctx := context.Background()

	t.Run("InvalidRange_NegativeFromSize", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()
		require.NoError(t, appendLeaves(t, repo, []byte("leaf-0")))

		_, err := repo.GenerateConsistencyProof(ctx, -1, 1)

		assert.ErrorIs(t, err, svcerrors.ErrInvalidConsistencyProofRange)
	})

	t.Run("InvalidRange_FromSizeExceedsToSize", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()
		require.NoError(t, appendLeaves(t, repo, []byte("a"), []byte("b"), []byte("c")))

		_, err := repo.GenerateConsistencyProof(ctx, 3, 1)

		assert.ErrorIs(t, err, svcerrors.ErrInvalidConsistencyProofRange)
	})

	t.Run("InvalidRange_ToSizeExceedsCurrentSize", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()
		require.NoError(t, appendLeaves(t, repo, []byte("leaf-0")))

		_, err := repo.GenerateConsistencyProof(ctx, 0, 100)

		assert.ErrorIs(t, err, svcerrors.ErrInvalidConsistencyProofRange)
	})

	t.Run("EqualSizesReturnsTrivialProof", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()
		require.NoError(t, appendLeaves(t, repo, []byte("a"), []byte("b")))

		proof, err := repo.GenerateConsistencyProof(ctx, 2, 2)

		require.NoError(t, err)
		assert.Equal(t, 2, proof.OldSize)
		assert.Equal(t, 2, proof.NewSize)
		assert.Empty(t, proof.OldPeaksHashes)
		assert.Empty(t, proof.ConsistencyPaths)
	})

	t.Run("GrowthProofVerifies", func(t *testing.T) {
		truncateTables(t)
		repo := newLedgerRepo()

		// build to size 1, record old root
		require.NoError(t, appendLeaves(t, repo, []byte("leaf-0")))
		oldRoot, err := repo.RootHash(ctx)
		require.NoError(t, err)

		// grow to size 3, record new root
		require.NoError(t, appendLeaves(t, repo, []byte("leaf-1"), []byte("leaf-2")))
		newRoot, err := repo.RootHash(ctx)
		require.NoError(t, err)

		proof, err := repo.GenerateConsistencyProof(ctx, 1, 3)

		require.NoError(t, err)
		assert.True(t, pkgmmr.VerifyConsistencyProof(proof, oldRoot, newRoot, hash.SHA3HashFunc))
	})
}

// appendLeaves is a helper that appends multiple leaf payloads in sequence and fails the test immediately if any append returns an error.
func appendLeaves(t *testing.T, repo *repository.LedgerRepository, payloads ...[]byte) error {
	t.Helper()
	for _, p := range payloads {
		if _, _, err := repo.AppendLeaf(context.Background(), p); err != nil {
			return err
		}
	}
	return nil
}
