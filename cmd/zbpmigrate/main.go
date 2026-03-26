package main

import (
	"fmt"
	"github.com/chucheka/zenblock-properties/database"
	"os"
	"time"
)

func main() {
	state, err := database.NewStateFromDisk()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer state.Close()

	block0 := database.NewBlock(
		database.Hash{},
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("smith", "smith", 3, ""),
			database.NewTx("smith", "smith", 700, "reward"),
		},
	)

	state.AddBlock(block0)
	block0hash, _ := state.Persist()

	block1 := database.NewBlock(
		block0hash,
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("smith", "ola", 2000, ""),
			database.NewTx("smith", "smith", 100, "reward"),
			database.NewTx("ola", "smith", 1, ""),
			database.NewTx("ola", "caesar", 1000, ""),
			database.NewTx("ola", "smith", 50, ""),
			database.NewTx("smith", "smith", 600, "reward"),
		},
	)

	state.AddBlock(block1)
	state.Persist()
}
