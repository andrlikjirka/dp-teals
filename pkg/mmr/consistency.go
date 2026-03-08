package mmr

import (
	"bytes"
	"errors"
	"math/bits"

	"github.com/andrlikjirka/hash"
	"github.com/andrlikjirka/merkle"
)

// ConsistencyPath represents the path from an old peak to a new peak in the MMR. It consists of the sibling hashes along the path and the direction (left/right) of each sibling (inclusion proof).
// This path is used to prove that a specific old peak is included in the new MMR structure, demonstrating that the old MMR is a prefix of the new MMR.
type ConsistencyPath struct {
	Siblings [][]byte
	Left     []bool
}

// ConsistencyProof represents the proof that a smaller MMR (with treeSize1 leaves) is a prefix of a larger MMR (with treeSize2 leaves).
type ConsistencyProof struct {
	OldSize          int
	NewSize          int
	OldPeaksHashes   [][]byte           // Hashes of the old peaks (for verification)
	ConsistencyPaths []*ConsistencyPath // Inclusion paths from old peaks to new peaks
	RightPeaks       [][]byte           // Additional peaks completing NewSize
}

// GenerateConsistencyProof proves that treeSize1 is a prefix of treeSize2.
// It constructs the proof by:
// 1. Identifying the peaks of the old tree (treeSize1) and the new tree (treeSize2).
// 2. For each old peak, it builds an inclusion path to the corresponding new peak in treeSize2.
// 3. Collecting any additional right peaks that exist in treeSize2 but not in treeSize1.
func (m *MMR) GenerateConsistencyProof(treeSize1, treeSize2 int) (*ConsistencyProof, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if treeSize1 < 0 || treeSize1 > treeSize2 || treeSize2 > m.size {
		return nil, errors.New("invalid tree sizes")
	}

	proof := &ConsistencyProof{
		OldSize:          treeSize1,
		NewSize:          treeSize2,
		OldPeaksHashes:   make([][]byte, 0),
		ConsistencyPaths: make([]*ConsistencyPath, 0),
		RightPeaks:       make([][]byte, 0),
	}

	if treeSize1 == treeSize2 {
		return proof, nil // Trivial case
	}

	oldPeaks := m.getPeaksAtSize(treeSize1)
	newPeaks := m.getPeaksAtSize(treeSize2)

	// Collect the hashes of the old peaks for the proof
	for _, peak := range oldPeaks {
		proof.OldPeaksHashes = append(proof.OldPeaksHashes, peak.Hash)
	}

	// Map newPeaks for $O(1)$ lookup to know when to stop traversing
	isNewPeak := make(map[*Node]int)
	for i, p := range newPeaks {
		isNewPeak[p] = i
	}

	lastNewPeakIndex := -1

	// 1. Build the consistency paths for each old peak
	// generating a standard inclusion proof for the old peak, proving that the old peak is buried inside the new mountain
	for _, oldPeak := range oldPeaks {
		var siblings [][]byte
		var left []bool
		current := oldPeak

		for {
			// Stop if we reach a node that is a peak in treeSize2
			if idx, ok := isNewPeak[current]; ok {
				if idx > lastNewPeakIndex {
					lastNewPeakIndex = idx
				}
				break
			}

			if current.Parent == nil {
				return nil, errors.New("internal state error: reached nil parent before finding new peak")
			}

			parent := current.Parent
			if parent.Left == current {
				siblings = append(siblings, parent.Right.Hash)
				left = append(left, false)
			} else {
				siblings = append(siblings, parent.Left.Hash)
				left = append(left, true)
			}
			current = parent
		}

		proof.ConsistencyPaths = append(proof.ConsistencyPaths, &ConsistencyPath{
			Siblings: siblings,
			Left:     left,
		})
	}

	// 2. Collect the right-peaks (any new peaks strictly to the right of the nodes affected by the old peaks)
	for i := lastNewPeakIndex + 1; i < len(newPeaks); i++ {
		proof.RightPeaks = append(proof.RightPeaks, newPeaks[i].Hash)
	}

	return proof, nil
}

