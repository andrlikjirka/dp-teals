package merkle

import (
	"bytes"
	"testing"
)

func TestNewTree(t *testing.T) {
	tests := []struct {
		name       string
		data       [][]byte
		wantErr    bool
		wantLeaves int
	}{
		{
			name:       "valid data with two elements",
			data:       [][]byte{[]byte("hello"), []byte("world")},
			wantErr:    false,
			wantLeaves: 2,
		},
		{
			name:       "single element",
			data:       [][]byte{[]byte("single")},
			wantErr:    false,
			wantLeaves: 1,
		},
		{
			name:       "empty data",
			data:       [][]byte{},
			wantErr:    true,
			wantLeaves: 0,
		},
		{
			name:       "odd number of elements",
			data:       [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			wantErr:    false,
			wantLeaves: 3,
		},
		{
			name:       "many elements",
			data:       [][]byte{[]byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5")},
			wantErr:    false,
			wantLeaves: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := NewTree(tt.data, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(tree.Leaves) != tt.wantLeaves {
				t.Errorf("NewTree() leaves = %d, want %d", len(tree.Leaves), tt.wantLeaves)
			}
		})
	}
}

func TestRootHash(t *testing.T) {
	tests := []struct {
		name          string
		data1         [][]byte
		data2         [][]byte
		shouldBeEqual bool
	}{
		{
			name:          "same data produces same hash",
			data1:         [][]byte{[]byte("test1"), []byte("test2")},
			data2:         [][]byte{[]byte("test1"), []byte("test2")},
			shouldBeEqual: true,
		},
		{
			name:          "different data produces different hash",
			data1:         [][]byte{[]byte("a"), []byte("b")},
			data2:         [][]byte{[]byte("c"), []byte("d")},
			shouldBeEqual: false,
		},
		{
			name:          "different order produces different hash",
			data1:         [][]byte{[]byte("x"), []byte("y")},
			data2:         [][]byte{[]byte("y"), []byte("x")},
			shouldBeEqual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree1, _ := NewTree(tt.data1, nil)
			tree2, _ := NewTree(tt.data2, nil)
			hash1 := tree1.RootHash()
			hash2 := tree2.RootHash()

			isEqual := bytes.Equal(hash1, hash2)
			if isEqual != tt.shouldBeEqual {
				t.Errorf("RootHash equality = %v, want %v", isEqual, tt.shouldBeEqual)
			}
		})
	}
}

func TestRootHashConsistency(t *testing.T) {
	tests := []struct {
		name string
		data [][]byte
	}{
		{
			name: "two elements",
			data: [][]byte{[]byte("test1"), []byte("test2")},
		},
		{
			name: "single element",
			data: [][]byte{[]byte("single")},
		},
		{
			name: "many elements",
			data: [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, _ := NewTree(tt.data, nil)
			hash1 := tree.RootHash()
			hash2 := tree.RootHash()
			if !bytes.Equal(hash1, hash2) {
				t.Error("RootHash not consistent on multiple calls")
			}
		})
	}
}

func TestAppend(t *testing.T) {
	tests := []struct {
		name             string
		initialData      [][]byte
		appendValues     [][]byte
		expectedFinalLen int
	}{
		{
			name:             "append to single element",
			initialData:      [][]byte{[]byte("initial")},
			appendValues:     [][]byte{[]byte("appended")},
			expectedFinalLen: 2,
		},
		{
			name:             "append multiple elements",
			initialData:      [][]byte{[]byte("a")},
			appendValues:     [][]byte{[]byte("b"), []byte("c"), []byte("d")},
			expectedFinalLen: 4,
		},
		{
			name:             "append to two elements",
			initialData:      [][]byte{[]byte("x"), []byte("y")},
			appendValues:     [][]byte{[]byte("z")},
			expectedFinalLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, _ := NewTree(tt.initialData, nil)
			initialHash := tree.RootHash()

			for _, val := range tt.appendValues {
				tree.Append(val)
			}

			if len(tree.Leaves) != tt.expectedFinalLen {
				t.Errorf("Append() leaves = %d, want %d", len(tree.Leaves), tt.expectedFinalLen)
			}
			if bytes.Equal(initialHash, tree.RootHash()) && len(tt.appendValues) > 0 {
				t.Error("Root hash should change after append")
			}
		})
	}
}
