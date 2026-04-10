package repository

import (
	"context"
	"fmt"
	"math/bits"

	"github.com/andrlikjirka/dp-teals/pkg/hash"
	"github.com/andrlikjirka/dp-teals/pkg/merkle"
	"github.com/andrlikjirka/dp-teals/pkg/mmr"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository/sql/query"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/georgysavva/scany/v2/pgxscan"
)

// LedgerRepository manages the Merkle Mountain Range (MMR) ledger, allowing appending of new leaf nodes and maintaining the MMR structure in the database.
type LedgerRepository struct {
	db       sql.Db
	hashFunc hash.Func
}

// NewLedgerRepository creates a new instance of LedgerRepository with the provided database connection and hash function. The hash function is used to compute the hashes for the MMR nodes, and the database connection is used to persist the MMR structure.
func NewLedgerRepository(db sql.Db, hashFunc hash.Func) *LedgerRepository {
	if hashFunc == nil {
		hashFunc = hash.DefaultHashFunc
	}

	return &LedgerRepository{
		db:       db,
		hashFunc: hashFunc,
	}
}

// RootHash computes the current root hash of the MMR ledger by retrieving all the current peaks from the database and combining their hashes from right to left. If there are no peaks (i.e., the MMR is empty), it returns nil. If there is an error during the retrieval of peaks, it wraps and returns the error.
func (r *LedgerRepository) RootHash(ctx context.Context) (rootHash []byte, err error) {
	var peaks []model.MmrNode
	err = pgxscan.Select(ctx, r.db, &peaks, query.GetMmrPeaks)
	if err != nil {
		return nil, fmt.Errorf("get current peaks: %w", err)
	}
	if len(peaks) == 0 {
		return nil, nil // empty MMR has no root
	}

	root := peaks[len(peaks)-1].Hash // start with the rightmost peak
	for i := len(peaks) - 2; i >= 0; i-- {
		root = merkle.HashInternalNodes(peaks[i].Hash, root, r.hashFunc) // combine peaks from right to left
	}
	return root, nil
}

// Size retrieves the current number of leaves in the MMR ledger. It executes a query to count the number of leaf nodes in the database and returns that count. If there is an error during the query execution, it wraps and returns the error.
func (r *LedgerRepository) Size(ctx context.Context) (size int64, err error) {
	var count int64
	err = pgxscan.Get(ctx, r.db, &count, query.GetMmrSize)
	if err != nil {
		return 0, fmt.Errorf("get mmr size: %w", err)
	}
	return count, nil
}

// --- APPEND LEAF ---

// AppendLeaf adds a new leaf node to the MMR ledger with the given payload.
func (r *LedgerRepository) AppendLeaf(ctx context.Context, payload []byte) (nodeID int64, err error) {
	if len(payload) == 0 {
		return 0, fmt.Errorf("payload cannot be empty")
	}

	// 1. Hash the audit event payload to create the leaf node hash
	leafHash := merkle.HashLeafData(payload, r.hashFunc)

	// 2. Get the current size of the MMR to determine the new leaf index
	size, err := r.Size(ctx)
	if err != nil {
		return 0, fmt.Errorf("determine next leaf index: %w", err)
	}

	// 3. Insert the new leaf node into the ledger and get its ID
	leaf := &model.MmrNode{
		Hash:      leafHash,
		Level:     0,
		LeafIndex: &size,
	}
	if err := r.insertNode(ctx, leaf); err != nil {
		return 0, fmt.Errorf("insert leaf node: %w", err)
	}
	leafNodeID := leaf.ID // populated by insertNode (returning the ID of the newly inserted node)

	// 4. Peeks merge loop
	currentID := leafNodeID
	currentLevel := 0
	currentHash := leafHash
	for {
		peak, err := r.getRightmostPeakAtLevel(ctx, currentLevel, currentID)
		if err != nil {
			return 0, fmt.Errorf("get peak at level %d: %w", currentLevel, err)
		}
		if peak == nil {
			break // no merge possible
		}

		mergedHash := merkle.HashInternalNodes(peak.Hash, currentHash, r.hashFunc)
		newNode := &model.MmrNode{
			Hash:         mergedHash,
			Level:        currentLevel + 1,
			LeftChildID:  &peak.ID,
			RightChildID: &currentID,
		}
		if err := r.insertNode(ctx, newNode); err != nil {
			return 0, fmt.Errorf("insert internal node: %w", err)
		}

		mergedID := newNode.ID
		if err := r.setParent(ctx, mergedID, peak.ID, currentID); err != nil {
			return 0, fmt.Errorf("set parent for node %d: %w", mergedID, err)
		}

		currentID = mergedID
		currentLevel++
		currentHash = mergedHash
	}

	return leafNodeID, nil
}

