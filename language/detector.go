package language

import (
	"strings"
	"unicode"

	"github.com/pemistahl/lingua-go"
)

const minMeaningfulLetters = 8

type Detector struct {
	detector lingua.LanguageDetector
}

func NewDetector() *Detector {
	languages := []lingua.Language{
		lingua.English,
		lingua.Spanish,
		lingua.French,
		lingua.German,
		lingua.Italian,
		lingua.Portuguese,
		lingua.Chinese,
		lingua.Japanese,
		lingua.Korean,
		lingua.Russian,
	}
	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()

	return &Detector{
		detector: detector,
	}
}

func (d *Detector) IsNonEnglish(text string) bool {
	if meaningfulLetterCount(text) < minMeaningfulLetters {
		return false
	}

	lang, exists := d.detector.DetectLanguageOf(text)
	if !exists {
		return false
	}

	return lang != lingua.English
}

func meaningfulLetterCount(text string) int {
	count := 0
	for _, r := range strings.TrimSpace(text) {
		if unicode.IsLetter(r) {
			count++
		}
	}
	return count
}
