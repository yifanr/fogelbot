package facts

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// fakeLLM records calls and returns configurable responses.
type fakeLLM struct {
	mu              sync.Mutex
	extractCalls    int
	compactCalls    int
	extractResult   []string
	extractErr      error
	compactResult   []string
	compactErr      error
	generateResult  string
	generateErr     error
}

func (f *fakeLLM) ExtractFacts(_ context.Context, _ string, _ []string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.extractCalls++
	return f.extractResult, f.extractErr
}

func (f *fakeLLM) CompactFacts(_ context.Context, _ string, _ []string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.compactCalls++
	return f.compactResult, f.compactErr
}

func (f *fakeLLM) GenerateResponse(_ context.Context, _ string, _ string, _ []string, _ []string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.generateResult, f.generateErr
}

func newTestCollector(t *testing.T, llmClient *fakeLLM) (*Collector, *fakeClock) {
	t.Helper()
	clk := newFakeClock()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := NewFactStore(path)
	if err != nil {
		t.Fatalf("NewFactStore: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	c := NewCollector(store, llmClient, clk)
	return c, clk
}

func TestCollector_SizeBasedFlush(t *testing.T) {
	fl := &fakeLLM{extractResult: []string{"likes testing"}}
	c, _ := newTestCollector(t, fl)

	// Add BufferSizeLimit messages to trigger a size-based flush
	for i := 0; i < BufferSizeLimit; i++ {
		c.Observe("user1", "alice", fmt.Sprintf("msg %d", i))
	}

	// The flush is async via goroutine, give it a moment
	time.Sleep(50 * time.Millisecond)

	facts, err := c.GetFacts("user1")
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	if len(facts) != 1 || facts[0] != "likes testing" {
		t.Errorf("unexpected facts: %v", facts)
	}

	fl.mu.Lock()
	if fl.extractCalls != 1 {
		t.Errorf("expected 1 extract call, got %d", fl.extractCalls)
	}
	fl.mu.Unlock()
}

func TestCollector_TimeBasedFlush(t *testing.T) {
	fl := &fakeLLM{extractResult: []string{"fact1"}}
	c, clk := newTestCollector(t, fl)

	c.Observe("user1", "alice", "hello")

	// Advance past flush interval
	clk.Advance(2 * time.Minute)

	// Manually trigger what the ticker would do
	c.flushAll()

	facts, err := c.GetFacts("user1")
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	if len(facts) != 1 || facts[0] != "fact1" {
		t.Errorf("unexpected facts: %v", facts)
	}
}

func TestCollector_CompactionTrigger(t *testing.T) {
	// Return enough facts to exceed CompactionThreshold when accumulated
	fl := &fakeLLM{
		extractResult: []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8"},
		compactResult: []string{"compacted1", "compacted2"},
	}
	c, clk := newTestCollector(t, fl)

	// First batch: 8 facts
	for i := 0; i < BufferSizeLimit; i++ {
		c.Observe("user1", "alice", fmt.Sprintf("msg %d", i))
	}
	time.Sleep(50 * time.Millisecond)

	// Second batch: 8 more facts (total 16 > 15 threshold)
	clk.Advance(2 * time.Minute)
	for i := 0; i < BufferSizeLimit; i++ {
		c.Observe("user1", "alice", fmt.Sprintf("msg2 %d", i))
	}
	time.Sleep(50 * time.Millisecond)

	fl.mu.Lock()
	compactCalls := fl.compactCalls
	fl.mu.Unlock()

	if compactCalls < 1 {
		t.Errorf("expected at least 1 compact call, got %d", compactCalls)
	}

	facts, _ := c.GetFacts("user1")
	if len(facts) != 2 || facts[0] != "compacted1" {
		t.Errorf("expected compacted facts, got %v", facts)
	}
}

func TestCollector_LLMError(t *testing.T) {
	fl := &fakeLLM{extractErr: fmt.Errorf("api down")}
	c, _ := newTestCollector(t, fl)

	for i := 0; i < BufferSizeLimit; i++ {
		c.Observe("user1", "alice", "msg")
	}
	time.Sleep(50 * time.Millisecond)

	// Should have no facts stored (error was swallowed)
	facts, err := c.GetFacts("user1")
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	if facts != nil {
		t.Errorf("expected nil facts after LLM error, got %v", facts)
	}
}

func TestCollector_EmptyExtract(t *testing.T) {
	fl := &fakeLLM{extractResult: []string{}}
	c, _ := newTestCollector(t, fl)

	for i := 0; i < BufferSizeLimit; i++ {
		c.Observe("user1", "alice", "msg")
	}
	time.Sleep(50 * time.Millisecond)

	facts, _ := c.GetFacts("user1")
	if facts != nil {
		t.Errorf("expected nil facts for empty extract, got %v", facts)
	}
}

func TestCollector_StartStop(t *testing.T) {
	fl := &fakeLLM{extractResult: []string{"fact1"}}
	c, clk := newTestCollector(t, fl)

	c.Observe("user1", "alice", "hello")
	c.Start()

	// Advance time so the ticker's ShouldFlush check passes
	clk.Advance(2 * time.Minute)

	// Give the ticker a chance to fire
	time.Sleep(200 * time.Millisecond)

	c.Stop()

	facts, _ := c.GetFacts("user1")
	if len(facts) != 1 {
		t.Errorf("expected 1 fact after stop, got %v", facts)
	}
}

func TestCollector_ConcurrentObserve(t *testing.T) {
	fl := &fakeLLM{extractResult: []string{"fact"}}
	c, _ := newTestCollector(t, fl)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Observe(fmt.Sprintf("user%d", n%5), "name", "msg")
		}(i)
	}
	wg.Wait()

	// Flush whatever's left
	c.flushAll()
	time.Sleep(50 * time.Millisecond)
}
