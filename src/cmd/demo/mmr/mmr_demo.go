package main

import (
	"fmt"

	"github.com/andrlikjirka/mmr"
)

func main() {
	leaves := [][]byte{
		[]byte("A"),
		[]byte("B"),
		[]byte("C"),
		[]byte("D"),
		[]byte("E"),
	}

	m := mmr.NewMMR(nil)
	m.PrintSummary()

	fmt.Println("\nAdding Leaves...")
	for _, leaf := range leaves {
		err := m.Append(leaf)
		if err != nil {
			panic(err)
		}
	}
	m.PrintSummary()
	m.PrintPeaks()
	m.PrintTree()
}
