package sentiment

import (
	"testing"
)

func TestAnalyzer_PositiveText(t *testing.T) {
	a := NewAnalyzer()
	score := a.Compound("This is wonderful and amazing!")
	if score <= 0 {
		t.Errorf("expected positive score for positive text, got %f", score)
	}
}

func TestAnalyzer_NegativeText(t *testing.T) {
	a := NewAnalyzer()
	score := a.Compound("This is terrible and horrible!")
	if score >= 0 {
		t.Errorf("expected negative score for negative text, got %f", score)
	}
}

func TestAnalyzer_NeutralText(t *testing.T) {
	a := NewAnalyzer()
	score := a.Compound("The sky is blue")
	if score < -0.5 || score > 0.5 {
		t.Errorf("expected roughly neutral score, got %f", score)
	}
}

func TestAnalyzer_EmptyText(t *testing.T) {
	a := NewAnalyzer()
	score := a.Compound("")
	if score != 0 {
		t.Errorf("expected 0 for empty text, got %f", score)
	}
}
