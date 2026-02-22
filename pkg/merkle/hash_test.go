package merkle

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func sha256Bytes(b []byte) []byte {
	s := sha256.Sum256(b)
	return s[:]
}

func TestDefaultHashFunc(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected []byte
	}{
		{
			name:     "empty data",
			data:     []byte{},
			expected: sha256Bytes([]byte{})},
		{
			name:     "simple string",
			data:     []byte("hello"),
			expected: sha256Bytes([]byte("hello"))},
		{
			name:     "longer string",
			data:     []byte("merkle tree test data"),
			expected: sha256Bytes([]byte("merkle tree test data"))},
		{
			name:     "binary data",
			data:     []byte{0x00, 0x01, 0x02, 0x03},
			expected: sha256Bytes([]byte{0x00, 0x01, 0x02, 0x03})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultHashFunc(tt.data)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("DefaultHashFunc() = %x, want %x", result, tt.expected)
			}
		})
	}
}

func TestHashLeafData(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		expectErr bool
		validate  func([]byte) bool
	}{
		{
			name:      "simple leaf data",
			data:      []byte("leaf1"),
			expectErr: false,
			validate: func(hash []byte) bool {
				return len(hash) == 32
			},
		},
		{
			name:      "empty leaf data",
			data:      []byte{},
			expectErr: false,
			validate: func(hash []byte) bool {
				return len(hash) == 32
			},
		},
		{
			name:      "large leaf data",
			data:      bytes.Repeat([]byte("x"), 1000),
			expectErr: false,
			validate: func(hash []byte) bool {
				return len(hash) == 32
			},
		},
		{
			name:      "leaf with prefix should differ from raw hash",
			data:      []byte("test"),
			expectErr: false,
			validate: func(hash []byte) bool {
				expected := DefaultHashFunc([]byte("test"))
				return !bytes.Equal(hash, expected)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := NewTree([][]byte{tt.data}, nil)
			if err != nil {
				t.Fatalf("Failed to create tree: %v", err)
			}

			result := hashLeafData(tt.data, tree.hashFunc)

			if (err != nil) != tt.expectErr {
				t.Errorf("hashLeafData() error = %v, wantErr %v", err, tt.expectErr)
				return
			}

			if !tt.validate(result) {
				t.Errorf("hashLeafData() validation failed for data: %s", string(tt.data))
			}
		})
	}
}

func TestHashInternalNodes(t *testing.T) {
	tests := []struct {
		name      string
		left      []byte
		right     []byte
		expectErr bool
		validate  func([]byte) bool
	}{
		{
			name:      "both nodes non-empty",
			left:      sha256Bytes([]byte("left")),
			right:     sha256Bytes([]byte("right")),
			expectErr: false,
			validate: func(hash []byte) bool {
				return len(hash) == 32
			},
		},
		{
			name:      "empty left node",
			left:      []byte{},
			right:     sha256Bytes([]byte("right")),
			expectErr: false,
			validate: func(hash []byte) bool {
				return len(hash) == 32
			},
		},
		{
			name:      "empty right node",
			left:      sha256Bytes([]byte("left")),
			right:     []byte{},
			expectErr: false,
			validate: func(hash []byte) bool {
				return len(hash) == 32
			},
		},
		{
			name:      "both nodes empty",
			left:      []byte{},
			right:     []byte{},
			expectErr: false,
			validate: func(hash []byte) bool {
				return len(hash) == 32
			},
		},
		{
			name:      "same left and right nodes",
			left:      sha256Bytes([]byte("same")),
			right:     sha256Bytes([]byte("same")),
			expectErr: false,
			validate: func(hash []byte) bool {
				return len(hash) == 32
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := NewTree([][]byte{[]byte("dummy")}, nil)
			if err != nil {
				t.Fatalf("Failed to create tree: %v", err)
			}

			result := hashInternalNodes(tt.left, tt.right, tree.hashFunc)

			if (err != nil) != tt.expectErr {
				t.Errorf("hashInternalNodes() error = %v, wantErr %v", err, tt.expectErr)
				return
			}

			if !tt.validate(result) {
				t.Errorf("hashInternalNodes() validation failed for left: %x, right: %x", tt.left, tt.right)
			}
		})
	}
}

func TestHashPrefixDifference(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		shouldDiffer bool
	}{
		{
			name:         "leaf and internal hash should differ",
			data:         []byte("test"),
			shouldDiffer: true,
		},
		{
			name:         "leaf and internal hash should differ for empty",
			data:         []byte{},
			shouldDiffer: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, _ := NewTree([][]byte{tt.data}, nil)

			leafHash := hashLeafData(tt.data, tree.hashFunc)
			internalHash := hashInternalNodes(tt.data, tt.data, tree.hashFunc)

			differ := !bytes.Equal(leafHash, internalHash)
			if differ != tt.shouldDiffer {
				t.Errorf("Hash prefix difference = %v, want %v", differ, tt.shouldDiffer)
			}
		})
	}
}
