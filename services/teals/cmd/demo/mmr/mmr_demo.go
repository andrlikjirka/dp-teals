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
		//[]byte("event3"),
		//[]byte("event4"),
		//[]byte("event5"),
		//[]byte("event6"),
		//[]byte("event7"),
		//[]byte("event8"),
		//[]byte("event9"),
		//[]byte("event10"),
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

	// 2. Test Inclusion Proof (Existing Data)
	oldRoot := m.RootHash()
	oldSize := len(m.Leaves)
	//demoInclusionProof(m, oldRoot, []byte("event4"))

	// 3. Test Append
	err := m.Append([]byte("event3"))
	err = m.Append([]byte("event4"))
	err = m.Append([]byte("event5"))
	err = m.Append([]byte("event6"))
	err = m.Append([]byte("event7"))
	err = m.Append([]byte("event8"))
	//err = m.Append([]byte("event9"))
	if err != nil {
		panic(err)
	}
	m.PrintSummary()
	m.PrintPeaks()
	m.PrintTree()

	newRoot := m.RootHash()
	newSize := len(m.Leaves)

	// 4. Test Consistency Proof (Old Tree vs New Tree)
	demoConsistencyProof(m, oldSize, newSize, oldRoot, newRoot)

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

	fmt.Printf("Old Peaks Hashes: %d\n", len(proof.OldPeaksHashes))
	fmt.Printf("Consistency Paths: %d\n", len(proof.ConsistencyPaths))
	for i, path := range proof.ConsistencyPaths {
		fmt.Printf("  Path %d: Lefts=%v, Siblings=%d\n", i, len(path.Left), len(path.Siblings))
		for j, sibling := range path.Siblings {
			fmt.Printf("    Sibling %d: %x, %v\n", j, sibling[:4], path.Left[j])
		}
	}

	fmt.Printf("Right Peaks (New Peaks appended): %d\n", len(proof.RightPeaks))
	for i, peak := range proof.RightPeaks {
		fmt.Printf("  Right Peak %d: %x\n", i, peak[:4])
	}

	// 2. Verify the proof using OldPeaksHashes from the proof
	valid := mmr.VerifyConsistencyProof(proof, oldRoot, newRoot, hash.DefaultHashFunc)
	fmt.Printf("Consistency proof valid: %v\n", valid)
	fmt.Println()
}