// insertNode inserts a new MMR node into the database and updates the node's ID field with the generated ID from the database. It takes a context and a pointer to an MmrNode struct, executes the insert query, and returns any error that occurs during the operation.
func (r *LedgerRepository) insertNode(ctx context.Context, node *model.MmrNode) error {
	return pgxscan.Get(ctx, r.db, &node.ID, query.InsertMmrNode,
		node.LeafIndex, node.LeftChildID, node.RightChildID, node.Hash, node.Level,
	)
}

// getRightmostPeakAtLevel retrieves the rightmost peak node at a given level, excluding a specific node ID (the one being merged). It executes a query to find the peak node at the specified level that is not the current node. If no such peak exists, it returns nil, indicating that no merge is possible at this level. If an error occurs during the query execution, it wraps and returns the error.
func (r *LedgerRepository) getRightmostPeakAtLevel(ctx context.Context, level int, excludeID int64) (*model.MmrNode, error) {
	var node model.MmrNode
	err := pgxscan.Get(ctx, r.db, &node, query.GetRightmostPeakAtLevel, level, excludeID)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, nil // no peak at this level — merge loop should stop
		}
		return nil, fmt.Errorf("get rightmost peak at level %d: %w", level, err)
	}
	return &node, nil
}

// setParent updates the parent ID of the left and right child nodes to point to the newly created parent node. It executes an update query to set the parent ID for both child nodes. If the number of affected rows is not 2 (indicating that both child nodes were not updated), it returns an error. If any error occurs during the query execution, it wraps and returns the error.
func (r *LedgerRepository) setParent(ctx context.Context, parentID, leftChildID, rightChildID int64) error {
	tag, err := r.db.Exec(ctx, query.SetMmrNodeParent, parentID, leftChildID, rightChildID)
	if err != nil {
		return fmt.Errorf("set parent: %w", err)
	}
	if tag.RowsAffected() != 2 {
		return fmt.Errorf("set parent: expected 2 rows affected, got %d", tag.RowsAffected())
	}
	return nil
}

// --- INCLUSION PROOF ---

// GenerateInclusionProof generates an inclusion proof for the leaf at the specified index in the MMR ledger. It first retrieves the path from the leaf to its peak, collecting sibling node IDs and their positions (left/right). Then it fetches the sibling hashes in a single query. Finally, it performs peak bagging by combining the peaks to the right and left of the leaf's peak to construct the full inclusion proof. If any error occurs during these steps, it wraps and returns the error.
func (r *LedgerRepository) GenerateInclusionProof(ctx context.Context, leafIndex int64) (proof *svcmodel.InclusionProofData, err error) {
	// phase 1: traverse from leaf up to its peak
	var path []model.MmrNode
	err = pgxscan.Select(ctx, r.db, &path, query.GetLeafToPeakPath, leafIndex)
	if err != nil {
		return nil, fmt.Errorf("get leaf path: %w", err)
	}
	if len(path) == 0 {
		return nil, fmt.Errorf("leaf at index %d not found", leafIndex)
	}

	var siblingIDs []int64
	var siblingLeft []bool

	// collect sibling IDs and their positions (left/right)
	for i := 0; i < len(path)-1; i++ {
		current, parent := path[i], path[i+1]
		if *parent.LeftChildID == current.ID {
			siblingIDs = append(siblingIDs, *parent.RightChildID)
			siblingLeft = append(siblingLeft, false) // sibling is on the right
		} else {
			siblingIDs = append(siblingIDs, *parent.LeftChildID)
			siblingLeft = append(siblingLeft, true) // sibling is on the left
		}
	}

	// fetch sibling hashes in one query
	siblingHashes, err := r.getNodeHashes(ctx, siblingIDs)
	if err != nil {
		return nil, fmt.Errorf("get sibling hashes: %w", err)
	}
	var siblings [][]byte
	for _, id := range siblingIDs {
		siblings = append(siblings, siblingHashes[id])
	}

	// phase 2: peak bagging
	peak := path[len(path)-1]
	var allPeaks []model.MmrNode
	if err := pgxscan.Select(ctx, r.db, &allPeaks, query.GetMmrPeaks); err != nil {
		return nil, fmt.Errorf("get peaks: %w", err)
	}
	peakIdx := -1
	for i, p := range allPeaks {
		if p.ID == peak.ID {
			peakIdx = i
			break
		}
	}
	if peakIdx == -1 {
		return nil, fmt.Errorf("peak not found (internal state error)")
	}
	if peakIdx < len(allPeaks)-1 {
		rightBag := r.bagPeaksRightToLeft(allPeaks[peakIdx+1:])
		siblings = append(siblings, rightBag)
		siblingLeft = append(siblingLeft, false)
	}
	for i := peakIdx - 1; i >= 0; i-- {
		siblings = append(siblings, allPeaks[i].Hash)
		siblingLeft = append(siblingLeft, true)
	}

	rootHash := r.bagPeaksRightToLeft(allPeaks)
	treeSize, err := r.Size(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tree size: %w", err)
	}

	return &svcmodel.InclusionProofData{
		LeafIndex:  leafIndex,
		LedgerSize: treeSize,
		LeafHash:   path[0].Hash,
		RootHash:   rootHash,
		Proof:      &mmr.InclusionProof{Siblings: siblings, Left: siblingLeft},
	}, nil
}

