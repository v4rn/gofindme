package main

import (
	"fmt"
	"os"

	_ "charm.land/bubbletea/v2"
)

func main() {

	index, err := indexPackages("std")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse std lib: %v\n", err)
		os.Exit(1)
	}
	index.Implements = computeImplementations(index)

	if err := loadDB(index); err != nil {
		fmt.Fprintf(os.Stderr, "loading of db failed: %v\n", err)
		os.Exit(1)
	}
}
