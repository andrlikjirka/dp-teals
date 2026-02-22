package merkle

import (
	"bytes"
	"encoding/hex"
	"errors"
)

type InclusionProof struct {
	Siblings [][]byte // Hashes of sibling nodes along the path to the root
	Left     []bool   // Indicates whether the sibling is a left sibling (true) or right sibling (false)
}

// GenerateInclusionProof generates an inclusion proof for the leaf at the specified index in the Merkle Tree.
func (t *Tree) GenerateInclusionProof(index int) (*InclusionProof, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if index < 0 || index >= len(t.Leaves) {
		return nil, errors.New("invalid index")
	}

	leaf := t.Leaves[index]
	current := leaf

	var siblings [][]byte
	var left []bool

	for current.Parent != nil { // start at the leaf and traverse up to the root
		parent := current.Parent    // jump to the parent node
		if parent.Left == current { // current is the left child of its parent
			siblings = append(siblings, parent.Right.Hash) // add the right sibling's hash to the proof
			left = append(left, false)                     // sibling is on the right
		} else {
			siblings = append(siblings, parent.Left.Hash) // add the left sibling's hash to the proof
			left = append(left, true)                     // sibling is on the left
		}
		current = parent // move up to the parent for the next iteration
	}

	proof := &InclusionProof{siblings, left}
	return proof, nil
}

func (t *Tree) GenerateInclusionProofByData(data []byte) (*InclusionProof, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	leafHash := hashLeafData(data, t.hashFunc)
	indices := t.indexMap[hex.EncodeToString(leafHash)]
	if len(indices) == 0 {
		return nil, errors.New("leaf not found in the tree")
	}

	return t.GenerateInclusionProof(indices[0]) // generate proof for the first occurrence of the leaf (if duplicates exist)
}

// VerifyInclusionProof verifies that the provided leaf data is included in the Merkle Tree with the given root hash using the provided inclusion proof.
func VerifyInclusionProof(leafData []byte, proof *InclusionProof, rootHash []byte, hashFunc HashFunc) bool {
	if hashFunc == nil {
		hashFunc = DefaultHashFunc
	}

	hash := hashFunc(append([]byte{0x00}, leafData...))

	for i, siblingHash := range proof.Siblings { // iterate through the proof and compute the hash up to the root
		if proof.Left[i] { // sibling is on the left}
			hash = hashFunc(append([]byte{0x01}, append(siblingHash, hash...)...))
		} else { // sibling is on the right
			hash = hashFunc(append([]byte{0x01}, append(hash, siblingHash...)...))
		}
	}

	return bytes.Equal(hash, rootHash)
}
