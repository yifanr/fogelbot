package triggers

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestFogelTrigger_MentionByName(t *testing.T) {
	trigger := &FogelTrigger{}
	ctx := newTestContext("hello fogel").WithSentiment(0.5).WithRand(newAlwaysRand(0, 0)).Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_MentionByAtBot(t *testing.T) {
	trigger := &FogelTrigger{}
	botUser := &discordgo.User{ID: "bot_id"}
	ctx := newTestContext("hey there").
		WithMentions(botUser).
		WithSentiment(0.5).
		WithRand(newAlwaysRand(0, 0)).
		Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_NoMention(t *testing.T) {
	trigger := &FogelTrigger{}
	ctx := newTestContext("hello world").Build()
	assertNotMatched(t, trigger, ctx)
}

func TestFogelTrigger_PositiveSentiment(t *testing.T) {
	trigger := &FogelTrigger{}
	ctx := newTestContext("I love fogel").WithSentiment(0.8).WithRand(newAlwaysRand(0, 0)).Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_NegativeSentiment(t *testing.T) {
	trigger := &FogelTrigger{}
	ctx := newTestContext("I hate fogel").WithSentiment(-0.8).WithRand(newAlwaysRand(0, 0)).Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, NegativeResponses)
}

func TestFogelTrigger_NeutralSentiment(t *testing.T) {
	trigger := &FogelTrigger{}
	ctx := newTestContext("fogel exists").WithSentiment(0.0).WithRand(newAlwaysRand(0, 0)).Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_ResponseSelection(t *testing.T) {
	trigger := &FogelTrigger{}
	// Select the second positive response (index 1)
	ctx := newTestContext("fogel is great").WithSentiment(0.5).WithRand(newAlwaysRand(1, 0)).Build()
	resp := assertMatched(t, trigger, ctx)
	if resp != PositiveResponses[1] {
		t.Errorf("expected %q, got %q", PositiveResponses[1], resp)
	}
}
