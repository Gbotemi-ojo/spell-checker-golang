package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"sc/checker"
)

func main() {
	dictPath := flag.String("dict", "", "Path to dictionary file (optional)")
	flag.Parse()

	// 1. Load Dictionary
	var dict checker.Dictionary
	var err error
	if *dictPath != "" {
		dict, err = checker.NewMemoryDictionary(*dictPath)
	} else {
		dict, err = checker.LoadDefaultDictionary()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading dictionary: %v\n", err)
		os.Exit(2)
	}

	// 2. Initialize Tokenizer and Checker
	tokenizer := checker.NewPlainTextTokenizer()
	sc := checker.New(dict, tokenizer)

	args := flag.Args()
	var allTypos []checker.Typo
	var checkedFilesCount int

	// 3. Process Input (Stdin vs File/Directory paths)
	if len(args) == 0 {
		if isStdinPiped() {
			typos, err := sc.Check(os.Stdin, "stdin")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(2)
			}
			allTypos = append(allTypos, typos...)
			checkedFilesCount = 1
		} else {
			fmt.Fprintln(os.Stderr, "Usage: sc [-dict path] <file-or-directory>...")
			os.Exit(2)
		}
	} else {
		files, err := collectFiles(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting files: %v\n", err)
			os.Exit(2)
		}
		checkedFilesCount = len(files)

		for _, file := range files {
			f, err := os.Open(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", file, err)
				os.Exit(2)
			}
			typos, err := sc.Check(f, file)
			f.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking %s: %v\n", file, err)
				os.Exit(2)
			}
			allTypos = append(allTypos, typos...)
		}
	}

	// 4. Output Results (Silence is golden if no typos)
	if len(allTypos) > 0 {
		fmt.Fprintln(os.Stderr, "Typos found:")
		for _, t := range allTypos {
			if checkedFilesCount > 1 {
				fmt.Fprintf(os.Stderr, "- %s: Line %d, Col %d: %q appears to be a typo\n", t.File, t.Line, t.Col, t.Word)
			} else {
				fmt.Fprintf(os.Stderr, "- Line %d, Col %d: %q appears to be a typo\n", t.Line, t.Col, t.Word)
			}
		}
		os.Exit(1)
	}
}

// isStdinPiped checks if data is being piped into stdin rather than from an interactive terminal.
func isStdinPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// collectFiles walks all given file or directory paths and returns valid files to check.
func collectFiles(paths []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// Skip hidden directories (like .git)
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir
			}
			if !d.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}