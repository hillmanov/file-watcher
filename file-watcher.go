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

type FileInfo struct {
	FileName  string
	ModTime   time.Time
	LineCount int
}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	pattern := flag.Arg(1)

	ticker := time.NewTicker(5 * time.Second)
	var previousFiles, currentFiles, modifiedFiles map[string]FileInfo

	previousFiles = getFilesMatchingPattern(pattern, root)

	doneCountingLinesChan := make(chan int)
	for fileName, fileInfo := range previousFiles {
		go func(fileName string, fileInfo FileInfo, doneCountingLinesChan chan<- int) {
			currentFileInfo, _ := previousFiles[fileName]
			currentFileInfo.LineCount = countFileLines(fileInfo)
			previousFiles[fileName] = currentFileInfo
			doneCountingLinesChan <- 1
		}(fileName, fileInfo, doneCountingLinesChan)
	}
	for i := 0; i < len(previousFiles); i++ {
		<-doneCountingLinesChan
	}

	for fileName, fileInfo := range previousFiles {
		fmt.Printf("Start values: %s %d\n", fileName, fileInfo.LineCount)
	}

	for {
		select {
		case <-ticker.C:
			// Get a current listing of the files and their mod times
			currentFiles = getFilesMatchingPattern(pattern, root)
			// Get a list of the file names that have been modified based off the mod time
			modifiedFiles = getModifiedFiles(previousFiles, currentFiles)
			// Go through and get updated line counts for just the modified files
			//updateFileLineCounts(modifiedFiles)

			// Loop through the modified files and list the line count change compared to the preivous check.
			// Update the LineCount property in the currentFiles list as we go.

			for fileName, fileInfo := range modifiedFiles {
				fmt.Printf("%s %s\n", fileName, fileInfo.ModTime)
			}
			previousFiles = currentFiles
		}
	}
}

func getFilesMatchingPattern(pattern string, root string) map[string]FileInfo {
	files := make(map[string]FileInfo)
	visit := func(fileName string, f os.FileInfo, err error) error {
		match, err := path.Match(pattern, f.Name())
		if match {
			files[fileName] = FileInfo{FileName: fileName, ModTime: f.ModTime()}
		}
		return nil
	}
	filepath.Walk(root, visit)
	return files
}

func countFileLines(fileInfo FileInfo) int {
	fmt.Printf("Reading %s\n", fileInfo.FileName)
	file, _ := os.Open(fileInfo.FileName)
	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
	}

	return lineCount
}

func getModifiedFiles(previousFiles, currentFiles map[string]FileInfo) map[string]FileInfo {
	modifiedFiles := make(map[string]FileInfo)

	for fileName, previousFileInfo := range previousFiles {
		if currentFileInfo, ok := currentFiles[fileName]; ok {
			if currentFileInfo.ModTime.After(previousFileInfo.ModTime) {
				modifiedFiles[fileName] = currentFileInfo
			}
		}
	}
	return modifiedFiles
}

func getDeletedFiles(oldList, newList []string) []string {
	return nil
}
