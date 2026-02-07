package facts

import (
	"sync"
	"testing"
	"time"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (f *fakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *fakeClock) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}
}

func TestBuffer_SizeThreshold(t *testing.T) {
	clk := newFakeClock()
	mb := NewMessageBuffer(clk)

	for i := 0; i < BufferSizeLimit-1; i++ {
		hit := mb.Add("user1", "alice", "msg")
		if hit {
			t.Fatalf("should not hit threshold at message %d", i+1)
		}
	}

	hit := mb.Add("user1", "alice", "msg10")
	if !hit {
		t.Fatal("expected threshold hit at message 10")
	}
}

func TestBuffer_FlushAll(t *testing.T) {
	clk := newFakeClock()
	mb := NewMessageBuffer(clk)

	mb.Add("user1", "alice", "hello")
	mb.Add("user2", "bob", "world")
	mb.Add("user1", "alice", "again")

	entries := mb.FlushAll()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	found := map[string]int{}
	for _, e := range entries {
		found[e.UserID] = len(e.Messages)
	}
	if found["user1"] != 2 {
		t.Errorf("expected 2 messages for user1, got %d", found["user1"])
	}
	if found["user2"] != 1 {
		t.Errorf("expected 1 message for user2, got %d", found["user2"])
	}

	// Flush again should be empty
	entries = mb.FlushAll()
	if len(entries) != 0 {
		t.Errorf("expected empty flush, got %d entries", len(entries))
	}
}

func TestBuffer_ShouldFlush_TimeThreshold(t *testing.T) {
	clk := newFakeClock()
	mb := NewMessageBuffer(clk)

	if mb.ShouldFlush() {
		t.Fatal("should not flush immediately after creation")
	}

	clk.Advance(30 * time.Second)
	if mb.ShouldFlush() {
		t.Fatal("should not flush after 30s")
	}

	clk.Advance(31 * time.Second)
	if !mb.ShouldFlush() {
		t.Fatal("should flush after 61s")
	}
}

func TestBuffer_FlushResetsTimer(t *testing.T) {
	clk := newFakeClock()
	mb := NewMessageBuffer(clk)

	clk.Advance(2 * time.Minute)
	mb.FlushAll()

	if mb.ShouldFlush() {
		t.Fatal("should not need flush right after flushing")
	}

	clk.Advance(FlushInterval)
	if !mb.ShouldFlush() {
		t.Fatal("should need flush after interval")
	}
}

func TestBuffer_ConcurrentAccess(t *testing.T) {
	clk := newFakeClock()
	mb := NewMessageBuffer(clk)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			mb.Add("user1", "alice", "msg")
		}(i)
	}
	wg.Wait()

	entries := mb.FlushAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if len(entries[0].Messages) != 100 {
		t.Errorf("expected 100 messages, got %d", len(entries[0].Messages))
	}
}