// getNodeHashes retrieves the hashes for a list of node IDs and returns a map of node ID to hash. It executes a query to fetch the nodes by their IDs, and constructs a map from the results. If there are no IDs provided, it returns an empty map. If an error occurs during the query execution, it wraps and returns the error.
func (r *LedgerRepository) getNodeHashes(ctx context.Context, ids []int64) (map[int64][]byte, error) {
	if len(ids) == 0 {
		return map[int64][]byte{}, nil
	}

	var nodes []model.MmrNode
	if err := pgxscan.Select(ctx, r.db, &nodes, query.GetNodesByIDs, ids); err != nil {
		return nil, fmt.Errorf("get nodes by ids: %w", err)
	}
	result := make(map[int64][]byte, len(nodes))
	for _, n := range nodes {
		result[n.ID] = n.Hash
	}
	return result, nil
}

// bagPeaksRightToLeft is a helper method that takes a slice of peaks and combines them into a single hash by hashing from right to left. It starts with the rightmost peak's hash and iteratively combines it with the next peak to the left until all peaks are combined into a single root hash. If there are no peaks, it returns nil. If there is only one peak, it returns that peak's hash directly.
func (r *LedgerRepository) bagPeaksRightToLeft(peaks []model.MmrNode) []byte {
	if len(peaks) == 0 {
		return nil
	}
	if len(peaks) == 1 {
		return peaks[0].Hash
	}
	root := peaks[len(peaks)-1].Hash
	for i := len(peaks) - 2; i >= 0; i-- {
		root = merkle.HashInternalNodes(peaks[i].Hash, root, r.hashFunc)
	}
	return root
}

// --- CONSISTENCY PROOF ---

// GenerateConsistencyProof generates a consistency proof that demonstrates the growth of the MMR ledger from a previous size (fromSize) to a new size (toSize). It first validates the input sizes and retrieves the peaks for both sizes. Then it constructs the consistency paths for each old peak, which show how the old peaks evolve into the new peaks. Finally, it collects any new peaks that are strictly to the right of the old peaks. If any error occurs during these steps, it wraps and returns the error.
func (r *LedgerRepository) GenerateConsistencyProof(ctx context.Context, fromSize int64, toSize int64) (proof *mmr.ConsistencyProof, err error) {
	currentSize, err := r.Size(ctx)
	if err != nil {
		return nil, fmt.Errorf("get mmr size: %w", err)
	}

	if fromSize < 0 || toSize < 0 || fromSize > toSize || toSize > currentSize {
		return nil, svcerrors.ErrInvalidConsistencyProofRange
	}

	proof = &mmr.ConsistencyProof{
		OldSize:          int(fromSize),
		NewSize:          int(toSize),
		OldPeaksHashes:   make([][]byte, 0),
		ConsistencyPaths: make([]*mmr.ConsistencyPath, 0),
		RightPeaks:       make([][]byte, 0),
	}

	if fromSize == toSize {
		return proof, nil // trivial: tree hasn't grown
	}

	oldPeaks, err := r.getPeaksAtSize(ctx, fromSize)
	if err != nil {
		return nil, fmt.Errorf("get old peaks: %w", err)
	}
	newPeaks, err := r.getPeaksAtSize(ctx, toSize)
	if err != nil {
		return nil, fmt.Errorf("get new peaks: %w", err)
	}

	// Collect the hashes of the old peaks for the proof
	for _, p := range oldPeaks {
		proof.OldPeaksHashes = append(proof.OldPeaksHashes, p.Hash)
	}

	newPeakIDs := make(map[int64]int, len(newPeaks))
	for i, p := range newPeaks {
		newPeakIDs[p.ID] = i
	}
	lastNewPeakIdx := -1

	// Build the consistency paths for each old peak
	for _, oldPeak := range oldPeaks {
		path, newPeakIdx, err := r.buildConsistencyPath(ctx, oldPeak.ID, newPeakIDs)
		if err != nil {
			return nil, fmt.Errorf("build consistency path for peak %d: %w", oldPeak.ID, err)
		}
		proof.ConsistencyPaths = append(proof.ConsistencyPaths, path)
		if newPeakIdx > lastNewPeakIdx {
			lastNewPeakIdx = newPeakIdx
		}
	}

	// Collect the right-peaks (any new peaks strictly to the right of the nodes affected by the old peaks)
	for i := lastNewPeakIdx + 1; i < len(newPeaks); i++ {
		proof.RightPeaks = append(proof.RightPeaks, newPeaks[i].Hash)
	}

	return proof, err
}

