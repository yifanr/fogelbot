package state

import (
	"sync"
	"time"
)

type CooldownManager struct {
	mu                      sync.RWMutex
	lastBotReplyTime        map[string]time.Time
	lastQuickReplyTrigger   map[string]time.Time
	lastSpeakEnglishTrigger map[string]time.Time
}

func NewCooldownManager() *CooldownManager {
	return &CooldownManager{
		lastBotReplyTime:        make(map[string]time.Time),
		lastQuickReplyTrigger:   make(map[string]time.Time),
		lastSpeakEnglishTrigger: make(map[string]time.Time),
	}
}

func (cm *CooldownManager) SetLastBotReply(channelID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.lastBotReplyTime[channelID] = time.Now()
}

func (cm *CooldownManager) GetLastBotReply(channelID string) time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.lastBotReplyTime[channelID]
}

func (cm *CooldownManager) SetLastQuickReplyTrigger(channelID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.lastQuickReplyTrigger[channelID] = time.Now()
}

func (cm *CooldownManager) GetLastQuickReplyTrigger(channelID string) time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.lastQuickReplyTrigger[channelID]
}

func (cm *CooldownManager) SetLastSpeakEnglishTrigger(channelID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.lastSpeakEnglishTrigger[channelID] = time.Now()
}

func (cm *CooldownManager) GetLastSpeakEnglishTrigger(channelID string) time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.lastSpeakEnglishTrigger[channelID]
}
