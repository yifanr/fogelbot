package language

import (
	"testing"
)

func TestDetector_EnglishNotFlagged(t *testing.T) {
	d := NewDetector()
	if d.IsNonEnglish("Hello how are you doing today") {
		t.Error("expected English text to not be flagged")
	}
}

func TestDetector_SpanishFlagged(t *testing.T) {
	d := NewDetector()
	if !d.IsNonEnglish("Hola como estas todo bien amigo") {
		t.Error("expected Spanish text to be flagged")
	}
}

func TestDetector_ShortTextNotFlagged(t *testing.T) {
	d := NewDetector()
	for _, input := range []string{"hi", "test", "???"} {
		if d.IsNonEnglish(input) {
			t.Errorf("expected short or low-signal text %q to not be flagged", input)
		}
	}
}

func TestDetector_EmptyTextNotFlagged(t *testing.T) {
	d := NewDetector()
	if d.IsNonEnglish("") {
		t.Error("expected empty text to not be flagged")
	}
}

func TestDetector_WhitespaceNotFlagged(t *testing.T) {
	d := NewDetector()
	if d.IsNonEnglish("   ") {
		t.Error("expected whitespace to not be flagged")
	}
}
