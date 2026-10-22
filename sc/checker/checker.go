package checker

import (
	"bytes"
	"io"
	"strings"
	"unicode"
)

type Typo struct {
	File string
	Word string
	Line int
	Col  int
}

type Checker struct {
	dict          Dictionary
	tokenizer     Tokenizer
	customWords   map[string]struct{}
	customPhrases [][]string
}

func New(dict Dictionary, tokenizer Tokenizer) *Checker {
	return &Checker{
		dict:        dict,
		tokenizer:   tokenizer,
		customWords: make(map[string]struct{}),
	}
}

func (c *Checker) AddCustom(entry string) {
	words := extractWordStrings(entry)
	if len(words) == 1 {
		c.AddCustomWord(words[0])
	} else if len(words) > 1 {
		c.customPhrases = append(c.customPhrases, words)
	}
}

func (c *Checker) AddCustomWord(word string) {
	w := strings.ToLower(strings.TrimSpace(word))
	if w != "" {
		c.customWords[w] = struct{}{}
	}
}

func (c *Checker) AddCustomWords(words ...string) {
	for _, w := range words {
		c.AddCustom(w)
	}
}

func (c *Checker) AddCustomPhrase(phrase string) {
	c.AddCustom(phrase)
}

func (c *Checker) AddCustomPhrases(phrases ...string) {
	for _, p := range phrases {
		c.AddCustom(p)
	}
}

func (c *Checker) Check(r io.Reader, sourceName string) ([]Typo, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	tokens, err := c.tokenizer.Tokenize(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}

	ignored := make([]bool, len(tokens))

	for _, phrase := range c.customPhrases {
		k := len(phrase)
		if k < 2 || len(tokens) < k {
			continue
		}

		for i := 0; i <= len(tokens)-k; i++ {
			matched := true
			for j := 0; j < k; j++ {
				if !strings.EqualFold(tokens[i+j].Word, phrase[j]) {
					matched = false
					break
				}
			}
			if !matched {
				continue
			}

			consecutive := true
			for j := 0; j < k-1; j++ {
				if !isValidPhraseTransition(tokens[i+j], tokens[i+j+1], lines) {
					consecutive = false
					break
				}
			}

			if consecutive {
				for j := 0; j < k; j++ {
					ignored[i+j] = true
				}
			}
		}
	}

	var typos []Typo
	for i, tok := range tokens {
		if ignored[i] {
			continue
		}

		if _, ok := c.customWords[strings.ToLower(tok.Word)]; ok {
			continue
		}

		if !c.dict.Contains(tok.Word) {
			typos = append(typos, Typo{
				File: sourceName,
				Word: tok.Word,
				Line: tok.Line,
				Col:  tok.Col,
			})
		}
	}

	return typos, nil
}

func isValidPhraseTransition(tokA, tokB Token, lines []string) bool {
	if tokA.Line == tokB.Line {
		if tokA.Line < 1 || tokA.Line > len(lines) {
			return false
		}
		runes := []rune(lines[tokA.Line-1])
		tokAEnd := (tokA.Col - 1) + len([]rune(tokA.Word))
		tokBStart := tokB.Col - 1
		if tokAEnd > tokBStart || tokAEnd > len(runes) || tokBStart > len(runes) {
			return false
		}
		between := string(runes[tokAEnd:tokBStart])
		return isAllowedPhraseSeparator(between)
	}

	if tokB.Line == tokA.Line+1 {
		if tokA.Line < 1 || tokB.Line > len(lines) {
			return false
		}
		runesA := []rune(lines[tokA.Line-1])
		tokAEnd := (tokA.Col - 1) + len([]rune(tokA.Word))
		if tokAEnd > len(runesA) {
			return false
		}
		afterA := string(runesA[tokAEnd:])
		if !isAllowedPhraseSeparator(afterA) {
			return false
		}

		runesB := []rune(lines[tokB.Line-1])
		tokBStart := tokB.Col - 1
		if tokBStart > len(runesB) {
			return false
		}
		beforeB := string(runesB[:tokBStart])
		return isAllowedPhraseSeparator(beforeB)
	}

	return false
}

func isAllowedPhraseSeparator(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return false
		}
		switch r {
		case '.', '!', '?', ';', ':':
			return false
		}
	}
	return true
}
