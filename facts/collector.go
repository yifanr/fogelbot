package facts

import (
	"context"
	"log"
	"sync"
	"time"

	"fogelbot/llm"
)

const (
	CompactionThreshold = 15
	tickInterval        = 10 * time.Second
)

// Collector orchestrates message buffering, fact extraction, and storage.
type Collector struct {
	buffer *MessageBuffer
	store  *FactStore
	llm    llm.LLM
	clock  Clock
	done   chan struct{}
	wg     sync.WaitGroup
}

// NewCollector creates a Collector wired to the given store, LLM, and clock.
func NewCollector(store *FactStore, llmClient llm.LLM, clock Clock) *Collector {
	return &Collector{
		buffer: NewMessageBuffer(clock),
		store:  store,
		llm:    llmClient,
		clock:  clock,
		done:   make(chan struct{}),
	}
}

// Start launches a background goroutine that periodically flushes the buffer.
func (c *Collector) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(tickInterval)
		defer ticker.Stop()
		for {
			select {
			case <-c.done:
				c.flushAll()
				return
			case <-ticker.C:
				if c.buffer.ShouldFlush() {
					c.flushAll()
				}
			}
		}
	}()
}

// Stop signals the background goroutine to stop and waits for it to finish.
func (c *Collector) Stop() {
	close(c.done)
	c.wg.Wait()
}

// Observe records a user message. If the buffer hits the size limit, flushes asynchronously.
func (c *Collector) Observe(userID, username, message string) {
	if c.buffer.Add(userID, username, message) {
		go c.flushAll()
	}
}

// GetFacts returns the stored facts for a user.
func (c *Collector) GetFacts(userID string) ([]string, error) {
	return c.store.GetFacts(userID)
}

func (c *Collector) flushAll() {
	entries := c.buffer.FlushAll()
	for _, entry := range entries {
		c.processBatch(entry)
	}
}

func (c *Collector) processBatch(entry BufferEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	facts, err := c.llm.ExtractFacts(ctx, entry.Username, entry.Messages)
	if err != nil {
		log.Printf("fact extraction failed for %s: %v", entry.Username, err)
		return
	}
	if len(facts) == 0 {
		return
	}

	total, err := c.store.AppendFacts(entry.UserID, facts)
	if err != nil {
		log.Printf("fact storage failed for %s: %v", entry.Username, err)
		return
	}

	if total > CompactionThreshold {
		c.compact(entry.UserID, entry.Username)
	}
}

func (c *Collector) compact(userID, username string) {
	allFacts, err := c.store.GetFacts(userID)
	if err != nil {
		log.Printf("compact read failed for %s: %v", username, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	compacted, err := c.llm.CompactFacts(ctx, username, allFacts)
	if err != nil {
		log.Printf("compact LLM call failed for %s: %v", username, err)
		return
	}

	if err := c.store.SetFacts(userID, compacted); err != nil {
		log.Printf("compact store failed for %s: %v", username, err)
	}
}
