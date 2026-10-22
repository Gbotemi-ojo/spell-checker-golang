package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"sc/checker"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func main() {
	dictPath := flag.String("dict", "", "Path to dictionary file (optional)")
	ignoreFile := flag.String("ignore-file", "", "Path to file containing custom words and phrases to ignore")
	var ignoreFlags stringSlice
	flag.Var(&ignoreFlags, "ignore", "Custom words or phrases to ignore (comma-separated or specified multiple times)")
	flag.Parse()

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

	tokenizer := checker.NewPlainTextTokenizer()
	sc := checker.New(dict, tokenizer)

	var customEntries []string
	for _, item := range ignoreFlags {
		for _, part := range strings.Split(item, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				customEntries = append(customEntries, part)
			}
		}
	}

	ignoreFilePath := *ignoreFile
	if ignoreFilePath == "" {
		if _, err := os.Stat(".scignore"); err == nil {
			ignoreFilePath = ".scignore"
		}
	}

	if ignoreFilePath != "" {
		fileEntries, err := loadIgnoreFile(ignoreFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading ignore file: %v\n", err)
			os.Exit(2)
		}
		customEntries = append(customEntries, fileEntries...)
	}

	for _, entry := range customEntries {
		sc.AddCustom(entry)
	}

	args := flag.Args()
	var allTypos []checker.Typo
	var checkedFilesCount int

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
			fmt.Fprintln(os.Stderr, "Usage: sc [-dict path] [-ignore words/phrases] [-ignore-file path] <file-or-directory>...")
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

func loadIgnoreFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func isStdinPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func collectFiles(paths []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
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
