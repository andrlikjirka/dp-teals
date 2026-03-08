package hash

import (
	"crypto/sha256"
	"crypto/sha3"
)

// Func defines the type for hash functions used in the Merkle tree.
type Func func([]byte) []byte

// DefaultHashFunc uses SHA256.
func DefaultHashFunc(data []byte) []byte {
	return SHA256HashFunc(data)
}

// SHA256HashFunc uses SHA256 hash function.
func SHA256HashFunc(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// SHA3HashFunc uses SHA3-256 hash function.
func SHA3HashFunc(data []byte) []byte {
	h := sha3.Sum256(data)
	return h[:]
}
