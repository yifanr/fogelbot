package sentiment

import (
	"github.com/grassmudhorses/vader-go"
)

type Analyzer struct {
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Compound returns the compound sentiment score between -1 and 1
func (a *Analyzer) Compound(text string) float64 {
	sentiment := vader.GetSentiment(text)
	return sentiment.Compound
}