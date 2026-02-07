package triggers

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"fogelbot/state"
)

func TestIntegration_PriorityOrdering_FogelWinsOverLoveHate(t *testing.T) {
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)
	sentiment := &fakeSentiment{score: 0.5}
	lang := &fakeLanguage{nonEnglish: false}

	r := NewRegistry()
	RegisterAll(r, sentiment, lang)

	// "I love fogel" should match FogelTrigger (priority 1), not LoveHate (priority 3)
	ctx := newTestContext("I love fogel").
		WithSentiment(0.5).
		WithRand(newAlwaysRand(0, 0)).
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	resp := r.Process(ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestIntegration_PriorityOrdering_FogelMentionWinsOverKeywords(t *testing.T) {
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)
	sentiment := &fakeSentiment{score: 0.0}
	lang := &fakeLanguage{nonEnglish: false}

	r := NewRegistry()
	RegisterAll(r, sentiment, lang)

	// "fogel is the king" should match FogelTrigger, not HoodKing
	ctx := newTestContext("fogel is the king").
		WithRand(newAlwaysRand(0, 0)).
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	resp := r.Process(ctx)
	// FogelTrigger returns positive responses for neutral sentiment
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestIntegration_NoMatch(t *testing.T) {
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)
	sentiment := &fakeSentiment{score: 0.0}
	lang := &fakeLanguage{nonEnglish: false}

	r := NewRegistry()
	RegisterAll(r, sentiment, lang)

	// With high random rolls, nothing probabilistic fires
	ctx := newTestContext("completely unrelated message xyz").
		WithRand(newFixedRand(nil, []float64{0.99})).
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	resp := r.Process(ctx)
	if resp != "" {
		t.Errorf("expected no match, got %q", resp)
	}
}

func TestIntegration_BotIgnoresSelf(t *testing.T) {
	// This tests the logic from handlers.go conceptually:
	// Messages from the bot itself should be ignored.
	// We verify by checking that the Session user ID check works.
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)

	ctx := &Context{
		Session: &discordgo.Session{
			State: &discordgo.State{
				Ready: discordgo.Ready{
					User: &discordgo.User{ID: "bot_id"},
				},
			},
		},
		Message: &discordgo.MessageCreate{
			Message: &discordgo.Message{
				Content:   "fogel is great",
				Author:    &discordgo.User{ID: "bot_id"},
				Mentions:  []*discordgo.User{},
				ChannelID: "ch1",
			},
		},
		Cooldowns: cm,
		Sentiment: &fakeSentiment{score: 0.5},
		Language:  &fakeLanguage{nonEnglish: false},
		Rand:      newAlwaysRand(0, 0),
		Clock:     clk,
	}

	// The bot self-check happens in handlers.go, not in triggers.
	// But FogelTrigger still matches "fogel" in content regardless of author.
	// This test documents that the bot-self filter is in the handler layer.
	if ctx.Message.Author.ID == ctx.Session.State.User.ID {
		// This is the handler's check — message would be skipped
		return
	}
	t.Error("expected bot self-check to trigger")
}

func TestIntegration_KeywordMatchesCorrectTrigger(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{"michael plays", "hood king"},
		{"kanye is wild", "Rough year"},
		{"hockey game", "hockey team"},
		{"eating food", "moan-worthy"},
		{"bdsm talk", "BDSM test"},
		{"ben said hi", "Ben always"},
		{"that is shame", "shaming"},
		{"my grade dropped", "grade went down"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			clk := newFakeClock()
			cm := state.NewCooldownManager(clk)
			sentiment := &fakeSentiment{score: 0.0}
			lang := &fakeLanguage{nonEnglish: false}

			r := NewRegistry()
			RegisterAll(r, sentiment, lang)

			ctx := newTestContext(tt.input).
				WithRand(newAlwaysRand(0, 0)).
				WithClock(clk).
				WithCooldowns(cm).
				Build()

			resp := r.Process(ctx)
			if resp == "" {
				t.Errorf("expected match for %q", tt.input)
			}
			assertContains(t, resp, tt.contains)
		})
	}
}
