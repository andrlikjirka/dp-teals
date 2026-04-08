package model

// MmrNode represents a node in the Merkle Mountain Range (MMR) structure, which can be either a leaf or an internal node.
type MmrNode struct {
	ID           int64  `db:"id"`
	LeafIndex    *int64 `db:"leaf_index"`     // nullable for internal nodes
	LeftChildID  *int64 `db:"left_child_id"`  // nullable for leaf nodes
	RightChildID *int64 `db:"right_child_id"` // nullable for leaf nodes
	ParentID     *int64 `db:"parent_id"`      // nullable for root node
	Hash         []byte `db:"hash"`
	Level        int    `db:"level"`
}
