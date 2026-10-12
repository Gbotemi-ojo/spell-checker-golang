package checker

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)


type Dictionary interface {
	Contains(word string) bool
}

type MemoryDictionary struct {
	words map[string]struct{}
}

func NewMemoryDictionary(path string) (*MemoryDictionary, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	words := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		w := strings.TrimSpace(scanner.Text())
		if w != "" {
			words[strings.ToLower(w)] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &MemoryDictionary{words: words}, nil
}

func LoadDefaultDictionary() (Dictionary, error) {
	if customPath := os.Getenv("SC_DICT_PATH"); customPath != "" {
		return NewMemoryDictionary(customPath)
	}

	commonPaths := []string{
		"/usr/share/dict/words",
		"/usr/dict/words",
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return NewMemoryDictionary(p)
		}
	}

	return nil, fmt.Errorf("no system dictionary found; set SC_DICT_PATH or pass -dict")
}

func (d *MemoryDictionary) Contains(word string) bool {
	_, ok := d.words[strings.ToLower(word)]
	return ok
}
