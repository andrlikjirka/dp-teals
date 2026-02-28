package main

import (
	"fmt"
	"log"

	"github.com/andrlikjirka/merkle"
)

func main() {
	fmt.Println("=== Merkle Tree Cryptography Demo ===")
	fmt.Println()

	// 1. Initialization
	initialData := [][]byte{
		[]byte("tx1"), []byte("tx2"), []byte("tx3"), []byte("tx4"), []byte("tx5"),
	}
	tree, err := merkle.NewTree(initialData, nil)
	if err != nil {
		log.Fatalf("Failed to initialize tree: %v", err)
	}

	oldSize := len(tree.Leaves)
	oldRoot := tree.RootHash()
	fmt.Printf("[+] Initialized tree with %d leaves\n", oldSize)
	fmt.Printf("[+] Initial Root Hash: %x\n", oldRoot)
	fmt.Println()
	tree.Print()
	fmt.Println()

	// 2. Test Inclusion Proof (Existing Data)
	demoInclusionProof(tree, oldRoot, []byte("tx3"))

	// 3. Test Append
	newData := []byte("tx6")
	newRoot := demoAppend(tree, newData)
	tree.Print()
	fmt.Println()

	// 4. Test Inclusion Proof (Newly Appended Data)
	demoInclusionProof(tree, newRoot, newData)

	// 5. Test Consistency Proof (Old Tree vs New Tree)
	newSize := len(tree.Leaves)
	demoConsistencyProof(tree, oldSize, newSize, oldRoot, newRoot)
}

// --- Helper Functions ---

func demoInclusionProof(tree *merkle.Tree, root []byte, targetData []byte) {
	fmt.Printf("--- Testing Inclusion Proof for '%s' ---\n", string(targetData))

	proof, err := tree.GenerateInclusionProofByData(targetData)
	if err != nil {
		log.Fatalf("Failed to generate inclusion proof: %v", err)
	}

	valid := merkle.VerifyInclusionProof(targetData, proof, root, nil)
	fmt.Printf("Proof generated with %d siblings\n", len(proof.Siblings))
	fmt.Printf("Inclusion proof valid: %v\n", valid)
	fmt.Println()
}

func demoAppend(tree *merkle.Tree, newData []byte) []byte {
	fmt.Printf("--- Testing Append for '%s' ---\n", string(newData))

	err := tree.Append(newData)
	if err != nil {
		log.Fatalf("Failed to append data: %v", err)
	}

	newRoot := tree.RootHash()
	fmt.Printf("Successfully appended data. New leaf count: %d\n", len(tree.Leaves))
	fmt.Printf("New Root Hash: %x\n", newRoot)
	fmt.Println()

	return newRoot
}

func demoConsistencyProof(tree *merkle.Tree, m, n int, oldRoot, newRoot []byte) {
	fmt.Println("--- Testing Consistency Proof ---")
	fmt.Printf("Proving tree of size %d is a prefix of tree size %d\n", m, n)

	// Optional: Print the tree structure for visual debugging
	// tree.Print()

	proof, err := tree.GenerateConsistencyProof(m)
	if err != nil {
		log.Fatalf("Failed to generate consistency proof: %v", err)
	}

	fmt.Printf("Consistency proof generated with %d hashes:\n", len(proof.Hashes))
	for i, h := range proof.Hashes {
		fmt.Printf("  Hash %d: %x\n", i, h)
	}

	valid := merkle.VerifyConsistencyProof(m, n, oldRoot, newRoot, proof, nil)
	fmt.Printf("Consistency proof valid: %v\n", valid)
	fmt.Println()
}