// getPeaksAtSize retrieves the peaks of the MMR ledger at a specific size. It calculates which peaks correspond to the given size by analyzing the binary representation of the size and retrieving the appropriate ancestor nodes from the database. If the size is zero or negative, it returns an empty slice. If there is an error during the retrieval of peaks, it wraps and returns the error.
func (r *LedgerRepository) getPeaksAtSize(ctx context.Context, size int64) ([]model.MmrNode, error) {
	if size <= 0 {
		return nil, nil
	}

	var peaks []model.MmrNode
	var offset int64

	bitLen := bits.Len(uint(size))
	for bit := bitLen - 1; bit >= 0; bit-- {
		if size&(1<<bit) != 0 {
			// rightmost leaf index in this 2^bit subtree
			leafIdx := offset + (1 << bit) - 1
			targetLevel := bit

			var peak model.MmrNode
			if targetLevel == 0 {
				if err := pgxscan.Get(ctx, r.db, &peak, query.GetAncestorAtLevel, leafIdx, 0); err != nil {
					return nil, fmt.Errorf("get leaf peak at index %d: %w", leafIdx, err)
				}
			} else {
				if err := pgxscan.Get(ctx, r.db, &peak, query.GetAncestorAtLevel, leafIdx, targetLevel); err != nil {
					return nil, fmt.Errorf("get ancestor at level %d for leaf %d: %w", targetLevel, leafIdx, err)
				}
			}
			peaks = append(peaks, peak)
			offset += 1 << bit
		}
	}
	return peaks, nil
}

// buildConsistencyPath constructs the consistency path from an old peak to the closest new peak in the MMR ledger. It traverses from the old peak up to the new peak, collecting sibling node IDs and their positions (left/right). Then it fetches the sibling hashes in a single query and constructs the ConsistencyPath struct. It also returns the index of the new peak that this path leads to. If any error occurs during these steps, it wraps and returns the error.
func (r *LedgerRepository) buildConsistencyPath(ctx context.Context, oldPeakID int64, newPeakIDs map[int64]int) (*mmr.ConsistencyPath, int, error) {
	peakIDslice := make([]int64, 0, len(newPeakIDs))
	for id := range newPeakIDs {
		peakIDslice = append(peakIDslice, id)
	}
	var path []model.MmrNode
	err := pgxscan.Select(ctx, r.db, &path, query.GetPathToClosestNewPeak, oldPeakID, peakIDslice)
	if err != nil {
		return nil, -1, fmt.Errorf("get consistency path from node %d: %w", oldPeakID, err)
	}
	if len(path) == 0 {
		return nil, -1, fmt.Errorf("empty path from node %d", oldPeakID)
	}

	newPeakNode := path[len(path)-1] // last node is the new peak the traversal stopped at
	newPeakIdx, ok := newPeakIDs[newPeakNode.ID]
	if !ok {
		return nil, -1, fmt.Errorf("traversal ended at node %d which is not a new peak (internal state error)", newPeakNode.ID)
	}

	var siblingIDs []int64
	var left []bool
	for i := 0; i < len(path)-1; i++ {
		child, parent := path[i], path[i+1]
		if parent.LeftChildID != nil && *parent.LeftChildID == child.ID { // child is the left child → sibling is on the right
			siblingIDs = append(siblingIDs, *parent.RightChildID)
			left = append(left, false)
		} else { // child is the right child → sibling is on the left
			siblingIDs = append(siblingIDs, *parent.LeftChildID)
			left = append(left, true)
		}
	}

	hashMap, err := r.getNodeHashes(ctx, siblingIDs)
	if err != nil {
		return nil, -1, fmt.Errorf("get sibling hashes for consistency path: %w", err)
	}
	siblings := make([][]byte, len(siblingIDs))
	for i, id := range siblingIDs {
		siblings[i] = hashMap[id]
	}

	return &mmr.ConsistencyPath{Siblings: siblings, Left: left}, newPeakIdx, nil
}
