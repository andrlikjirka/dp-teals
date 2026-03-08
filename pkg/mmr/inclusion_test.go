package mmr

import (
	"bytes"
	"testing"
)

func TestGenerateInclusionProof_Table(t *testing.T) {
	tests := []struct {
		name     string
		leaves   [][]byte
		index    int
		wantErr  bool
		validate func(*InclusionProof) bool
	}{
		{
			name:    "single leaf MMR - no siblings needed",
			leaves:  [][]byte{[]byte("a")},
			index:   0,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				// Single leaf, single peak, no merging -> should have peak bagging for this peak only
				// Since there's only one peak and no other peaks, no siblings needed
				return len(proof.Siblings) == 0 && len(proof.Left) == 0
			},
		},
		{
			name:    "two leaves - merged to single peak",
			leaves:  [][]byte{[]byte("a"), []byte("b")},
			index:   0,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				// Two leaves merge to one peak of height 1
				// Proof from leaf to peak requires 1 sibling (the other leaf)
				return len(proof.Siblings) == 1 && len(proof.Left) == 1
			},
		},
		{
			name:    "two leaves - second leaf",
			leaves:  [][]byte{[]byte("a"), []byte("b")},
			index:   1,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				return len(proof.Siblings) == 1 && len(proof.Left) == 1
			},
		},
		{
			name:    "three leaves - creates two peaks",
			leaves:  [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			index:   0,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				// Leaf 0 is part of first peak (height 1, contains a+b)
				// Need 1 sibling to reach peak, then 1 sibling for the other peak
				return len(proof.Siblings) == 2 && len(proof.Siblings) == len(proof.Left)
			},
		},
		{
			name:    "three leaves - last leaf is standalone peak",
			leaves:  [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			index:   2,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				// Leaf 2 is a standalone peak (height 0)
				// Need 1 sibling for peak bagging (the merged peak of a+b)
				return len(proof.Siblings) == 1 && len(proof.Left) == 1
			},
		},
		{
			name:    "four leaves - single peak height 2",
			leaves:  [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
			index:   1,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				// All merge to single peak, need to climb to root
				return len(proof.Siblings) == 2 && len(proof.Left) == 2
			},
		},
		{
			name:    "five leaves - two peaks (height 2, height 0)",
			leaves:  [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")},
			index:   0,
			wantErr: false,
			validate: func(proof *InclusionProof) bool {
				// Leaf 0 is in first peak (height 2)
				// Need siblings to climb to peak + peak bagging sibling
				return len(proof.Siblings) >= 2 && len(proof.Siblings) == len(proof.Left)
			},
		},
		{
			name:    "invalid index - negative",
			leaves:  [][]byte{[]byte("a"), []byte("b")},
			index:   -1,
			wantErr: true,
			validate: func(proof *InclusionProof) bool {
				return true // not called when error expected
			},
		},
		{
			name:    "invalid index - out of bounds",
			leaves:  [][]byte{[]byte("a"), []byte("b")},
			index:   5,
			wantErr: true,
			validate: func(proof *InclusionProof) bool {
				return true // not called when error expected
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := buildMMRFromLeaves(t, tc.leaves)

			proof, err := m.GenerateInclusionProof(tc.index)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if proof == nil {
				t.Fatalf("proof is nil")
			}

			if !tc.validate(proof) {
				t.Fatalf("proof validation failed: siblings=%d, left=%d", len(proof.Siblings), len(proof.Left))
			}

			// Ensure siblings and left arrays match
			if len(proof.Siblings) != len(proof.Left) {
				t.Fatalf("siblings length %d != left length %d", len(proof.Siblings), len(proof.Left))
			}
		})
	}
}

func TestGenerateInclusionProofByData_Table(t *testing.T) {
	tests := []struct {
		name       string
		leaves     [][]byte
		searchData []byte
		wantErr    bool
	}{
		{
			name:       "find first leaf",
			leaves:     [][]byte{[]byte("hello"), []byte("world")},
			searchData: []byte("hello"),
			wantErr:    false,
		},
		{
			name:       "find middle leaf",
			leaves:     [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			searchData: []byte("b"),
			wantErr:    false,
		},
		{
			name:       "find last leaf",
			leaves:     [][]byte{[]byte("x"), []byte("y"), []byte("z")},
			searchData: []byte("z"),
			wantErr:    false,
		},
		{
			name:       "duplicate leaves - returns first occurrence proof",
			leaves:     [][]byte{[]byte("dup"), []byte("unique"), []byte("dup")},
			searchData: []byte("dup"),
			wantErr:    false,
		},
		{
			name:       "leaf not found",
			leaves:     [][]byte{[]byte("a"), []byte("b")},
			searchData: []byte("c"),
			wantErr:    true,
		},
		{
			name:       "empty search data",
			leaves:     [][]byte{[]byte("a"), []byte("b")},
			searchData: []byte{},
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := buildMMRFromLeaves(t, tc.leaves)

			proof, err := m.GenerateInclusionProofByData(tc.searchData)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if proof == nil {
				t.Fatalf("proof is nil")
			}

			// Verify proof structure is valid
			if len(proof.Siblings) != len(proof.Left) {
				t.Fatalf("siblings length %d != left length %d", len(proof.Siblings), len(proof.Left))
			}
		})
	}
}

