package examplelib

import (
	"slices"
	"unicode"
)

// Символы Unicode, которые принимаются как дефис.
var hyphens = []rune{'\u2010', '-'}

// Символы Unicode, которые принимаются как апостроф.
var apostrophes = []rune{'\'', '`'}

type wordState int

const (
	noWord wordState = iota
	letter
	hyphen
	apostrophe
)

// ExtractWords распознаёт слова в тексте.
// Поддерживает Unicode, в том числе английский и русский языки.
// Слово состоит из букв, может содержать дефис в середине
// и апостроф в середине либо в конце.
func ExtractWords(text string) []string {
	state := noWord

	results := make([]string, 0)
	currentWord := make([]rune, 0)

	pushCurrentWord := func() {
		if len(currentWord) > 0 {
			results = append(results, string(currentWord))
			currentWord = currentWord[:0]
		}
	}

	for _, ch := range text {
		switch state {
		case noWord:
			if unicode.IsLetter(ch) {
				pushCurrentWord()
				currentWord = append(currentWord, ch)
				state = letter
			}

		case letter:
			if unicode.IsLetter(ch) {
				currentWord = append(currentWord, ch)
			} else if slices.Contains(hyphens, ch) {
				currentWord = append(currentWord, ch)
				state = hyphen
			} else if slices.Contains(apostrophes, ch) {
				currentWord = append(currentWord, ch)
				state = apostrophe
			} else {
				state = noWord
			}

		case hyphen:
			if unicode.IsLetter(ch) {
				currentWord = append(currentWord, ch)
				state = letter
			} else {
				// Убираем дефис, которого не должно быть в конце слова.
				currentWord = currentWord[:len(currentWord)-1]
				state = noWord
			}

		case apostrophe:
			if unicode.IsLetter(ch) {
				currentWord = append(currentWord, ch)
				state = letter
			} else {
				state = noWord
			}
		}
	}

	pushCurrentWord()

	return results
}
