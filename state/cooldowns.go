package state

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type CooldownManager struct {
	mu                      sync.RWMutex
	clock                   Clock
	lastBotReplyTime        map[string]time.Time
	lastQuickReplyTrigger   map[string]time.Time
	lastSpeakEnglishTrigger map[string]time.Time
	generic                 map[string]time.Time
}

func NewCooldownManager(clock Clock) *CooldownManager {
	return &CooldownManager{
		clock:                   clock,
		lastBotReplyTime:        make(map[string]time.Time),
		lastQuickReplyTrigger:   make(map[string]time.Time),
		lastSpeakEnglishTrigger: make(map[string]time.Time),
		generic:                 make(map[string]time.Time),
	}
}

func (cm *CooldownManager) SetLastBotReply(channelID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.lastBotReplyTime[channelID] = cm.clock.Now()
}

func (cm *CooldownManager) GetLastBotReply(channelID string) time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.lastBotReplyTime[channelID]
}

func (cm *CooldownManager) SetLastQuickReplyTrigger(channelID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.lastQuickReplyTrigger[channelID] = cm.clock.Now()
}

func (cm *CooldownManager) GetLastQuickReplyTrigger(channelID string) time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.lastQuickReplyTrigger[channelID]
}

func (cm *CooldownManager) SetLastSpeakEnglishTrigger(channelID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.lastSpeakEnglishTrigger[channelID] = cm.clock.Now()
}

func (cm *CooldownManager) GetLastSpeakEnglishTrigger(channelID string) time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.lastSpeakEnglishTrigger[channelID]
}

// CheckAndSet checks if a cooldown with the given key has expired. If so, it
// sets the cooldown to now and returns true (meaning "you may fire"). If the
// cooldown is still active, it returns false.
func (cm *CooldownManager) CheckAndSet(key string, duration time.Duration) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	last, ok := cm.generic[key]
	now := cm.clock.Now()
	if ok && now.Sub(last) < duration {
		return false
	}
	cm.generic[key] = now
	return true
}
