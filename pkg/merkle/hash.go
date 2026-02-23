package merkle

import "github.com/andrlikjirka/hash"

// HashLeafData computes the hash of the leaf data by prefixing it with 0x00 and applying the hash function.
func HashLeafData(data []byte, hashFunc hash.HashFunc) []byte {
	prefix := []byte{0x00}
	return hashFunc(append(prefix, data...))
}

// HashInternalNodes computes the hash of the internal nodes by prefixing the concatenated left and right child hashes with 0x01 and applying the hash function.
func HashInternalNodes(left, right []byte, hashFunc hash.HashFunc) []byte {
	prefix := []byte{0x01}
	return hashFunc(append(prefix, append(left, right...)...))
}
