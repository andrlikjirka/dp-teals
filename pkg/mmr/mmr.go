package mmr

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/andrlikjirka/hash"
	"github.com/andrlikjirka/merkle"
)

type Node struct {
	Hash   []byte
	Left   *Node
	Right  *Node
	Height int // 0 for leaves, +1 for each merge
}

type MMR struct {
	peaks    []*Node
	hashFunc hash.HashFunc
	size     int // Number of leaves appended
	lock     sync.RWMutex
}

// NewMMR initializes a new MMR with the provided hash function.
func NewMMR(hashFunc hash.HashFunc) *MMR {
	if hashFunc == nil {
		hashFunc = hash.DefaultHashFunc
	}
	return &MMR{
		peaks:    make([]*Node, 0),
		hashFunc: hashFunc,
		size:     0,
	}
}

func (m *MMR) Append(data []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(data) == 0 {
		return errors.New("empty leaf not allowed")
	}

	leafHash := m.hashFunc(data)
	newNode := &Node{
		Hash:   leafHash,
		Height: 0,
	}
	m.size++ // update the MMR size to reflect the new leaf

	// check if we can merge with existing peaks
	for len(m.peaks) > 0 {
		lastPeak := m.peaks[len(m.peaks)-1] // get the last peak (the rightmost one)
		if lastPeak.Height != newNode.Height {
			break // if the heights don't match, we can't merge, so we stop (only balanced peaks of the same height are merged )
		}
		m.peaks = m.peaks[:len(m.peaks)-1] // pop the last peak from the list

		// merge the two nodes
		mergedHash := merkle.HashInternalNodes(lastPeak.Hash, newNode.Hash, m.hashFunc)
		newNode = &Node{
			Hash:   mergedHash,
			Left:   lastPeak,
			Right:  newNode,
			Height: lastPeak.Height + 1,
		}

	}
	m.peaks = append(m.peaks, newNode) // push the resulting mountain peak back onto the list

	return nil
}

// RootHash computes the root hash of the MMR by combining all peaks. The order of peaks is important for consistency.
// The MMR root is the hash of all current peaks combined from right to left.
func (m *MMR) RootHash() []byte {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if len(m.peaks) == 0 {
		return nil
	}

	root := m.peaks[len(m.peaks)-1].Hash // start with the rightmost peak
	for i := len(m.peaks) - 2; i >= 0; i-- {
		root = merkle.HashInternalNodes(m.peaks[i].Hash, root, m.hashFunc) // combine peaks from right to left
	}
	return root
}

// PrintSummary provides a concise overview of the MMR's current state, including the number of leaves, the number of peaks, and the current root hash.
// This is useful for quickly assessing the MMR's status without delving into the full tree structure.
func (m *MMR) PrintSummary() {
	m.lock.RLock()
	defer m.lock.RUnlock()

	fmt.Println("=========== MMR Summary ===========")
	fmt.Printf("Size (leaves): %d\n", m.size)
	fmt.Printf("Number of peaks: %d\n", len(m.peaks))

	root := m.RootHash()
	if root == nil {
		fmt.Println("Root: <nil>")
	} else {
		fmt.Printf("Root: %s\n", hex.EncodeToString(root))
	}
}

// PrintPeaks displays the current peaks in the MMR, showing their height and a truncated hash for easy visualization.
func (m *MMR) PrintPeaks() {
	m.lock.RLock()
	defer m.lock.RUnlock()

	fmt.Println("----------- Peaks -----------")
	for i, peak := range m.peaks {
		hashStr := hex.EncodeToString(peak.Hash)
		fmt.Printf("Peak %d | Height: %d | Hash: %s\n",
			i,
			peak.Height,
			hashStr[:8],
		)
	}
	fmt.Println("-----------------------------")
}

// PrintTree visualizes the MMR structure in a tree-like format, showing the relationships between peaks and their hashes. It uses indentation to represent the tree structure, with the rightmost peak at the top and leftmost at the bottom.
func (m *MMR) PrintTree() {
	m.lock.RLock()
	defer m.lock.RUnlock()

	fmt.Println("============= MMR Tree =============")

	for i, peak := range m.peaks {
		fmt.Printf("Peak %d (height %d):\n", i, peak.Height)
		printNodeRecursive(peak, "", true)
		fmt.Println()
	}

	fmt.Println("=====================================")
}

// printNodeRecursive is a helper function to recursively print the tree structure of the MMR. It uses indentation and special characters to visually represent the tree hierarchy. The right subtree is printed first to make the tree grow upwards visually.
func printNodeRecursive(n *Node, prefix string, isTail bool) {
	if n == nil {
		return
	}

	hashStr := hex.EncodeToString(n.Hash)

	if n.Right != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		printNodeRecursive(n.Right, newPrefix, false)
	}

	fmt.Printf("%s", prefix)
	if isTail {
		fmt.Printf("└── ")
	} else {
		fmt.Printf("┌── ")
	}
	fmt.Printf("%s\n", hashStr[:8])

	if n.Left != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		printNodeRecursive(n.Left, newPrefix, true)
	}
}
