package merkle

import "crypto/sha256"

type HashFunc func([]byte) []byte

// DefaultHashFunc uses SHA256
func DefaultHashFunc(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// hashLeafData computes the hash of the leaf data by prefixing it with 0x00 and applying the hash function.
func hashLeafData(data []byte, hashFunc HashFunc) []byte {
	prefix := []byte{0x00}
	return hashFunc(append(prefix, data...))
}

// hashInternalNodes computes the hash of the internal nodes by prefixing the concatenated left and right child hashes with 0x01 and applying the hash function.
func hashInternalNodes(left, right []byte, hashFunc HashFunc) []byte {
	prefix := []byte{0x01}
	return hashFunc(append(prefix, append(left, right...)...))
}
