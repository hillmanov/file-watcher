package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)

type FileMeta struct {
	FileName  string
	ModTime   time.Time
	LineCount int
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) != 2 {
		fmt.Printf("Command line usage: file-watcher <folder> <file-pattern>\n")
		return
	}

	root := flag.Arg(0)
	pattern := flag.Arg(1)

	var previousFiles, currentFiles, modifiedFiles map[string]FileMeta

	previousFiles = getFilesMatchingPattern(pattern, root)
	countFilesLines(previousFiles)

	for fileName, fileMeta := range previousFiles {
		fmt.Printf("Start values: %s %d\n", fileName, fileMeta.LineCount)
	}

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			// Get a current listing of the files and their mod times
			currentFiles = getFilesMatchingPattern(pattern, root)

			newFiles, deletedFiles := getNewAndDeletedFiles(previousFiles, currentFiles)

			for newFile := range newFiles {
				fmt.Printf("New file: %s\n", newFile)
			}

			for deletedFile := range deletedFiles {
				fmt.Printf("Deleted file: %s\n", deletedFile)
			}

			// Get a list of the file names that have been modified based off the mod time
			// Go through and get updated line counts for just the modified files
			modifiedFiles = getModifiedFiles(previousFiles, currentFiles)
			if len(modifiedFiles) > 0 {
				countFilesLines(modifiedFiles)
				// Loop through the modified files and list the line count change compared to the preivous check.
				// Update the LineCount property in the currentFiles list as we go.
				for fileName, fileMeta := range modifiedFiles {
					previousFileMeta, _ := previousFiles[fileName]
					fmt.Printf("Last count: %d new count: %d\n", previousFileMeta.LineCount, fileMeta.LineCount)
					currentFiles[fileName] = fileMeta
				}

			}
			previousFiles = currentFiles
		}
	}
}

func getFilesMatchingPattern(pattern string, root string) map[string]FileMeta {
	files := make(map[string]FileMeta)
	visit := func(fileName string, f os.FileInfo, err error) error {
		match, err := path.Match(pattern, f.Name())
		if match {
			files[fileName] = FileMeta{FileName: fileName, ModTime: f.ModTime()}
		}
		return nil
	}
	filepath.Walk(root, visit)
	return files
}

func getModifiedFiles(previousFiles, currentFiles map[string]FileMeta) map[string]FileMeta {
	modifiedFiles := make(map[string]FileMeta)

	for fileName, previousFileMeta := range previousFiles {
		if currentFileMeta, ok := currentFiles[fileName]; ok {
			if currentFileMeta.ModTime.After(previousFileMeta.ModTime) {
				modifiedFiles[fileName] = currentFileMeta
			}
		}
	}
	return modifiedFiles
}

func getNewAndDeletedFiles(previousFiles, currentFiles map[string]FileMeta) ([]string, []string) {
	var newFiles, deletedFiles []string
	for fileName, _ := range previousFiles {
		if _, ok := currentFiles[fileName]; !ok {
			deletedFiles = append(deletedFiles, fileName)
		}
	}

	for fileName, _ := range currentFiles {
		if _, ok := previousFiles[fileName]; !ok {
			newFiles = append(newFiles, fileName)
		}
	}

	return newFiles, deletedFiles
}

func countFilesLines(files map[string]FileMeta) {
	doneCountingLinesChan := make(chan int)
	for fileName, fileMeta := range files {
		go func(fileName string, fileMeta FileMeta, doneCountingLinesChan chan<- int) {
			currentFileMeta, _ := files[fileName]
			currentFileMeta.LineCount = countFileLines(fileMeta)
			files[fileName] = currentFileMeta
			doneCountingLinesChan <- 1
		}(fileName, fileMeta, doneCountingLinesChan)
	}
	for i := 0; i < len(files); i++ {
		<-doneCountingLinesChan
	}
	return
}

func countFileLines(fileMeta FileMeta) int {
	file, _ := os.Open(fileMeta.FileName)
	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
	}

	return lineCount
}
