package mmr

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/andrlikjirka/hash"
	"github.com/andrlikjirka/merkle"
)

// GenerateInclusionProof generates the proof for a leaf in the MMR by its index
func (m *MMR) GenerateInclusionProof(index int) (*merkle.InclusionProof, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if index < 0 || index >= len(m.Leaves) {
		return nil, errors.New("invalid index")
	}

	current := m.Leaves[index]
	var siblings [][]byte
	var left []bool

	// PHASE 1: Traverse from leaf up to its peak
	for current.Parent != nil {
		parent := current.Parent
		if parent.Left == current {
			siblings = append(siblings, parent.Right.Hash)
			left = append(left, false) // sibling is on the right
		} else {
			siblings = append(siblings, parent.Left.Hash)
			left = append(left, true) // sibling is on the left
		}
		current = parent
	}

	// 'current' is now the peak. We need to find its index in m.peaks.
	peakIdx := -1
	for i, p := range m.peaks {
		if bytes.Equal(p.Hash, current.Hash) {
			peakIdx = i
			break
		}
	}
	if peakIdx == -1 {
		return nil, errors.New("peak not found (internal state error)")
	}

	// PHASE 2: Peak Bagging
	// Combine all peaks to the RIGHT into a single hash, add as RIGHT sibling
	if peakIdx < len(m.peaks)-1 {
		rightBag := m.bagPeaksRightToLeft(m.peaks[peakIdx+1:])
		siblings = append(siblings, rightBag)
		left = append(left, false) // sibling is on the right
	}
	// Add all peaks to the LEFT sequentially as LEFT siblings
	for i := peakIdx - 1; i >= 0; i-- {
		siblings = append(siblings, m.peaks[i].Hash)
		left = append(left, true) // sibling is on the left
	}

	proof := &merkle.InclusionProof{Siblings: siblings, Left: left}
	return proof, nil
}

// GenerateInclusionProofByData generates the proof for a leaf in the MMR by its data content
func (m *MMR) GenerateInclusionProofByData(data []byte) (*merkle.InclusionProof, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	leafHash := merkle.HashLeafData(data, m.hashFunc)
	indices := m.indexMap[hex.EncodeToString(leafHash)]
	if len(indices) == 0 {
		return nil, errors.New("leaf not found in the MMR")
	}

	return m.GenerateInclusionProof(indices[0]) // generate proof for the first occurrence of the leaf (if duplicates exist)
}

// bagPeaksRightToLeft is a helper method that takes a slice of peaks and combines them into a single hash by hashing from right to left.
func (m *MMR) bagPeaksRightToLeft(peaks []*Node) []byte {
	if len(peaks) == 0 {
		return nil
	}
	if len(peaks) == 1 {
		return peaks[0].Hash
	}
	root := peaks[len(peaks)-1].Hash
	for i := len(peaks) - 2; i >= 0; i-- {
		root = merkle.HashInternalNodes(peaks[i].Hash, root, m.hashFunc)
	}
	return root
}

// VerifyInclusionProof verifies the inclusion proof for a given leaf data against the MMR root hash using the provided hash function.
func VerifyInclusionProof(leafData []byte, proof *merkle.InclusionProof, rootHash []byte, hashFunc hash.HashFunc) bool {
	return merkle.VerifyInclusionProof(leafData, proof, rootHash, hashFunc)
}
