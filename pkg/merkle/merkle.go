package merkle

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/andrlikjirka/hash"
)

type Node struct {
	Hash   []byte
	Left   *Node
	Right  *Node
	Parent *Node
}

type Tree struct {
	root     *Node
	Leaves   []*Node
	indexMap map[string][]int // hash → indices
	hashFunc hash.HashFunc
	lock     sync.RWMutex
}

// NewTree creates a new Merkle Tree from the provided data.
func NewTree(data [][]byte, hashFunc hash.HashFunc) (*Tree, error) {
	if len(data) == 0 {
		return nil, errors.New("no data provided")
	}

	if hashFunc == nil {
		hashFunc = hash.DefaultHashFunc
	}

	t := build(data, hashFunc)
	return t, nil
}

// build constructs the Merkle Tree from the provided data.
func build(data [][]byte, hashFunc hash.HashFunc) *Tree {
	var leaves []*Node
	indexMap := make(map[string][]int)
	// create leaf nodes
	for i, d := range data {
		leafHash := HashLeafData(d, hashFunc)
		leaves = append(leaves, &Node{Hash: leafHash})

		hashHex := hex.EncodeToString(leafHash)
		indexMap[hashHex] = append(indexMap[hashHex], i)
	}

	t := &Tree{
		Leaves:   leaves,
		indexMap: indexMap,
		hashFunc: hashFunc,
		root:     buildRecursive(leaves, hashFunc),
	}
	return t
}

// buildRecursive builds the tree recursively from the given nodes and returns the root node. It implements the tree construction logic defined in RFC 6962 to construct deterministic append-only binary trees (avoid data padding).
func buildRecursive(nodes []*Node, hashFunc hash.HashFunc) *Node {
	n := len(nodes)
	if n == 1 {
		return nodes[0] // Base case: if only one node, return it
	}

	k := largestPowerOfTwoLessThan(n) // find the largest power of two less than n to determine how to split the nodes into left and right halves

	// split the slice into left and right halves
	left := buildRecursive(nodes[:k], hashFunc)
	right := buildRecursive(nodes[k:], hashFunc)

	parentHash := HashInternalNodes(left.Hash, right.Hash, hashFunc) // compute the parent hash by combining the left and right child hashes

	parent := &Node{ // create a new parent node with the combined hash and set its children
		Hash:  parentHash,
		Left:  left,
		Right: right,
	}
	left.Parent = parent
	right.Parent = parent

	return parent
}

// RootHash returns the hash of the root node of the Merkle Tree.
func (t *Tree) RootHash() []byte {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.root != nil {
		return t.root.Hash
	}
	return nil
}

func (t *Tree) Append(data []byte) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.indexMap == nil {
		t.indexMap = make(map[string][]int)
	}

	leafHash := HashLeafData(data, t.hashFunc)
	t.Leaves = append(t.Leaves, &Node{Hash: leafHash})

	hashHex := hex.EncodeToString(leafHash)
	t.indexMap[hashHex] = append(t.indexMap[hashHex], len(t.Leaves)-1)

	t.root = buildRecursive(t.Leaves, t.hashFunc)
	return nil
}

func (t *Tree) Print() {
	t.lock.RLock()
	root := t.root // Capture the root while under lock
	t.lock.RUnlock()

	printNode(root, "", true)
}

func printNode(n *Node, prefix string, isTail bool) {
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
		printNode(n.Right, newPrefix, false)
	}

	fmt.Printf("%s", prefix)
	if isTail {
		fmt.Printf("└── ")
	} else {
		fmt.Printf("┌── ")
	}
	fmt.Printf("%s\n", hashStr[:8]) // print first 8 chars

	if n.Left != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		printNode(n.Left, newPrefix, true)
	}
}
