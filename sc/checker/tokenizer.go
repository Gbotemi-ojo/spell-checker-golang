package checker

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

type Token struct {
	Word string
	Line int
	Col  int
}

type Tokenizer interface {
	Tokenize(r io.Reader) ([]Token, error)
}

type PlainTextTokenizer struct{}

func NewPlainTextTokenizer() *PlainTextTokenizer {
	return &PlainTextTokenizer{}
}

func (t *PlainTextTokenizer) Tokenize(r io.Reader) ([]Token, error) {
	var tokens []Token
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		tokens = append(tokens, extractWordsFromLine(line, lineNum)...)
	}

	return tokens, scanner.Err()
}

func extractWordsFromLine(line string, lineNum int) []Token {
	var tokens []Token
	runes := []rune(line)
	n := len(runes)

	i := 0
	for i < n {
		for i < n && !isWordChar(runes[i]) {
			i++
		}
		if i >= n {
			break
		}

		start := i
		for i < n && isWordChar(runes[i]) {
			i++
		}
		end := i

		for start < end && runes[start] == '\'' {
			start++
		}
		for end > start && runes[end-1] == '\'' {
			end--
		}

		if start < end {
			tokens = append(tokens, Token{
				Word: string(runes[start:end]),
				Line: lineNum,
				Col:  start + 1,
			})
		}
	}

	return tokens
}

func extractWordStrings(text string) []string {
	var words []string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		for _, tok := range extractWordsFromLine(line, 1) {
			words = append(words, strings.ToLower(tok.Word))
		}
	}
	return words
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || r == '\''
}
