package merkle

import (
	"testing"
)

func TestGenerateInclusionProof(t *testing.T) {
	tests := []struct {
		name     string
		data     [][]byte
		index    int
		wantErr  bool
		validate func(*InclusionProof) bool
	}{
		{
			name:    "first leaf in tree with two leaves",
			data:    [][]byte{[]byte("leaf1"), []byte("leaf2")},
			index:   0,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				return len(proof.Siblings) > 0 && len(proof.Left) == len(proof.Siblings)
			},
		},
		{
			name:    "second leaf in tree with two leaves",
			data:    [][]byte{[]byte("leaf1"), []byte("leaf2")},
			index:   1,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				return len(proof.Siblings) > 0 && len(proof.Left) == len(proof.Siblings)
			},
		},
		{
			name:    "single leaf tree",
			data:    [][]byte{[]byte("only")},
			index:   0,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				return len(proof.Siblings) == 0 && len(proof.Left) == 0
			},
		},
		{
			name:    "middle leaf in large tree",
			data:    [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
			index:   1,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				return len(proof.Siblings) > 0
			},
		},
		{
			name:    "last leaf in tree",
			data:    [][]byte{[]byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5")},
			index:   4,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				return len(proof.Siblings) > 0
			},
		},
		{
			name:    "invalid index negative",
			data:    [][]byte{[]byte("leaf1"), []byte("leaf2")},
			index:   -1,
			wantErr: true,
		},
		{
			name:    "invalid index out of bounds",
			data:    [][]byte{[]byte("leaf1"), []byte("leaf2")},
			index:   2,
			wantErr: true,
		},
		{
			name:    "invalid index at boundary",
			data:    [][]byte{[]byte("leaf1")},
			index:   1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := NewTree(tt.data, nil)
			if err != nil {
				t.Fatalf("Failed to create tree: %v", err)
			}

			proof, err := tree.GenerateInclusionProof(tt.index)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateInclusionProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.validate(proof) {
				t.Errorf("GenerateInclusionProof() validation failed")
			}
		})
	}
}

