package language

import (
	"strings"

	"github.com/pemistahl/lingua-go"
)

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
	if len(strings.TrimSpace(text)) < 3 {
		return false
	}

	lang, exists := d.detector.DetectLanguageOf(text)
	if !exists {
		return false
	}

	return lang != lingua.English
}
