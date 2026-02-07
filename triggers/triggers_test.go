package triggers

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"fogelbot/language"
	"fogelbot/sentiment"
	"fogelbot/state"
)

func setupContext(content string) *Context {
	return &Context{
		Session: &discordgo.Session{
			State: &discordgo.State{
				Ready: discordgo.Ready{
					User: &discordgo.User{
						ID: "bot_id",
					},
				},
			},
		},
		Message: &discordgo.MessageCreate{
			Message: &discordgo.Message{
				Content: content,
				Author: &discordgo.User{
					ID:       "user_id",
					Username: "TestUser",
				},
				Mentions:  []*discordgo.User{},
				ChannelID: "channel_1",
			},
		},
		Cooldowns: state.NewCooldownManager(),
		Sentiment: sentiment.NewAnalyzer(),
		Language:  language.NewDetector(),
	}
}

func TestFogelTrigger(t *testing.T) {
	trigger := &FogelTrigger{}

	// Test mention by name
	ctx := setupContext("hello fogel")
	resp, matched := trigger.Check(ctx)
	if !matched {
		t.Error("Expected match for 'hello fogel'")
	}
	if resp == "" {
		t.Error("Expected response")
	}

	// Test no mention
	ctx = setupContext("hello world")
	resp, matched = trigger.Check(ctx)
	if matched {
		t.Error("Expected no match for 'hello world'")
	}
}

func TestKeywordTrigger(t *testing.T) {
	triggers := NewKeywordTriggers()
	reg := NewRegistry()
	for _, tr := range triggers {
		reg.Register(tr)
	}

	tests := []struct {
		input    string
		match    bool
		contains string
	}{
		{"I love this", true, "I know you do"},
		{"Kanye West", true, "Rough year"},
		{"playing hockey", true, "hockey team"},
		{"random text", false, ""},
	}

	for _, test := range tests {
		ctx := setupContext(test.input)
		resp := reg.Process(ctx)
		if test.match {
			if resp == "" {
				t.Errorf("Expected match for '%s'", test.input)
			}
			if !strings.Contains(resp, test.contains) {
				t.Errorf("Expected response for '%s' to contain '%s', got '%s'", test.input, test.contains, resp)
			}
		} else {
			if resp != "" {
				t.Errorf("Expected no match for '%s', got '%s'", test.input, resp)
			}
		}
	}
}

func TestLanguageTrigger(t *testing.T) {
	trigger := &LanguageTrigger{}

	// Spanish
	ctx := setupContext("Hola como estas todo bien")
	resp, matched := trigger.Check(ctx)
	if !matched {
		t.Error("Expected match for Spanish text")
	}
	if resp != "SPEAK ENGLISH!!" {
		t.Errorf("Expected 'SPEAK ENGLISH!!', got '%s'", resp)
	}

	// English
	ctx = setupContext("Hello how are you doing today")
	resp, matched = trigger.Check(ctx)
	if matched {
		t.Error("Expected no match for English text")
	}
}
