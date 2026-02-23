package hash

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
