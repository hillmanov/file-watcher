package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Read command line args
	flag.Parse()
	root := flag.Arg(0)

	visit := func(path string, f os.FileInfo, err error) error {
		fmt.Printf("Visted %s\n", path)
		return nil
	}

	filepath.Walk(root, visit)
}
