package config

import (
	"testing"
	"time"
)

func TestConstants(t *testing.T) {
	if QuickReplyWindow != 20*time.Second {
		t.Errorf("expected QuickReplyWindow=20s, got %v", QuickReplyWindow)
	}
	if QuickReplyCooldown != 5*time.Minute {
		t.Errorf("expected QuickReplyCooldown=5m, got %v", QuickReplyCooldown)
	}
	if SpeakEnglishCooldown != 5*time.Minute {
		t.Errorf("expected SpeakEnglishCooldown=5m, got %v", SpeakEnglishCooldown)
	}
	if SentimentThreshold != 0.05 {
		t.Errorf("expected SentimentThreshold=0.05, got %v", SentimentThreshold)
	}
}

func TestLoad_FromEnvVars(t *testing.T) {
	t.Setenv("DISCORD_TOKEN", "test-token-123")
	t.Setenv("ELECTROSHK_ID", "elec-id")
	t.Setenv("MODRIVER_ID", "mod-id")

	Load()

	if DiscordToken != "test-token-123" {
		t.Errorf("expected DiscordToken='test-token-123', got %q", DiscordToken)
	}
	if ElectroshkID != "elec-id" {
		t.Errorf("expected ElectroshkID='elec-id', got %q", ElectroshkID)
	}
	if ModriverID != "mod-id" {
		t.Errorf("expected ModriverID='mod-id', got %q", ModriverID)
	}
}
