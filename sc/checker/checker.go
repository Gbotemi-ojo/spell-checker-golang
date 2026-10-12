package checker

import "io"

type Typo struct {
	File string
	Word string
	Line int
	Col  int
}

type Checker struct {
	dict      Dictionary
	tokenizer Tokenizer
}

func New(dict Dictionary, tokenizer Tokenizer) *Checker {
	return &Checker{
		dict:      dict,
		tokenizer: tokenizer,
	}
}

func (c *Checker) Check(r io.Reader, sourceName string) ([]Typo, error) {
	tokens, err := c.tokenizer.Tokenize(r)
	if err != nil {
		return nil, err
	}

	var typos []Typo
	for _, tok := range tokens {
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
