package hash

import "crypto/sha256"

type HashFunc func([]byte) []byte

// DefaultHashFunc uses SHA256
func DefaultHashFunc(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}
