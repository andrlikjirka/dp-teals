package mmr

import (
	"fmt"
	"testing"

	"github.com/andrlikjirka/merkle"
)

func TestGenerateConsistencyProof_Table(t *testing.T) {
	leaves := [][]byte{
		[]byte("a"), []byte("b"), []byte("c"), []byte("d"),
		[]byte("e"), []byte("f"), []byte("g"), []byte("h"),
	}

	tests := []struct {
		name       string
		oldSize    int
		newSize    int
		wantErr    bool
		wantOldLen int
		wantPathEq bool
	}{
		{
			name:       "equal sizes returns trivial proof",
			oldSize:    4,
			newSize:    4,
			wantErr:    false,
			wantOldLen: 0,
			wantPathEq: true,
		},
		{
			name:       "growth from 1 to 2",
			oldSize:    1,
			newSize:    2,
			wantErr:    false,
			wantOldLen: 1,
			wantPathEq: true,
		},
		{
			name:       "growth from 3 to 7",
			oldSize:    3,
			newSize:    7,
			wantErr:    false,
			wantOldLen: 2,
			wantPathEq: true,
		},
		{
			name:       "growth from 5 to 8",
			oldSize:    5,
			newSize:    8,
			wantErr:    false,
			wantOldLen: 2,
			wantPathEq: true,
		},
		{
			name:    "invalid old greater than new",
			oldSize: 5,
			newSize: 4,
			wantErr: true,
		},
		{
			name:    "invalid negative old size",
			oldSize: -1,
			newSize: 3,
			wantErr: true,
		},
		{
			name:    "invalid new size above current",
			oldSize: 2,
			newSize: 9,
			wantErr: true,
		},
	}

	m := buildMMRFromLeaves(t, leaves)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			proof, err := m.GenerateConsistencyProof(tc.oldSize, tc.newSize)
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
			if proof.OldSize != tc.oldSize || proof.NewSize != tc.newSize {
				t.Fatalf("proof sizes = (%d,%d), want (%d,%d)", proof.OldSize, proof.NewSize, tc.oldSize, tc.newSize)
			}
			if got := len(proof.OldPeaksHashes); got != tc.wantOldLen {
				t.Fatalf("len(OldPeaksHashes) = %d, want %d", got, tc.wantOldLen)
			}
			if tc.wantPathEq && len(proof.ConsistencyPaths) != len(proof.OldPeaksHashes) {
				t.Fatalf("len(ConsistencyPaths) = %d, want %d", len(proof.ConsistencyPaths), len(proof.OldPeaksHashes))
			}
			for i, p := range proof.ConsistencyPaths {
				if p == nil {
					t.Fatalf("ConsistencyPaths[%d] is nil", i)
				}
				if len(p.Siblings) != len(p.Left) {
					t.Fatalf("path[%d] siblings/left mismatch: %d/%d", i, len(p.Siblings), len(p.Left))
				}
			}
		})
	}
}

func TestVerifyConsistencyProof_Table(t *testing.T) {
	leaves := [][]byte{
		[]byte("a"), []byte("b"), []byte("c"), []byte("d"),
		[]byte("e"), []byte("f"), []byte("g"),
	}
	m := buildMMRFromLeaves(t, leaves)

	type pair struct {
		oldSize int
		newSize int
	}
	pairs := []pair{
		{oldSize: 1, newSize: 2},
		{oldSize: 2, newSize: 3},
		{oldSize: 3, newSize: 7},
		{oldSize: 4, newSize: 7},
		{oldSize: 5, newSize: 7},
		{oldSize: 7, newSize: 7},
	}

	for _, p := range pairs {
		p := p
		t.Run(fmt.Sprintf("old_%d_new_%d", p.oldSize, p.newSize), func(t *testing.T) {
			proof, err := m.GenerateConsistencyProof(p.oldSize, p.newSize)
			if err != nil {
				t.Fatalf("GenerateConsistencyProof(%d,%d) error: %v", p.oldSize, p.newSize, err)
			}

			oldRoot := buildMMRFromLeaves(t, leaves[:p.oldSize]).RootHash()
			newRoot := buildMMRFromLeaves(t, leaves[:p.newSize]).RootHash()

			if ok := VerifyConsistencyProof(proof, oldRoot, newRoot, nil); !ok {
				t.Fatalf("VerifyConsistencyProof() = false, want true for (%d,%d)", p.oldSize, p.newSize)
			}
		})
	}
}

