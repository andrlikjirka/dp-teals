package main

import (
	"fmt"

	"github.com/andrlikjirka/mmr"
)

func main() {
	data := [][]byte{
		[]byte("tx1"),
		[]byte("tx2"),
		[]byte("tx3"),
		[]byte("tx4"),
		[]byte("tx5"),
		[]byte("tx6"),
		[]byte("tx7"),
		[]byte("tx8"),
		[]byte("tx9"),
	}

	m := mmr.NewMMR(nil)
	m.PrintSummary()

	fmt.Println("\nAdding Leaves...")
	for _, d := range data {
		err := m.Append(d)
		if err != nil {
			panic(err)
		}
	}
	m.PrintSummary()
	m.PrintPeaks()
	m.PrintTree()

	demoInclusionProof(m, m.RootHash(), []byte("tx2"))
}

func demoInclusionProof(m *mmr.MMR, root []byte, targetData []byte) {
	fmt.Printf("--- Testing Inclusion Proof for '%s' ---\n", string(targetData))

	proof, err := m.GenerateInclusionProofByData(targetData)
	if err != nil {
		fmt.Printf("Error generating inclusion proof: %v\n", err)
		return
	}
	fmt.Printf("Proof generated with %d siblings:\n", len(proof.Siblings))
	for i, sibling := range proof.Siblings {
		fmt.Printf("  Sibling %d: %x\n", i, sibling[:4])
	}

	valid := mmr.VerifyInclusionProof(targetData, proof, root, nil)
	fmt.Printf("Inclusion proof valid: %v\n", valid)
	fmt.Println()
}
