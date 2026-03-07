package mmr

import (
	"bytes"
	"testing"
)

func buildMMRFromLeaves(t *testing.T, leaves [][]byte) *MMR {
	t.Helper()
	m := NewMMR(nil)
	for _, leaf := range leaves {
		if err := m.Append(leaf); err != nil {
			t.Fatalf("append failed for %q: %v", string(leaf), err)
		}
	}
	return m
}

func TestMMRAppendValidation_Table(t *testing.T) {
	tests := []struct {
		name    string
		leaf    []byte
		wantErr bool
	}{
		{name: "nil leaf", leaf: nil, wantErr: true},
		{name: "empty leaf", leaf: []byte{}, wantErr: true},
		{name: "single byte", leaf: []byte("a"), wantErr: false},
		{name: "multi byte", leaf: []byte("hello"), wantErr: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := NewMMR(nil)
			err := m.Append(tc.leaf)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestMMRAppendProgression_Table(t *testing.T) {
	tests := []struct {
		name            string
		leaves          [][]byte
		wantLeavesLen   int
		wantSize        int
		wantPeaksLen    int
		wantPeakHeights []int
	}{
		{
			name:            "1 leaf",
			leaves:          [][]byte{[]byte("a")},
			wantLeavesLen:   1,
			wantSize:        1,
			wantPeaksLen:    1,
			wantPeakHeights: []int{0},
		},
		{
			name:            "2 leaves merge to one peak",
			leaves:          [][]byte{[]byte("a"), []byte("b")},
			wantLeavesLen:   2,
			wantSize:        2,
			wantPeaksLen:    1,
			wantPeakHeights: []int{1},
		},
		{
			name:            "3 leaves produce heights 1,0",
			leaves:          [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			wantLeavesLen:   3,
			wantSize:        3,
			wantPeaksLen:    2,
			wantPeakHeights: []int{1, 0},
		},
		{
			name:            "4 leaves merge to single height 2 peak",
			leaves:          [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
			wantLeavesLen:   4,
			wantSize:        4,
			wantPeaksLen:    1,
			wantPeakHeights: []int{2},
		},
		{
			name:            "5 leaves produce heights 2,0",
			leaves:          [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")},
			wantLeavesLen:   5,
			wantSize:        5,
			wantPeaksLen:    2,
			wantPeakHeights: []int{2, 0},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := buildMMRFromLeaves(t, tc.leaves)

			if got := len(m.Leaves); got != tc.wantLeavesLen {
				t.Fatalf("len(Leaves) = %d, want %d", got, tc.wantLeavesLen)
			}
			if got := m.size; got != tc.wantSize {
				t.Fatalf("size = %d, want %d", got, tc.wantSize)
			}
			if got := len(m.peaks); got != tc.wantPeaksLen {
				t.Fatalf("len(peaks) = %d, want %d", got, tc.wantPeaksLen)
			}
			for i, wantHeight := range tc.wantPeakHeights {
				if got := m.peaks[i].Height; got != wantHeight {
					t.Fatalf("peaks[%d].Height = %d, want %d", i, got, wantHeight)
				}
			}
		})
	}
}

func TestMMRRootHashBehavior_Table(t *testing.T) {
	tests := []struct {
		name          string
		leftLeaves    [][]byte
		rightLeaves   [][]byte
		wantRootsSame bool
	}{
		{
			name:          "same sequence gives same root",
			leftLeaves:    [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			rightLeaves:   [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			wantRootsSame: true,
		},
		{
			name:          "same leaves different order gives different root",
			leftLeaves:    [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			rightLeaves:   [][]byte{[]byte("b"), []byte("a"), []byte("c")},
			wantRootsSame: false,
		},
		{
			name:          "different content gives different root",
			leftLeaves:    [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			rightLeaves:   [][]byte{[]byte("a"), []byte("b"), []byte("x")},
			wantRootsSame: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			left := buildMMRFromLeaves(t, tc.leftLeaves)
			right := buildMMRFromLeaves(t, tc.rightLeaves)

			leftRoot := left.RootHash()
			rightRoot := right.RootHash()
			same := bytes.Equal(leftRoot, rightRoot)
			if same != tc.wantRootsSame {
				t.Fatalf("roots equality = %v, want %v", same, tc.wantRootsSame)
			}
		})
	}
}

func TestMMRRootHashEmptyIsNil(t *testing.T) {
	m := NewMMR(nil)
	if got := m.RootHash(); got != nil {
		t.Fatalf("RootHash() = %x, want nil", got)
	}
}

func TestMMRParentLinksAfterMerge(t *testing.T) {
	m := buildMMRFromLeaves(t, [][]byte{[]byte("a"), []byte("b")})
	if len(m.peaks) != 1 {
		t.Fatalf("len(peaks) = %d, want 1", len(m.peaks))
	}

	root := m.peaks[0]
	if root.Left == nil || root.Right == nil {
		t.Fatalf("merged root should have both children")
	}
	if root.Left.Parent != root {
		t.Fatalf("left child parent pointer not set to root")
	}
	if root.Right.Parent != root {
		t.Fatalf("right child parent pointer not set to root")
	}
}
