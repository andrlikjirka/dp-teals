package main

import (
	"fmt"

	"github.com/andrlikjirka/dp-teals/pkg/hash"
	"github.com/andrlikjirka/dp-teals/pkg/mmr"
)

func main() {
	// 1. Initialization
	data := [][]byte{
		[]byte("event1"),
		[]byte("event2"),
		[]byte("event3"),
		[]byte("event4"),
		[]byte("event5"),
		[]byte("event6"),
		[]byte("event7"),
		[]byte("event8"),
		[]byte("event9"),
		[]byte("event10"),
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

	// 2. Test Inclusion Proof (Existing Data)
	oldRoot := m.RootHash()
	//oldSize := len(m.Leaves)
	demoInclusionProof(m, oldRoot, []byte("event4"))

	/*
		// 3. Test Append
			err := m.Append([]byte("tx6"))
			err = m.Append([]byte("tx7"))
			err = m.Append([]byte("tx8"))
			err = m.Append([]byte("tx9"))
			if err != nil {
				panic(err)
			}
			newRoot := m.RootHash()
			newSize := len(m.Leaves)

			// 4. Test Consistency Proof (Old Tree vs New Tree)
			demoConsistencyProof(m, oldSize, newSize, oldRoot, newRoot)
	*/

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

	valid := mmr.VerifyInclusionProof(targetData, proof, root, hash.DefaultHashFunc)
	fmt.Printf("Inclusion proof valid: %v\n", valid)
	fmt.Println()
}

func demoConsistencyProof(m *mmr.MMR, oldSize int, newSize int, oldRoot []byte, newRoot []byte) {
	fmt.Printf("--- Testing Consistency Proof (Size %d -> %d) ---\n", oldSize, newSize)

	// 1. Generate the consistency proof
	proof, err := m.GenerateConsistencyProof(oldSize, newSize)
	if err != nil {
		fmt.Printf("Error generating consistency proof: %v\n", err)
		return
	}

	fmt.Println("Proof generated successfully:")
	fmt.Printf("  Consistency Paths (Old Peaks advancing): %d\n", len(proof.ConsistencyPaths))
	fmt.Printf("  Right Peaks (New Peaks appended): %d\n", len(proof.RightPeaks))

	// 2. Verify the proof using OldPeaksHashes from the proof
	valid := mmr.VerifyConsistencyProof(proof, oldRoot, newRoot, hash.DefaultHashFunc)
	fmt.Printf("Consistency proof valid: %v\n", valid)
	fmt.Println()
}
