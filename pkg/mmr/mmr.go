package mmr

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/andrlikjirka/dp-teals/pkg/hash"
)

// Node represents a single node in the MMR, which can be a leaf or an internal node. Each node stores its hash, pointers to its left and right children (if any), a pointer to its parent, and its height in the tree. The height is 0 for leaves and increases by 1 for each merge of two nodes.
type Node struct {
	Hash   []byte
	Left   *Node
	Right  *Node
	Parent *Node
	Height int // 0 for leaves, +1 for each merge
}

// MMR (Merkle Mountain Range) is a data structure that maintains a dynamic collection of leaves and their corresponding peaks. It allows for efficient appending of new leaves and provides methods to compute the root hash, generate inclusion proofs, and consistency proofs. The MMR maintains an index map to track the positions of leaf hashes for quick proof generation. It uses a mutex to ensure thread-safe operations when modifying the structure.
type MMR struct {
	peaks    []*Node
	Leaves   []*Node
	indexMap map[string][]int // hash → indices
	hashFunc hash.Func
	size     int // Number of leaves appended
	lock     sync.RWMutex
}

// NewMMR initializes a new MMR instance with an optional custom hash function. If no hash function is provided, it defaults to the standard hash function defined in the hash package. The MMR starts with empty peaks and leaves, and an empty index map for tracking leaf hashes.
func NewMMR(hashFunc hash.Func) *MMR {
	if hashFunc == nil {
		hashFunc = hash.DefaultHashFunc
	}
	return &MMR{
		peaks:    make([]*Node, 0),
		Leaves:   make([]*Node, 0),
		indexMap: make(map[string][]int),
		hashFunc: hashFunc,
		size:     0,
	}
}

// Append adds a new leaf to the MMR with the given data.
// It computes the hash of the new leaf, creates a new node, and appends it to the list of leaves. The method then checks if the new node can be merged with existing peaks (if they have the same height) and merges them accordingly, updating the peaks list. The index map is updated to track the new leaf's hash and its index for future proof generation. The method returns an error if an attempt is made to append an empty leaf.
func (m *MMR) Append(data []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(data) == 0 {
		return errors.New("empty leaf not allowed")
	}

	leafHash := HashLeafData(data, m.hashFunc)
	newNode := &Node{
		Hash:   leafHash,
		Height: 0,
	}
	m.Leaves = append(m.Leaves, newNode)
	m.size++ // update the MMR size to reflect the new leaf

	// Update indexMap to track this leaf's hash to its index
	hashHex := hex.EncodeToString(leafHash)
	m.indexMap[hashHex] = append(m.indexMap[hashHex], len(m.Leaves)-1)

	// check if we can merge with existing peaks
	for len(m.peaks) > 0 {
		lastPeak := m.peaks[len(m.peaks)-1] // get the last peak (the rightmost one)
		if lastPeak.Height != newNode.Height {
			break // if the heights don't match, we can't merge, so we stop (only balanced peaks of the same height are merged )
		}
		m.peaks = m.peaks[:len(m.peaks)-1] // pop the last peak from the list

		rightChild := newNode
		mergedHash := HashInternalNodes(lastPeak.Hash, newNode.Hash, m.hashFunc) // merge the two nodes
		newNode = &Node{
			Hash:   mergedHash,
			Left:   lastPeak,
			Right:  rightChild,
			Height: lastPeak.Height + 1,
		}
		lastPeak.Parent = newNode
		rightChild.Parent = newNode
	}
	m.peaks = append(m.peaks, newNode) // push the resulting mountain peak back onto the list

	return nil
}

// RootHash computes the root hash of the MMR by combining all peaks (peak bagging). The order of peaks is important for consistency.
// The MMR root is the hash of all current peaks combined from right to left.
func (m *MMR) RootHash() []byte {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if len(m.peaks) == 0 {
		return nil
	}

	root := m.peaks[len(m.peaks)-1].Hash // start with the rightmost peak
	for i := len(m.peaks) - 2; i >= 0; i-- {
		root = HashInternalNodes(m.peaks[i].Hash, root, m.hashFunc) // combine peaks from right to left
	}
	return root
}

// ============ Debugging and Visualization Methods ============

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
