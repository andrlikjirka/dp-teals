package merkle

import (
	"testing"
)

// TestLargestPowerOfTwoLessThan ensures the bitwise math exactly matches RFC 6962 split boundaries
func TestLargestPowerOfTwoLessThan(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{n: 2, want: 1},
		{n: 3, want: 2},
		{n: 4, want: 2},
		{n: 5, want: 4},
		{n: 6, want: 4},
		{n: 7, want: 4},
		{n: 8, want: 4},
		{n: 9, want: 8},
		{n: 16, want: 8},
		{n: 17, want: 16},
	}

	for _, tt := range tests {
		got := largestPowerOfTwoLessThan(tt.n)
		if got != tt.want {
			t.Errorf("largestPowerOfTwoLessThan(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestGenerateConsistencyProof_Errors(t *testing.T) {
	data := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	tree, _ := NewTree(data, nil)

	tests := []struct {
		name string
		m    int
	}{
		{"m is zero", 0},
		{"m is negative", -1},
		{"m is larger than tree", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tree.GenerateConsistencyProof(tt.m)
			if err == nil {
				t.Errorf("GenerateConsistencyProof(%d) expected error, got nil", tt.m)
			}
		})
	}
}

// TestConsistencyProof_Standard valid mathematical generation and verification
func TestConsistencyProof_Standard(t *testing.T) {
	allData := [][]byte{
		[]byte("leaf1"), []byte("leaf2"), []byte("leaf3"),
		[]byte("leaf4"), []byte("leaf5"), []byte("leaf6"),
	}

	// We test trees growing from size m to size n
	tests := []struct {
		name string
		m    int
		n    int
	}{
		{"1 to 2", 1, 2},
		{"1 to 3", 1, 3},
		{"2 to 3", 2, 3},
		{"2 to 4", 2, 4},
		{"3 to 4", 3, 4},
		{"3 to 5", 3, 5},
		{"4 to 6", 4, 6},
		{"same size (3 to 3)", 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the older tree (size m)
			oldTree, _ := NewTree(allData[:tt.m], nil)
			oldRoot := oldTree.RootHash()

			// Build the newer tree (size n)
			newTree, _ := NewTree(allData[:tt.n], nil)
			newRoot := newTree.RootHash()

			// Generate the proof from the new tree
			proof, err := newTree.GenerateConsistencyProof(tt.m)
			if err != nil {
				t.Fatalf("Failed to generate proof: %v", err)
			}

			// Verify the proof
			valid := VerifyConsistencyProof(tt.m, tt.n, oldRoot, newRoot, proof, nil)
			if !valid {
				t.Errorf("VerifyConsistencyProof returned false for m=%d, n=%d", tt.m, tt.n)
			}
		})
	}
}

// TestConsistencyProof_Tampering checks against malicious proofs
func TestConsistencyProof_Tampering(t *testing.T) {
	data := [][]byte{[]byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5")}
	oldTree, _ := NewTree(data[:3], nil) // m = 3
	newTree, _ := NewTree(data, nil)     // n = 5

	oldRoot := oldTree.RootHash()
	newRoot := newTree.RootHash()

	t.Run("tampered hash", func(t *testing.T) {
		proof, _ := newTree.GenerateConsistencyProof(3)

		// Flip a bit in the first proof hash
		if len(proof.Hashes) > 0 {
			proof.Hashes[0][0] ^= 0xFF
		}

		valid := VerifyConsistencyProof(3, 5, oldRoot, newRoot, proof, nil)
		if valid {
			t.Error("VerifyConsistencyProof passed with a tampered hash")
		}
	})

	t.Run("bloated proof", func(t *testing.T) {
		proof, _ := newTree.GenerateConsistencyProof(3)

		// Add an extra, unneeded hash to the end of the proof array
		proof.Hashes = append(proof.Hashes, []byte("extra_fake_hash_data"))

		valid := VerifyConsistencyProof(3, 5, oldRoot, newRoot, proof, nil)
		if valid {
			t.Error("VerifyConsistencyProof passed with extra unused hashes (bloated proof)")
		}
	})

	t.Run("truncated proof", func(t *testing.T) {
		proof, _ := newTree.GenerateConsistencyProof(3)

		// Remove the last hash from the proof
		if len(proof.Hashes) > 0 {
			proof.Hashes = proof.Hashes[:len(proof.Hashes)-1]
		}

		valid := VerifyConsistencyProof(3, 5, oldRoot, newRoot, proof, nil)
		if valid {
			t.Error("VerifyConsistencyProof passed with missing hashes (truncated proof)")
		}
	})

	t.Run("wrong roots", func(t *testing.T) {
		proof, _ := newTree.GenerateConsistencyProof(3)
		fakeRoot := []byte("this_is_not_the_real_root_hash!")

		validOld := VerifyConsistencyProof(3, 5, fakeRoot, newRoot, proof, nil)
		if validOld {
			t.Error("VerifyConsistencyProof passed with fake old root")
		}

		validNew := VerifyConsistencyProof(3, 5, oldRoot, fakeRoot, proof, nil)
		if validNew {
			t.Error("VerifyConsistencyProof passed with fake new root")
		}
	})
}

// TestConsistencyProof_ContinuousAppend simulates a live, growing log
func TestConsistencyProof_ContinuousAppend(t *testing.T) {
	// Start with a 1-leaf tree
	tree, _ := NewTree([][]byte{[]byte("tx0")}, nil)

	// Keep a history of roots as the tree grows
	history := [][]byte{tree.RootHash()}

	// Append 15 more items, one by one
	for i := 1; i <= 15; i++ {
		// Append to the live tree
		newData := []byte{byte('t'), byte('x'), byte(i)}
		err := tree.Append(newData)
		if err != nil {
			t.Fatalf("Append failed at step %d: %v", i, err)
		}

		newRoot := tree.RootHash()
		n := len(tree.Leaves) // current size

		// Verify consistency against EVERY previous state in history
		for m, oldRoot := range history {
			treeSize := m + 1 // history index 0 means tree size 1

			proof, err := tree.GenerateConsistencyProof(treeSize)
			if err != nil {
				t.Fatalf("GenerateConsistencyProof failed for m=%d, n=%d: %v", treeSize, n, err)
			}

			valid := VerifyConsistencyProof(treeSize, n, oldRoot, newRoot, proof, nil)
			if !valid {
				t.Errorf("Continuous validation failed! The tree at size %d is NOT consistent with historic size %d", n, treeSize)
			}
		}

		// Add the new valid root to history for the next iteration
		history = append(history, newRoot)
	}
}
