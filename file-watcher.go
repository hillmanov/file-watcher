package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)

func main() {
	flag.Parse()
	root := flag.Arg(0)
	pattern := flag.Arg(1)

	ticker := time.NewTicker(5 * time.Second)
	var modifiedFiles map[string]time.Time

	currentFiles := getFilesMatchingPattern(pattern, root)

	for {
		select {
		case <-ticker.C:
			modifiedFiles = getModifiedFiles(currentFiles)

		}

	}

	for name, modTime := range currentFiles {
		fmt.Printf("%s %s\n", name, modTime)
	}
}

func getFilesMatchingPattern(pattern string, root string) map[string]time.Time {
	files := make(map[string]time.Time)

	visit := func(fileName string, f os.FileInfo, err error) error {
		match, err := path.Match(pattern, f.Name())
		if match {
			files[fileName] = f.ModTime()
		}
		return nil
	}

	filepath.Walk(root, visit)

	return files
}

func checkForDeletedFiles(oldList, newList []string) []string {
	return nil
}

func countFileLines(path string) int {
	return 1
}

func getModifiedFiles(originalList map[string]time.Time) map[string]time.Time {
	return nil
}
