package facts

import (
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *FactStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := NewFactStore(path)
	if err != nil {
		t.Fatalf("NewFactStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestFactStore_RoundTrip(t *testing.T) {
	s := newTestStore(t)

	facts, err := s.GetFacts("user1")
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	if facts != nil {
		t.Errorf("expected nil for unknown user, got %v", facts)
	}

	total, err := s.AppendFacts("user1", []string{"likes Go", "lives in SF"})
	if err != nil {
		t.Fatalf("AppendFacts: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}

	facts, err = s.GetFacts("user1")
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	if len(facts) != 2 || facts[0] != "likes Go" || facts[1] != "lives in SF" {
		t.Errorf("unexpected facts: %v", facts)
	}
}

func TestFactStore_AppendAccumulates(t *testing.T) {
	s := newTestStore(t)

	s.AppendFacts("user1", []string{"fact1"})
	s.AppendFacts("user1", []string{"fact2", "fact3"})

	facts, err := s.GetFacts("user1")
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	if len(facts) != 3 {
		t.Errorf("expected 3 facts, got %d", len(facts))
	}
}

func TestFactStore_SetFactsReplaces(t *testing.T) {
	s := newTestStore(t)

	s.AppendFacts("user1", []string{"old1", "old2", "old3"})
	err := s.SetFacts("user1", []string{"compacted"})
	if err != nil {
		t.Fatalf("SetFacts: %v", err)
	}

	facts, err := s.GetFacts("user1")
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	if len(facts) != 1 || facts[0] != "compacted" {
		t.Errorf("expected [compacted], got %v", facts)
	}
}

func TestFactStore_MultipleUsers(t *testing.T) {
	s := newTestStore(t)

	s.AppendFacts("alice", []string{"fact-a"})
	s.AppendFacts("bob", []string{"fact-b"})

	factsA, _ := s.GetFacts("alice")
	factsB, _ := s.GetFacts("bob")

	if len(factsA) != 1 || factsA[0] != "fact-a" {
		t.Errorf("alice: unexpected facts %v", factsA)
	}
	if len(factsB) != 1 || factsB[0] != "fact-b" {
		t.Errorf("bob: unexpected facts %v", factsB)
	}
}

func TestFactStore_EmptyAppend(t *testing.T) {
	s := newTestStore(t)

	total, err := s.AppendFacts("user1", []string{})
	if err != nil {
		t.Fatalf("AppendFacts: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
}