func TestVerifyConsistencyProof_TamperingAndMismatches(t *testing.T) {
	leaves := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")}
	m := buildMMRFromLeaves(t, leaves)

	proof, err := m.GenerateConsistencyProof(3, 5)
	if err != nil {
		t.Fatalf("GenerateConsistencyProof() error: %v", err)
	}
	oldRoot := buildMMRFromLeaves(t, leaves[:3]).RootHash()
	newRoot := buildMMRFromLeaves(t, leaves[:5]).RootHash()

	if ok := VerifyConsistencyProof(proof, oldRoot, newRoot, nil); !ok {
		t.Fatalf("baseline verification failed")
	}

	t.Run("wrong old root", func(t *testing.T) {
		wrongOldRoot := buildMMRFromLeaves(t, [][]byte{[]byte("x"), []byte("y"), []byte("z")}).RootHash()
		if ok := VerifyConsistencyProof(proof, wrongOldRoot, newRoot, nil); ok {
			t.Fatalf("expected false with wrong old root")
		}
	})

	t.Run("wrong new root", func(t *testing.T) {
		wrongNewRoot := buildMMRFromLeaves(t, [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("x")}).RootHash()
		if ok := VerifyConsistencyProof(proof, oldRoot, wrongNewRoot, nil); ok {
			t.Fatalf("expected false with wrong new root")
		}
	})

	t.Run("tampered old peak hash", func(t *testing.T) {
		tampered := cloneConsistencyProof(proof)
		tampered.OldPeaksHashes[0][0] ^= 0x01
		if ok := VerifyConsistencyProof(tampered, oldRoot, newRoot, nil); ok {
			t.Fatalf("expected false for tampered old peak hash")
		}
	})

	t.Run("tampered right peak", func(t *testing.T) {
		if len(proof.RightPeaks) == 0 {
			t.Skip("no right peaks for this fixture")
		}
		tampered := cloneConsistencyProof(proof)
		tampered.RightPeaks[0][0] ^= 0x01
		if ok := VerifyConsistencyProof(tampered, oldRoot, newRoot, nil); ok {
			t.Fatalf("expected false for tampered right peak")
		}
	})

	t.Run("mismatched path count", func(t *testing.T) {
		tampered := cloneConsistencyProof(proof)
		tampered.ConsistencyPaths = tampered.ConsistencyPaths[:len(tampered.ConsistencyPaths)-1]
		if ok := VerifyConsistencyProof(tampered, oldRoot, newRoot, nil); ok {
			t.Fatalf("expected false for mismatched proof arrays")
		}
	})

	t.Run("old equals new returns true", func(t *testing.T) {
		trivial := &ConsistencyProof{OldSize: 5, NewSize: 5}
		if ok := VerifyConsistencyProof(trivial, []byte("bad"), []byte("bad"), nil); !ok {
			t.Fatalf("expected true for OldSize == NewSize")
		}
	})
}

func cloneConsistencyProof(in *ConsistencyProof) *ConsistencyProof {
	if in == nil {
		return nil
	}
	out := &ConsistencyProof{
		OldSize:          in.OldSize,
		NewSize:          in.NewSize,
		OldPeaksHashes:   make([][]byte, len(in.OldPeaksHashes)),
		ConsistencyPaths: make([]*merkle.InclusionProof, len(in.ConsistencyPaths)),
		RightPeaks:       make([][]byte, len(in.RightPeaks)),
	}
	for i := range in.OldPeaksHashes {
		out.OldPeaksHashes[i] = append([]byte(nil), in.OldPeaksHashes[i]...)
	}
	for i := range in.RightPeaks {
		out.RightPeaks[i] = append([]byte(nil), in.RightPeaks[i]...)
	}
	for i, p := range in.ConsistencyPaths {
		if p == nil {
			continue
		}
		cp := &merkle.InclusionProof{
			Siblings: make([][]byte, len(p.Siblings)),
			Left:     append([]bool(nil), p.Left...),
		}
		for j := range p.Siblings {
			cp.Siblings[j] = append([]byte(nil), p.Siblings[j]...)
		}
		out.ConsistencyPaths[i] = cp
	}
	return out
}