func TestGenerateInclusionProofByData(t *testing.T) {
	tests := []struct {
		name       string
		data       [][]byte
		searchData []byte
		wantErr    bool
		validate   func(*InclusionProof) bool
	}{
		{
			name:       "existing leaf in tree",
			data:       [][]byte{[]byte("hello"), []byte("world")},
			searchData: []byte("hello"),
			wantErr:    false,
			validate: func(proof *InclusionProof) bool {
				return proof != nil && len(proof.Left) == len(proof.Siblings)
			},
		},
		{
			name:       "second leaf",
			data:       [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")},
			searchData: []byte("bar"),
			wantErr:    false,
			validate: func(proof *InclusionProof) bool {
				return proof != nil
			},
		},
		{
			name:       "leaf not in tree",
			data:       [][]byte{[]byte("a"), []byte("b")},
			searchData: []byte("c"),
			wantErr:    true,
		},
		{
			name:       "empty search data",
			data:       [][]byte{[]byte(""), []byte("nonempty")},
			searchData: []byte(""),
			wantErr:    false,
			validate: func(proof *InclusionProof) bool {
				return proof != nil
			},
		},
		{
			name:       "duplicate leaves, returns first",
			data:       [][]byte{[]byte("dup"), []byte("dup"), []byte("other")},
			searchData: []byte("dup"),
			wantErr:    false,
			validate: func(proof *InclusionProof) bool {
				return proof != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := NewTree(tt.data, nil)
			if err != nil {
				t.Fatalf("Failed to create tree: %v", err)
			}

			proof, err := tree.GenerateInclusionProofByData(tt.searchData)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateInclusionProofByData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.validate(proof) {
				t.Errorf("GenerateInclusionProofByData() validation failed")
			}
		})
	}
}

func TestVerifyInclusionProof(t *testing.T) {
	tests := []struct {
		name         string
		treeData     [][]byte
		leafIndex    int
		verifyData   []byte
		hashFunc     HashFunc
		shouldVerify bool
		wantErr      bool
	}{
		{
			name:         "valid proof for first leaf",
			treeData:     [][]byte{[]byte("leaf1"), []byte("leaf2")},
			leafIndex:    0,
			verifyData:   []byte("leaf1"),
			hashFunc:     nil,
			shouldVerify: true,
			wantErr:      false,
		},
		{
			name:         "valid proof for second leaf",
			treeData:     [][]byte{[]byte("leaf1"), []byte("leaf2")},
			leafIndex:    1,
			verifyData:   []byte("leaf2"),
			hashFunc:     nil,
			shouldVerify: true,
			wantErr:      false,
		},
		{
			name:         "invalid proof with wrong data",
			treeData:     [][]byte{[]byte("leaf1"), []byte("leaf2")},
			leafIndex:    0,
			verifyData:   []byte("wrong"),
			hashFunc:     nil,
			shouldVerify: false,
			wantErr:      false,
		},
		{
			name:         "valid proof with custom hash func",
			treeData:     [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
			leafIndex:    2,
			verifyData:   []byte("c"),
			hashFunc:     DefaultHashFunc,
			shouldVerify: true,
			wantErr:      false,
		},
		{
			name:         "single leaf tree verification",
			treeData:     [][]byte{[]byte("single")},
			leafIndex:    0,
			verifyData:   []byte("single"),
			hashFunc:     nil,
			shouldVerify: true,
			wantErr:      false,
		},
		{
			name:         "odd number of leaves",
			treeData:     [][]byte{[]byte("1"), []byte("2"), []byte("3")},
			leafIndex:    1,
			verifyData:   []byte("2"),
			hashFunc:     nil,
			shouldVerify: true,
			wantErr:      false,
		},
		{
			name:         "large tree verification",
			treeData:     [][]byte{[]byte("0"), []byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5"), []byte("6"), []byte("7")},
			leafIndex:    3,
			verifyData:   []byte("3"),
			hashFunc:     nil,
			shouldVerify: true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := NewTree(tt.treeData, nil)
			if err != nil {
				t.Fatalf("Failed to create tree: %v", err)
			}

			rootHash := tree.RootHash()

			proof, err := tree.GenerateInclusionProof(tt.leafIndex)
			if err != nil {
				t.Fatalf("Failed to generate proof: %v", err)
			}

			result := VerifyInclusionProof(tt.verifyData, proof, rootHash, tt.hashFunc)

			if result != tt.shouldVerify {
				t.Errorf("VerifyInclusionProof() = %v, want %v", result, tt.shouldVerify)
			}
		})
	}
}

func TestVerifyInclusionProofConsistency(t *testing.T) {
	tests := []struct {
		name     string
		treeData [][]byte
	}{
		{
			name:     "two leaves",
			treeData: [][]byte{[]byte("a"), []byte("b")},
		},
		{
			name:     "four leaves",
			treeData: [][]byte{[]byte("1"), []byte("2"), []byte("3"), []byte("4")},
		},
		{
			name:     "five leaves",
			treeData: [][]byte{[]byte("x"), []byte("y"), []byte("z"), []byte("p"), []byte("q")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, _ := NewTree(tt.treeData, nil)
			rootHash := tree.RootHash()

			// Verify all leaves
			for i := 0; i < len(tt.treeData); i++ {
				proof, err := tree.GenerateInclusionProof(i)
				if err != nil {
					t.Fatalf("Failed to generate proof for index %d: %v", i, err)
				}

				if !VerifyInclusionProof(tt.treeData[i], proof, rootHash, nil) {
					t.Errorf("VerifyInclusionProof failed for leaf at index %d", i)
				}
			}
		})
	}
}

func TestInclusionProofStructure(t *testing.T) {
	tests := []struct {
		name                string
		treeData            [][]byte
		expectedProofLevels func(int) int
	}{
		{
			name:     "two leaves - one level",
			treeData: [][]byte{[]byte("a"), []byte("b")},
			expectedProofLevels: func(index int) int {
				return 1
			},
		},
		{
			name:     "four leaves - two levels",
			treeData: [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
			expectedProofLevels: func(index int) int {
				return 2
			},
		},
		{
			name:     "three leaves - one level",
			treeData: [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			expectedProofLevels: func(index int) int {
				if index == 2 {
					return 1 // Leaf "c" is higher up in the tree!
				}
				return 2 // Leaves "a" and "b" are deeper
			},
		},
		{
			name:     "five leaves - heavily unbalanced",
			treeData: [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")},
			expectedProofLevels: func(index int) int {
				if index == 4 {
					return 1 // Leaf "e" only needs the hash of the massive left subtree
				}
				return 3 // Leaves "a" through "d" are deeply nested in the left side
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, _ := NewTree(tt.treeData, nil)

			for i := 0; i < len(tt.treeData); i++ {
				proof, err := tree.GenerateInclusionProof(i)
				if err != nil {
					t.Fatalf("Failed to generate proof: %v", err)
				}

				expectedLevels := tt.expectedProofLevels(i)
				if len(proof.Siblings) != expectedLevels {
					t.Errorf("Index %d: proof levels = %d, want %d", i, len(proof.Siblings), expectedLevels)
				}

				// Verify siblings and left arrays have same length
				if len(proof.Siblings) != len(proof.Left) {
					t.Errorf("Index %d: siblings length %d != left length %d", i, len(proof.Siblings), len(proof.Left))
				}

				// Verify each sibling is not nil
				for j, sibling := range proof.Siblings {
					if sibling == nil {
						t.Errorf("Index %d, sibling %d is nil", i, j)
					}
				}
			}
		})
	}
}

func TestVerifyInclusionProofWithModifiedProof(t *testing.T) {
	tests := []struct {
		name              string
		treeData          [][]byte
		modifyProof       func(*InclusionProof)
		shouldStillVerify bool
	}{
		{
			name:     "modify sibling hash",
			treeData: [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
			modifyProof: func(proof *InclusionProof) {
				if len(proof.Siblings) > 0 {
					proof.Siblings[0][0] ^= 0xFF // flip bits in first sibling
				}
			},
			shouldStillVerify: false,
		},
		{
			name:     "modify left boolean",
			treeData: [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
			modifyProof: func(proof *InclusionProof) {
				if len(proof.Left) > 0 {
					proof.Left[0] = !proof.Left[0] // flip the direction
				}
			},
			shouldStillVerify: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, _ := NewTree(tt.treeData, nil)
			rootHash := tree.RootHash()

			proof, _ := tree.GenerateInclusionProof(0)
			tt.modifyProof(proof)

			result := VerifyInclusionProof(tt.treeData[0], proof, rootHash, nil)

			if result != tt.shouldStillVerify {
				t.Errorf("VerifyInclusionProof after modification = %v, want %v", result, tt.shouldStillVerify)
			}
		})
	}
}
