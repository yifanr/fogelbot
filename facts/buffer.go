package facts

import (
	"sync"
	"time"
)

// Clock abstracts time for testability (mirrors triggers.Clock, defined separately to avoid circular imports).
type Clock interface {
	Now() time.Time
}

const (
	BufferSizeLimit = 10
	FlushInterval   = 1 * time.Minute
)

// BufferEntry holds a flushed batch of messages for one user.
type BufferEntry struct {
	UserID   string
	Username string
	Messages []string
}

type userBuffer struct {
	username string
	messages []string
}

// MessageBuffer accumulates per-user messages and supports size- and time-based flushing.
type MessageBuffer struct {
	mu        sync.Mutex
	clock     Clock
	buffers   map[string]*userBuffer
	lastFlush time.Time
}

// NewMessageBuffer creates a new buffer using the given clock.
func NewMessageBuffer(clock Clock) *MessageBuffer {
	return &MessageBuffer{
		clock:     clock,
		buffers:   make(map[string]*userBuffer),
		lastFlush: clock.Now(),
	}
}

// Add appends a message to the user's buffer. Returns true if the buffer hit the size limit.
func (mb *MessageBuffer) Add(userID, username, message string) bool {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	buf, ok := mb.buffers[userID]
	if !ok {
		buf = &userBuffer{username: username}
		mb.buffers[userID] = buf
	}
	buf.messages = append(buf.messages, message)
	return len(buf.messages) >= BufferSizeLimit
}

// ShouldFlush returns true if FlushInterval has elapsed since the last flush.
func (mb *MessageBuffer) ShouldFlush() bool {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	return mb.clock.Now().Sub(mb.lastFlush) >= FlushInterval
}

// FlushAll drains all non-empty buffers and returns them as entries.
func (mb *MessageBuffer) FlushAll() []BufferEntry {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	var entries []BufferEntry
	for userID, buf := range mb.buffers {
		if len(buf.messages) > 0 {
			entries = append(entries, BufferEntry{
				UserID:   userID,
				Username: buf.username,
				Messages: buf.messages,
			})
		}
	}
	mb.buffers = make(map[string]*userBuffer)
	mb.lastFlush = mb.clock.Now()
	return entries
}
