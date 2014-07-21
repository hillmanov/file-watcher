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

type FileStatus struct {
	FileName  string
	ModTime   time.Time
	LineCount int
}

type FileStatuses map[string]FileStatus

var root, pattern string

func init() {
	flag.Parse()
	args := flag.Args()

	if len(args) != 2 {
		fmt.Printf("Command line usage: file-watcher <folder> <file-pattern>\n")
		return
	}

	root = args[0]
	pattern = args[1]
}

func main() {
	var previousFiles, currentFiles, modifiedFiles FileStatuses

	previousFiles = getFilesMatchingPattern(pattern, root)
	countFilesLines(previousFiles)

	for fileName, fileStatus := range previousFiles {
		fmt.Printf("Start values: %s %d\n", fileName, fileStatus.LineCount)
	}

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			// Get a current listing of the files and their mod times
			currentFiles = getFilesMatchingPattern(pattern, root)

			// Get a list of new files and deleted files
			newFiles, deletedFiles := getNewAndDeletedFiles(previousFiles, currentFiles)
			for fileName, fileStatus := range newFiles {
				fmt.Printf("%s: %s: %d\n", "New", fileName, fileStatus.LineCount)
			}

			for _, fileName := range deletedFiles {
				fmt.Printf("%s: %s\n", "Deleted", fileName)
			}

			// Get a list of the file names that have been modified based off the mod time
			// Go through and get updated line counts for just the modified files
			modifiedFiles = getModifiedFiles(previousFiles, currentFiles)
			if len(modifiedFiles) > 0 {
				countFilesLines(modifiedFiles)
				// Loop through the modified files and list the line count change compared to the previous check.
				for fileName, currentFileStatus := range modifiedFiles {
					previousFileStatus, _ := previousFiles[fileName]
					displayLineCountDiff(previousFileStatus, currentFileStatus)
				}
			}

			// Make sure new files have line counts, and only include the actual current state of the file system.
			previousFiles = syncFileStatus(previousFiles, currentFiles, modifiedFiles, newFiles)
		}
	}
}

func displayLineCountDiff(previousFileStatus, currentFileStatus FileStatus) {
	var (
		sign     string
		lineDiff int
	)
	switch lineDiff = currentFileStatus.LineCount - previousFileStatus.LineCount; {
	case lineDiff == 0:
		return
	case lineDiff > 0:
		sign = "+"
	}
	fmt.Printf("%s %s%d\n", currentFileStatus.FileName, sign, lineDiff)
}

func getFilesMatchingPattern(pattern string, root string) FileStatuses {
	files := make(FileStatuses)
	visit := func(fileName string, f os.FileInfo, err error) error {
		match, err := path.Match(pattern, f.Name())
		if match {
			files[fileName] = FileStatus{FileName: fileName, ModTime: f.ModTime()}
		}
		return nil
	}
	filepath.Walk(root, visit)
	return files
}

func getModifiedFiles(previousFiles, currentFiles FileStatuses) FileStatuses {
	modifiedFiles := make(FileStatuses)

	for fileName, previousFileStatus := range previousFiles {
		if currentFileStatus, ok := currentFiles[fileName]; ok {
			if currentFileStatus.ModTime.After(previousFileStatus.ModTime) {
				modifiedFiles[fileName] = currentFileStatus
			}
		}
	}
	return modifiedFiles
}

func getNewAndDeletedFiles(previousFiles, currentFiles FileStatuses) (FileStatuses, []string) {
	newFiles := make(FileStatuses)
	var deletedFiles []string

	for fileName, fileStatus := range currentFiles {
		if _, ok := previousFiles[fileName]; !ok {
			newFiles[fileName] = fileStatus
		}
	}
	countFilesLines(newFiles)

	for fileName, _ := range previousFiles {
		if _, ok := currentFiles[fileName]; !ok {
			deletedFiles = append(deletedFiles, fileName)
		}
	}

	return newFiles, deletedFiles
}

func countFilesLines(files FileStatuses) {
	done := make(chan int)
	for fileName, fileStatus := range files {
		go func(fileName string, fileStatus FileStatus, done chan<- int) {
			currentFileStatus, _ := files[fileName]
			currentFileStatus.LineCount = countFileLines(fileStatus.FileName)
			files[fileName] = currentFileStatus
			done <- 1
		}(fileName, fileStatus, done)
	}
	for i := 0; i < len(files); i++ {
		<-done
	}
	return
}

func countFileLines(fileName string) int {
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		panic("Error opening file: " + fileName)
	}

	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
	}

	return lineCount
}

func syncFileStatus(previousFiles, currentFiles, modifiedFiles, newFiles FileStatuses) FileStatuses {
	// Loop through the current files, and get the FileStatus information from the previous files so we don't have to count the lines again.
	// Only do this for items where the ModTime is the same - otherwise keep the currentFiles copy.
	for fileName, fileStatus := range currentFiles {
		if previousFileStatus, ok := previousFiles[fileName]; ok {
			if fileStatus.ModTime == previousFileStatus.ModTime {
				// Keep the old state - it has the accurate line count!
				currentFiles[fileName] = previousFileStatus
			}
		}
	}

	for fileName, fileStatus := range modifiedFiles {
		currentFiles[fileName] = fileStatus
	}

	for fileName, fileStatus := range newFiles {
		currentFiles[fileName] = fileStatus
	}

	return currentFiles
}
