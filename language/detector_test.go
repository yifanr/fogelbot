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
	if d.IsNonEnglish("hi") {
		t.Error("expected short text to not be flagged")
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