func TestVerifyInclusionProof_Table(t *testing.T) {
	tests := []struct {
		name       string
		leaves     [][]byte
		leafIndex  int
		shouldPass bool
	}{
		{
			name:       "valid proof for first leaf in 2-leaf MMR",
			leaves:     [][]byte{[]byte("a"), []byte("b")},
			leafIndex:  0,
			shouldPass: true,
		},
		{
			name:       "valid proof for second leaf in 2-leaf MMR",
			leaves:     [][]byte{[]byte("a"), []byte("b")},
			leafIndex:  1,
			shouldPass: true,
		},
		{
			name:       "valid proof in 3-leaf MMR (two peaks)",
			leaves:     [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			leafIndex:  1,
			shouldPass: true,
		},
		{
			name:       "valid proof for standalone peak",
			leaves:     [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			leafIndex:  2,
			shouldPass: true,
		},
		{
			name:       "valid proof in larger MMR",
			leaves:     [][]byte{[]byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5")},
			leafIndex:  3,
			shouldPass: true,
		},
		{
			name:       "single leaf MMR",
			leaves:     [][]byte{[]byte("only")},
			leafIndex:  0,
			shouldPass: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := buildMMRFromLeaves(t, tc.leaves)
			root := m.RootHash()

			proof, err := m.GenerateInclusionProof(tc.leafIndex)
			if err != nil {
				t.Fatalf("failed to generate proof: %v", err)
			}

			leafData := tc.leaves[tc.leafIndex]
			valid := VerifyInclusionProof(leafData, proof, root, nil)

			if valid != tc.shouldPass {
				t.Fatalf("verification result = %v, want %v", valid, tc.shouldPass)
			}
		})
	}
}

func TestVerifyInclusionProofWithTamperedData_Table(t *testing.T) {
	tests := []struct {
		name        string
		leaves      [][]byte
		leafIndex   int
		modifyProof func(*InclusionProof)
	}{
		{
			name:      "tampered sibling hash",
			leaves:    [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			leafIndex: 0,
			modifyProof: func(proof *InclusionProof) {
				if len(proof.Siblings) > 0 {
					proof.Siblings[0][0] ^= 0xFF // flip bits
				}
			},
		},
		{
			name:      "flipped left boolean",
			leaves:    [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
			leafIndex: 1,
			modifyProof: func(proof *InclusionProof) {
				if len(proof.Left) > 0 {
					proof.Left[0] = !proof.Left[0]
				}
			},
		},
		{
			name:      "removed sibling",
			leaves:    [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			leafIndex: 0,
			modifyProof: func(proof *InclusionProof) {
				if len(proof.Siblings) > 1 {
					proof.Siblings = proof.Siblings[:len(proof.Siblings)-1]
					proof.Left = proof.Left[:len(proof.Left)-1]
				}
			},
		},
		{
			name:      "added extra sibling",
			leaves:    [][]byte{[]byte("a"), []byte("b")},
			leafIndex: 0,
			modifyProof: func(proof *InclusionProof) {
				proof.Siblings = append(proof.Siblings, make([]byte, 32))
				proof.Left = append(proof.Left, false)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := buildMMRFromLeaves(t, tc.leaves)
			root := m.RootHash()

			proof, err := m.GenerateInclusionProof(tc.leafIndex)
			if err != nil {
				t.Fatalf("failed to generate proof: %v", err)
			}

			// Tamper with the proof
			tc.modifyProof(proof)

			leafData := tc.leaves[tc.leafIndex]
			valid := VerifyInclusionProof(leafData, proof, root, nil)

			if valid {
				t.Fatalf("tampered proof should not verify, but it did")
			}
		})
	}
}

func TestVerifyInclusionProofWithWrongLeafData(t *testing.T) {
	leaves := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	m := buildMMRFromLeaves(t, leaves)
	root := m.RootHash()

	proof, err := m.GenerateInclusionProof(0)
	if err != nil {
		t.Fatalf("failed to generate proof: %v", err)
	}

	// Try to verify with wrong leaf data
	wrongData := []byte("wrong")
	valid := VerifyInclusionProof(wrongData, proof, root, nil)

	if valid {
		t.Fatalf("proof with wrong leaf data should not verify")
	}
}

func TestVerifyInclusionProofWithWrongRoot(t *testing.T) {
	leaves := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	m := buildMMRFromLeaves(t, leaves)

	proof, err := m.GenerateInclusionProof(0)
	if err != nil {
		t.Fatalf("failed to generate proof: %v", err)
	}

	// Create a different MMR to get a different root
	m2 := buildMMRFromLeaves(t, [][]byte{[]byte("x"), []byte("y"), []byte("z")})
	wrongRoot := m2.RootHash()

	valid := VerifyInclusionProof(leaves[0], proof, wrongRoot, nil)

	if valid {
		t.Fatalf("proof with wrong root should not verify")
	}
}

func TestInclusionProofStructure_Table(t *testing.T) {
	tests := []struct {
		name             string
		leaves           [][]byte
		leafIndex        int
		wantSiblingsLen  int
		validateSiblings func([][]byte, []bool) bool
	}{
		{
			name:            "single leaf - no siblings",
			leaves:          [][]byte{[]byte("a")},
			leafIndex:       0,
			wantSiblingsLen: 0,
			validateSiblings: func(siblings [][]byte, left []bool) bool {
				return len(siblings) == 0 && len(left) == 0
			},
		},
		{
			name:            "two leaves - one sibling",
			leaves:          [][]byte{[]byte("a"), []byte("b")},
			leafIndex:       0,
			wantSiblingsLen: 1,
			validateSiblings: func(siblings [][]byte, left []bool) bool {
				return len(siblings) == 1 && siblings[0] != nil
			},
		},
		{
			name:            "four leaves - two siblings to reach root",
			leaves:          [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
			leafIndex:       0,
			wantSiblingsLen: 2,
			validateSiblings: func(siblings [][]byte, left []bool) bool {
				// Verify all siblings are non-nil
				for _, sib := range siblings {
					if sib == nil {
						return false
					}
				}
				return true
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := buildMMRFromLeaves(t, tc.leaves)

			proof, err := m.GenerateInclusionProof(tc.leafIndex)
			if err != nil {
				t.Fatalf("failed to generate proof: %v", err)
			}

			if len(proof.Siblings) != tc.wantSiblingsLen {
				t.Fatalf("siblings count = %d, want %d", len(proof.Siblings), tc.wantSiblingsLen)
			}

			if !tc.validateSiblings(proof.Siblings, proof.Left) {
				t.Fatalf("sibling validation failed")
			}
		})
	}
}

func TestInclusionProofConsistency(t *testing.T) {
	// Test that generating proof multiple times gives same result
	leaves := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")}
	m := buildMMRFromLeaves(t, leaves)

	proof1, err1 := m.GenerateInclusionProof(1)
	proof2, err2 := m.GenerateInclusionProof(1)

	if err1 != nil || err2 != nil {
		t.Fatalf("failed to generate proofs: %v, %v", err1, err2)
	}

	if len(proof1.Siblings) != len(proof2.Siblings) {
		t.Fatalf("proof siblings length mismatch: %d vs %d", len(proof1.Siblings), len(proof2.Siblings))
	}

	for i := range proof1.Siblings {
		if !bytes.Equal(proof1.Siblings[i], proof2.Siblings[i]) {
			t.Fatalf("proof sibling %d mismatch", i)
		}
		if proof1.Left[i] != proof2.Left[i] {
			t.Fatalf("proof left[%d] mismatch", i)
		}
	}
}

func TestInclusionProofByDataWithDuplicates(t *testing.T) {
	// Test that duplicate leaves correctly map to first occurrence
	leaves := [][]byte{[]byte("dup"), []byte("unique"), []byte("dup"), []byte("other")}
	m := buildMMRFromLeaves(t, leaves)

	proof, err := m.GenerateInclusionProofByData([]byte("dup"))
	if err != nil {
		t.Fatalf("failed to generate proof: %v", err)
	}

	// Verify the proof is valid for the first occurrence
	root := m.RootHash()
	valid := VerifyInclusionProof([]byte("dup"), proof, root, nil)

	if !valid {
		t.Fatalf("proof for duplicate leaf should verify")
	}
}
