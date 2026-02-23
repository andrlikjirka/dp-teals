package merkle

import (
	"bytes"
	"errors"
	"math/bits"

	"github.com/andrlikjirka/hash"
)

type ConsistencyProof struct {
	Hashes [][]byte // Hashes of the nodes needed to verify consistency
}

// GenerateConsistencyProof generates a consistency proof for the first m leaves of the tree. It returns an error if m is invalid.
func (t *Tree) GenerateConsistencyProof(m int) (*ConsistencyProof, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	n := len(t.Leaves)
	if m <= 0 || m > n {
		return nil, errors.New("invalid m: must be between 1 and the number of leaves")
	}
	hashes := t.subProofRecursively(m, 0, n, true)
	return &ConsistencyProof{Hashes: hashes}, nil
}

// subProofRecursively generates the consistency proof recursively. It returns the hashes needed to verify that the first m leaves are consistent with the full tree.
func (t *Tree) subProofRecursively(m int, start int, n int, b bool) [][]byte {
	if m == n {
		if b {
			return [][]byte{}
		}
		return [][]byte{t.subtreeHash(start, n)}
	}

	k := largestPowerOfTwoLessThan(n)
	if m <= k {
		proof := t.subProofRecursively(m, start, k, b)
		rightHash := t.subtreeHash(start+k, n-k)
		return append(proof, rightHash)
	}
	proof := t.subProofRecursively(m-k, start+k, n-k, false)
	leftHash := t.subtreeHash(start, k)
	return append(proof, leftHash)
}

func (t *Tree) subtreeHash(start int, n int) []byte {
	if n == 1 { // if it's a leaf, return its hash directly
		return t.Leaves[start].Hash
	}

	return t.findHashTopDown(t.root, 0, len(t.Leaves), start, n) // if the subtree is not a leaf, we need to find its root hash by navigating the tree
}

// findHashTopDown navigates the tree boundaries to locate a pre-computed hash
func (t *Tree) findHashTopDown(node *Node, nodeStart int, nodeN int, targetStart int, targetN int) []byte {
	// base Case: We found the exact internal node representing this subtree!
	if nodeStart == targetStart && nodeN == targetN {
		return node.Hash
	}

	// find where the current node splits its children
	k := largestPowerOfTwoLessThan(nodeN)

	// if the target starts before the split point, it MUST be down the left branch
	if targetStart < nodeStart+k {
		return t.findHashTopDown(node.Left, nodeStart, k, targetStart, targetN)
	}

	// otherwise, it MUST be down the right branch
	return t.findHashTopDown(node.Right, nodeStart+k, nodeN-k, targetStart, targetN)
}

// largestPowerOfTwoLessThan returns the largest power of two less than n. For example, if n is 10, it returns 8.
func largestPowerOfTwoLessThan(n int) int {
	return 1 << (bits.Len(uint(n-1)) - 1)
}

// VerifyConsistencyProof verifies that the new root is consistent with the old root using the provided consistency proof.
func VerifyConsistencyProof(m, n int, oldRoot, newRoot []byte, proof *ConsistencyProof, hashFunc hash.HashFunc) bool {
	if hashFunc == nil {
		hashFunc = hash.DefaultHashFunc
	}

	if m == n {
		return bytes.Equal(oldRoot, newRoot) && len(proof.Hashes) == 0
	}
	if m <= 0 || m > n {
		return false
	}

	// the consistency proof verification process involves reconstructing the old root and the new root using the provided proof hashes
	// helper function verifySubProof is used to do this recursively
	computedOld, computedNew, remaining, err := verifySubProof(m, n, true, proof.Hashes, oldRoot, hashFunc)

	if err != nil { // if there was an error during verification, the proof is invalid
		return false
	}
	if len(remaining) != 0 { // if there are any remaining hashes in the proof that were not used, the proof is invalid
		return false
	}
	return bytes.Equal(computedOld, oldRoot) && bytes.Equal(computedNew, newRoot) // return true if both the computed old root and the computed new root match the provided old and new roots
}

// verifySubProof is a helper function that recursively verifies the consistency proof. It returns the computed old root, the computed new root, any remaining proof hashes, and an error if the proof is invalid.
func verifySubProof(m, n int, b bool, proofHashes [][]byte, oldRoot []byte, hashFunc hash.HashFunc) ([]byte, []byte, [][]byte, error) {
	if m == n { //zoomed in on a subtree that is perfectly identical in both trees
		if b { // looking at the exact branch that formed the original oldRoot
			return oldRoot, oldRoot, proofHashes, nil
		}
		// looking at a new branch that didn't exist in the old tree
		if len(proofHashes) == 0 {
			return nil, nil, nil, errors.New("proof too short")
		}
		h := proofHashes[0]               // pop the very first hash off the proof array
		return h, h, proofHashes[1:], nil // return this hash as both the computed old root and the computed new root for this subtree, and return the remaining proof hashes
	}

	k := largestPowerOfTwoLessThan(n) // find the split point of the current subtree to look deeper

	if m <= k { // if the old tree fits entirely inside the left half of the new tree
		oldHash, newLeft, remainingProof, err := verifySubProof(m, k, b, proofHashes, oldRoot, hashFunc) // recursively verify the left subtree
		if err != nil {
			return nil, nil, nil, err
		}
		if len(remainingProof) == 0 {
			return nil, nil, nil, errors.New("proof too short")
		}
		newRight := remainingProof[0]                                     // right side is entirely new, so the prover provides its hash directly
		combinedNewRoot := HashInternalNodes(newLeft, newRight, hashFunc) // combine the new left and new right to get the computed new root for this subtree
		return oldHash, combinedNewRoot, remainingProof[1:], nil          // return the computed old root, the computed new root, and the remaining proof hashes
	}
	// if old tree was large enough that it completely filled the left half and spilled over into the right half
	oldRight, newRight, remainingProof, err := verifySubProof(m-k, n-k, false, proofHashes, oldRoot, hashFunc) // recursively verify the right subtree
	if err != nil {
		return nil, nil, nil, err
	}
	if len(remainingProof) == 0 {
		return nil, nil, nil, errors.New("proof too short")
	}
	leftHash := remainingProof[0] //entire left half is identical in both the old and new trees, so the prover provides its single combined hash
	combinedOldRoot := HashInternalNodes(leftHash, oldRight, hashFunc)
	combinedNewRoot := HashInternalNodes(leftHash, newRight, hashFunc)

	return combinedOldRoot, combinedNewRoot, remainingProof[1:], nil // return the computed old root, the computed new root, and the remaining proof hashes
}
