package state

import (
	"sync"
	"testing"
	"time"
)

type testClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *testClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *testClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func newTestClock() *testClock {
	return &testClock{now: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}
}

func TestCooldownManager_BotReply(t *testing.T) {
	clk := newTestClock()
	cm := NewCooldownManager(clk)

	cm.SetLastBotReply("ch1")
	got := cm.GetLastBotReply("ch1")
	if got != clk.Now() {
		t.Errorf("expected %v, got %v", clk.Now(), got)
	}

	// Different channel returns zero
	got2 := cm.GetLastBotReply("ch2")
	if !got2.IsZero() {
		t.Errorf("expected zero time for unknown channel, got %v", got2)
	}
}

func TestCooldownManager_QuickReplyTrigger(t *testing.T) {
	clk := newTestClock()
	cm := NewCooldownManager(clk)

	cm.SetLastQuickReplyTrigger("ch1")
	got := cm.GetLastQuickReplyTrigger("ch1")
	if got != clk.Now() {
		t.Errorf("expected %v, got %v", clk.Now(), got)
	}
}

func TestCooldownManager_SpeakEnglishTrigger(t *testing.T) {
	clk := newTestClock()
	cm := NewCooldownManager(clk)

	cm.SetLastSpeakEnglishTrigger("ch1")
	got := cm.GetLastSpeakEnglishTrigger("ch1")
	if got != clk.Now() {
		t.Errorf("expected %v, got %v", clk.Now(), got)
	}
}

func TestCooldownManager_PerChannelIsolation(t *testing.T) {
	clk := newTestClock()
	cm := NewCooldownManager(clk)

	cm.SetLastBotReply("ch1")
	clk.Advance(1 * time.Second)
	cm.SetLastBotReply("ch2")

	t1 := cm.GetLastBotReply("ch1")
	t2 := cm.GetLastBotReply("ch2")
	if t1 == t2 {
		t.Error("per-channel times should be different")
	}
}

func TestCooldownManager_CheckAndSet_FirstCall(t *testing.T) {
	clk := newTestClock()
	cm := NewCooldownManager(clk)

	ok := cm.CheckAndSet("key1", 5*time.Minute)
	if !ok {
		t.Error("first CheckAndSet should return true")
	}
}

func TestCooldownManager_CheckAndSet_WithinCooldown(t *testing.T) {
	clk := newTestClock()
	cm := NewCooldownManager(clk)

	cm.CheckAndSet("key1", 5*time.Minute)
	clk.Advance(1 * time.Minute)

	ok := cm.CheckAndSet("key1", 5*time.Minute)
	if ok {
		t.Error("CheckAndSet within cooldown should return false")
	}
}

func TestCooldownManager_CheckAndSet_AfterCooldown(t *testing.T) {
	clk := newTestClock()
	cm := NewCooldownManager(clk)

	cm.CheckAndSet("key1", 5*time.Minute)
	clk.Advance(6 * time.Minute)

	ok := cm.CheckAndSet("key1", 5*time.Minute)
	if !ok {
		t.Error("CheckAndSet after cooldown should return true")
	}
}

func TestCooldownManager_CheckAndSet_DifferentKeys(t *testing.T) {
	clk := newTestClock()
	cm := NewCooldownManager(clk)

	cm.CheckAndSet("key1", 5*time.Minute)

	// Different key should be independent
	ok := cm.CheckAndSet("key2", 5*time.Minute)
	if !ok {
		t.Error("different key should be independent")
	}
}

func TestCooldownManager_ConcurrentAccess(t *testing.T) {
	clk := newTestClock()
	cm := NewCooldownManager(clk)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(ch string) {
			defer wg.Done()
			cm.SetLastBotReply(ch)
			cm.GetLastBotReply(ch)
			cm.SetLastQuickReplyTrigger(ch)
			cm.GetLastQuickReplyTrigger(ch)
			cm.SetLastSpeakEnglishTrigger(ch)
			cm.GetLastSpeakEnglishTrigger(ch)
			cm.CheckAndSet("key:"+ch, time.Minute)
		}("ch" + string(rune('0'+i%10)))
	}
	wg.Wait()
}
