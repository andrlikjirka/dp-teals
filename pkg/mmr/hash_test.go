package mmr

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/andrlikjirka/dp-teals/pkg/hash"
)

func sha256Bytes(b []byte) []byte {
	s := sha256.Sum256(b)
	return s[:]
}

func TestHashLeafData(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		validate func([]byte) bool
	}{
		{
			name: "simple leaf data",
			data: []byte("leaf1"),
			validate: func(h []byte) bool {
				expected := sha256Bytes(append([]byte{0x00}, []byte("leaf1")...))
				return bytes.Equal(h, expected)
			},
		},
		{
			name: "empty leaf data",
			data: []byte{},
			validate: func(h []byte) bool {
				expected := sha256Bytes([]byte{0x00})
				return bytes.Equal(h, expected)
			},
		},
		{
			name: "large leaf data",
			data: bytes.Repeat([]byte("x"), 1000),
			validate: func(h []byte) bool {
				expected := sha256Bytes(append([]byte{0x00}, bytes.Repeat([]byte("x"), 1000)...))
				return bytes.Equal(h, expected)
			},
		},
		{
			name: "leaf with prefix should differ from raw hash",
			data: []byte("test"),
			validate: func(h []byte) bool {
				rawHash := hash.DefaultHashFunc([]byte("test"))
				return !bytes.Equal(h, rawHash)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashLeafData(tt.data, hash.DefaultHashFunc)

			if !tt.validate(result) {
				t.Errorf("HashLeafData() validation failed for data: %s", string(tt.data))
			}
		})
	}
}

func TestHashInternalNodes(t *testing.T) {
	tests := []struct {
		name     string
		left     []byte
		right    []byte
		validate func([]byte) bool
	}{
		{
			name:  "both nodes non-empty",
			left:  sha256Bytes([]byte("left")),
			right: sha256Bytes([]byte("right")),
			validate: func(h []byte) bool {
				left := sha256Bytes([]byte("left"))
				right := sha256Bytes([]byte("right"))
				expected := sha256Bytes(append([]byte{0x01}, append(left, right...)...))
				return bytes.Equal(h, expected)
			},
		},
		{
			name:  "empty left node",
			left:  []byte{},
			right: sha256Bytes([]byte("right")),
			validate: func(h []byte) bool {
				right := sha256Bytes([]byte("right"))
				expected := sha256Bytes(append([]byte{0x01}, right...))
				return bytes.Equal(h, expected)
			},
		},
		{
			name:  "empty right node",
			left:  sha256Bytes([]byte("left")),
			right: []byte{},
			validate: func(h []byte) bool {
				left := sha256Bytes([]byte("left"))
				expected := sha256Bytes(append([]byte{0x01}, left...))
				return bytes.Equal(h, expected)
			},
		},
		{
			name:  "both nodes empty",
			left:  []byte{},
			right: []byte{},
			validate: func(h []byte) bool {
				expected := sha256Bytes([]byte{0x01})
				return bytes.Equal(h, expected)
			},
		},
		{
			name:  "same left and right nodes",
			left:  sha256Bytes([]byte("same")),
			right: sha256Bytes([]byte("same")),
			validate: func(h []byte) bool {
				node := sha256Bytes([]byte("same"))
				expected := sha256Bytes(append([]byte{0x01}, append(node, node...)...))
				return bytes.Equal(h, expected)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashInternalNodes(tt.left, tt.right, hash.DefaultHashFunc)

			if !tt.validate(result) {
				t.Errorf("HashInternalNodes() validation failed for left: %x, right: %x", tt.left, tt.right)
			}
		})
	}
}

func TestHashPrefixDifference(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "leaf and internal hash should differ",
			data: []byte("test"),
		},
		{
			name: "leaf and internal hash should differ for empty",
			data: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leafHash := HashLeafData(tt.data, hash.DefaultHashFunc)
			internalHash := HashInternalNodes(tt.data, tt.data, hash.DefaultHashFunc)

			if bytes.Equal(leafHash, internalHash) {
				t.Errorf("leaf and internal hashes should differ but are the same")
			}
		})
	}
}
