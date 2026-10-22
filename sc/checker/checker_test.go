package checker

import (
	"strings"
	"testing"
)

type mockDictionary struct {
	words map[string]struct{}
}

func newMockDict(words ...string) *mockDictionary {
	m := make(map[string]struct{})
	for _, w := range words {
		m[strings.ToLower(w)] = struct{}{}
	}
	return &mockDictionary{words: m}
}

func (m *mockDictionary) Contains(word string) bool {
	_, ok := m.words[strings.ToLower(word)]
	return ok
}

func TestCheckerBasic(t *testing.T) {
	dict := newMockDict("hello", "world")
	tokenizer := NewPlainTextTokenizer()
	c := New(dict, tokenizer)

	input := "hello wrld"
	typos, err := c.Check(strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(typos) != 1 {
		t.Fatalf("expected 1 typo, got %d", len(typos))
	}

	if typos[0].Word != "wrld" || typos[0].Line != 1 || typos[0].Col != 7 {
		t.Errorf("unexpected typo details: %+v", typos[0])
	}
}

func TestCheckerCustomWords(t *testing.T) {
	dict := newMockDict("we", "use")
	tokenizer := NewPlainTextTokenizer()
	c := New(dict, tokenizer)
	c.AddCustomWord("Golang")
	c.AddCustomWord("KUBERNETES")

	input := "We use golang and kubernetes."
	typos, err := c.Check(strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(typos) != 1 {
		t.Fatalf("expected 1 typo ('and'), got %d: %+v", len(typos), typos)
	}

	if typos[0].Word != "and" {
		t.Errorf("expected typo 'and', got %s", typos[0].Word)
	}
}

func TestCheckerCustomPhrases(t *testing.T) {
	dict := newMockDict("the", "was", "maintained", "decision")
	tokenizer := NewPlainTextTokenizer()
	c := New(dict, tokenizer)
	c.AddCustomPhrase("status quo")
	c.AddCustomPhrase("ad hoc")

	input := "The status quo was maintained in an ad-hoc decision."
	typos, err := c.Check(strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(typos) != 2 {
		t.Fatalf("expected 2 typos ('in', 'an'), got %d: %+v", len(typos), typos)
	}

	for _, typo := range typos {
		if typo.Word != "in" && typo.Word != "an" {
			t.Errorf("unexpected typo: %s", typo.Word)
		}
	}
}

func TestCheckerPhraseWordOutsidePhrase(t *testing.T) {
	dict := newMockDict("the", "is", "good", "but")
	tokenizer := NewPlainTextTokenizer()
	c := New(dict, tokenizer)
	c.AddCustomPhrase("status quo")

	input := "The status quo is good, but quo is not."
	typos, err := c.Check(strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundQuo := false
	for _, typo := range typos {
		if typo.Word == "quo" {
			foundQuo = true
			if typo.Col != 29 {
				t.Errorf("expected quo typo at Col 29, got %d", typo.Col)
			}
		}
	}

	if !foundQuo {
		t.Errorf("expected 'quo' outside phrase to be flagged as typo, but was not")
	}
}

func TestCheckerPhraseAcrossLineBreak(t *testing.T) {
	dict := newMockDict("the", "is", "fine")
	tokenizer := NewPlainTextTokenizer()
	c := New(dict, tokenizer)
	c.AddCustomPhrase("status quo")

	input := "The status\n  quo is fine."
	typos, err := c.Check(strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(typos) != 0 {
		t.Fatalf("expected 0 typos, got %d: %+v", len(typos), typos)
	}
}

func TestCheckerPhraseSentenceBoundary(t *testing.T) {
	dict := newMockDict("the", "fine", "is", "latin")
	tokenizer := NewPlainTextTokenizer()
	c := New(dict, tokenizer)
	c.AddCustomPhrase("status quo")

	input := "The status. Quo is latin."
	typos, err := c.Check(strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundQuo := false
	for _, typo := range typos {
		if typo.Word == "Quo" {
			foundQuo = true
		}
	}

	if !foundQuo {
		t.Errorf("expected 'Quo' across sentence boundary to be flagged as typo")
	}
}

func TestCheckerPhraseWordBoundary(t *testing.T) {
	dict := newMockDict("the", "is", "fine")
	tokenizer := NewPlainTextTokenizer()
	c := New(dict, tokenizer)
	c.AddCustomPhrase("status quo")

	input := "The status quota is fine."
	typos, err := c.Check(strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundQuota := false
	for _, typo := range typos {
		if typo.Word == "quota" {
			foundQuota = true
		}
	}

	if !foundQuota {
		t.Errorf("expected 'quota' to be flagged as typo")
	}
}

func TestCheckerAddCustomVariants(t *testing.T) {
	dict := newMockDict("test")
	tokenizer := NewPlainTextTokenizer()
	c := New(dict, tokenizer)

	c.AddCustom("singleword")
	c.AddCustom("multi word phrase")
	c.AddCustomWords("anotherword", "yet another phrase")
	c.AddCustomPhrases("three word term")

	input := "test singleword multi word phrase anotherword yet another phrase three word term"
	typos, err := c.Check(strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(typos) != 0 {
		t.Fatalf("expected 0 typos, got %d: %+v", len(typos), typos)
	}
}