// getPeaksAtSize retrieves the peak nodes as they existed when the MMR had exactly 'size' leaves.
// This is done by analyzing the binary representation of given MMR size to determine which peaks would have existed at that point in time.
func (m *MMR) getPeaksAtSize(size int) []*Node {
	// ensure size is within bounds
	if size > m.size {
		size = m.size
	}
	if size <= 0 {
		return nil
	}

	var peaks []*Node
	offset := 0

	// iterate through the bits of 'size' from MSB to LSB
	bitLen := bits.Len(uint(size))
	for bit := bitLen - 1; bit >= 0; bit-- {
		if (size & (1 << bit)) != 0 {
			// find the root of the 2^bit subtree starting at 'offset'
			node := m.Leaves[offset]
			for i := 0; i < bit; i++ {
				if node.Parent != nil {
					node = node.Parent
				}
			}
			peaks = append(peaks, node)
			offset += 1 << bit
		}
	}
	return peaks
}

// VerifyConsistencyProof checks if old peaks legally transition into newRoot.
// It verifies that the old peaks match the old root and that following the consistency paths from the old peaks leads to the new peaks, which then combine to form the new root.
func VerifyConsistencyProof(proof *ConsistencyProof, oldRoot []byte, newRoot []byte, hashFunc hash.Func) bool {
	if hashFunc == nil {
		hashFunc = hash.DefaultHashFunc
	}

	if proof.OldSize == proof.NewSize {
		return true // assumes roots matched out-of-band
	}
	if len(proof.OldPeaksHashes) != len(proof.ConsistencyPaths) {
		return false
	}

	// 1. Verify that the provided old peaks perfectly match our trusted old root
	// result: verifier has certainty that the provided OldPeaksHashes are the genuine foundation of the old tree
	if proof.OldSize > 0 {
		if len(proof.OldPeaksHashes) == 0 {
			return false
		}

		calculatedOldRoot := proof.OldPeaksHashes[len(proof.OldPeaksHashes)-1]
		for i := len(proof.OldPeaksHashes) - 2; i >= 0; i-- {
			calculatedOldRoot = merkle.HashInternalNodes(proof.OldPeaksHashes[i], calculatedOldRoot, hashFunc)
		}

		if !bytes.Equal(calculatedOldRoot, oldRoot) {
			return false
		}
	}

	// 2. For each old peak, follow the consistency path to calculate the corresponding new peak hash.
	// result: verifier has calculated the left side of the new mountain range
	var newPeaksHashes [][]byte
	for i, oldHash := range proof.OldPeaksHashes {
		currentHash := oldHash
		path := proof.ConsistencyPaths[i]
		if path == nil || len(path.Left) != len(path.Siblings) {
			return false
		}

		for j, sibling := range path.Siblings {
			if path.Left[j] {
				currentHash = merkle.HashInternalNodes(sibling, currentHash, hashFunc)
			} else {
				currentHash = merkle.HashInternalNodes(currentHash, sibling, hashFunc)
			}
		}

		// because MMRs only grow rightward, multiple old peaks often merge into the exact same new, taller peak
		// verifier checks for this and deduplicates them so they don't count the same mountain twice
		if len(newPeaksHashes) == 0 || !bytes.Equal(newPeaksHashes[len(newPeaksHashes)-1], currentHash) {
			newPeaksHashes = append(newPeaksHashes, currentHash)
		}
	}

	// 3. Append the right peaks from the proof to complete the new mountain range
	newPeaksHashes = append(newPeaksHashes, proof.RightPeaks...)
	if len(newPeaksHashes) == 0 {
		return false
	}

	// 4. Combine all the new peaks to calculate the new root and compare it with the provided new root.
	calculatedRoot := newPeaksHashes[len(newPeaksHashes)-1]
	for i := len(newPeaksHashes) - 2; i >= 0; i-- {
		calculatedRoot = merkle.HashInternalNodes(newPeaksHashes[i], calculatedRoot, hashFunc)
	}

	return bytes.Equal(calculatedRoot, newRoot)
}
