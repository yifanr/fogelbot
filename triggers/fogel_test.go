package triggers

import (
	"context"
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestFogelTrigger_MentionByName(t *testing.T) {
	trigger := NewFogelTrigger(nil, nil)
	ctx := newTestContext("hello fogel").WithSentiment(0.5).WithRand(newAlwaysRand(0, 0)).Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_MentionByName_ProbabilityMiss(t *testing.T) {
	trigger := NewFogelTrigger(nil, nil)
	ctx := newTestContext("hello fogel").
		WithSentiment(0.5).
		WithRand(newAlwaysRand(0, fogelKeywordProbability)).
		Build()
	assertNotMatched(t, trigger, ctx)
}

func TestFogelTrigger_MentionByAtBot_NoLLM(t *testing.T) {
	trigger := NewFogelTrigger(nil, nil)
	botUser := &discordgo.User{ID: "bot_id"}
	ctx := newTestContext("hey there").
		WithMentions(botUser).
		WithSentiment(0.5).
		WithRand(newAlwaysRand(0, 0)).
		Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_AtMentionBypassesKeywordProbability(t *testing.T) {
	trigger := NewFogelTrigger(nil, nil)
	botUser := &discordgo.User{ID: "bot_id"}
	ctx := newTestContext("hey there").
		WithMentions(botUser).
		WithSentiment(0.5).
		WithRand(newAlwaysRand(0, 0.99)).
		Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_NoMention(t *testing.T) {
	trigger := NewFogelTrigger(nil, nil)
	ctx := newTestContext("hello world").Build()
	assertNotMatched(t, trigger, ctx)
}

func TestFogelTrigger_PositiveSentiment(t *testing.T) {
	trigger := NewFogelTrigger(nil, nil)
	ctx := newTestContext("I love fogel").WithSentiment(0.8).WithRand(newAlwaysRand(0, 0)).Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_NegativeSentiment(t *testing.T) {
	trigger := NewFogelTrigger(nil, nil)
	ctx := newTestContext("I hate fogel").WithSentiment(-0.8).WithRand(newAlwaysRand(0, 0)).Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, NegativeResponses)
}

func TestFogelTrigger_NeutralSentiment(t *testing.T) {
	trigger := NewFogelTrigger(nil, nil)
	ctx := newTestContext("fogel exists").WithSentiment(0.0).WithRand(newAlwaysRand(0, 0)).Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_ResponseSelection(t *testing.T) {
	trigger := NewFogelTrigger(nil, nil)
	// Select the second positive response (index 1)
	ctx := newTestContext("fogel is great").WithSentiment(0.5).WithRand(newAlwaysRand(1, 0)).Build()
	resp := assertMatched(t, trigger, ctx)
	if resp != PositiveResponses[1] {
		t.Errorf("expected %q, got %q", PositiveResponses[1], resp)
	}
}

// --- @mention + LLM tests ---

func TestFogelTrigger_AtMention_LLMFires(t *testing.T) {
	fl := &fakeLLM{generateResult: "Oh, it's you again."}
	fp := &fakeFactProvider{facts: []string{"likes Go"}}
	trigger := NewFogelTrigger(fl, fp)

	botUser := &discordgo.User{ID: "bot_id"}
	// Low roll (0.1 < 0.15) -> LLM fires
	ctx := newTestContext("hey bot").
		WithMentions(botUser).
		WithRand(newAlwaysRand(0, 0.1)).
		Build()

	resp := assertMatched(t, trigger, ctx)
	if resp != "Oh, it's you again." {
		t.Errorf("expected LLM response, got %q", resp)
	}
}

func TestFogelTrigger_AtMention_LLMRollFails(t *testing.T) {
	fl := &fakeLLM{generateResult: "should not see this"}
	fp := &fakeFactProvider{facts: []string{"likes Go"}}
	trigger := NewFogelTrigger(fl, fp)

	botUser := &discordgo.User{ID: "bot_id"}
	// High roll (0.5 >= 0.15) -> LLM skipped, canned response
	ctx := newTestContext("hey bot").
		WithMentions(botUser).
		WithSentiment(0.5).
		WithRand(newFixedRand([]int{0}, []float64{0.5})).
		Build()

	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_AtMention_LLMError_FallsThrough(t *testing.T) {
	fl := &fakeLLM{generateErr: fmt.Errorf("api down")}
	fp := &fakeFactProvider{facts: []string{"likes Go"}}
	trigger := NewFogelTrigger(fl, fp)

	botUser := &discordgo.User{ID: "bot_id"}
	// Low roll -> LLM fires but errors -> falls through to canned
	ctx := newTestContext("hey bot").
		WithMentions(botUser).
		WithSentiment(0.5).
		WithRand(newFixedRand([]int{0}, []float64{0.1})).
		Build()

	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_AtMention_LLMEmptyResponse_FallsThrough(t *testing.T) {
	fl := &fakeLLM{generateResult: ""}
	fp := &fakeFactProvider{facts: nil}
	trigger := NewFogelTrigger(fl, fp)

	botUser := &discordgo.User{ID: "bot_id"}
	ctx := newTestContext("hey bot").
		WithMentions(botUser).
		WithSentiment(0.0).
		WithRand(newFixedRand([]int{0}, []float64{0.1})).
		Build()

	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestFogelTrigger_KeywordOnly_NoLLMAttempt(t *testing.T) {
	// "fogel" keyword without @mention should never try LLM, even with low roll
	callCount := 0
	fl := &countingFakeLLM{onGenerate: func() { callCount++ }}
	fp := &fakeFactProvider{facts: []string{"likes Go"}}
	trigger := NewFogelTrigger(fl, fp)

	ctx := newTestContext("hello fogel").
		WithSentiment(0.5).
		WithRand(newAlwaysRand(0, 0.01)).
		Build()

	assertMatched(t, trigger, ctx)
	if callCount != 0 {
		t.Errorf("expected no LLM calls for keyword-only mention, got %d", callCount)
	}
}

// countingFakeLLM tracks whether GenerateResponse was called.
type countingFakeLLM struct {
	onGenerate func()
}

func (f *countingFakeLLM) ExtractFacts(_ context.Context, _ string, _ []string) ([]string, error) {
	return nil, nil
}

func (f *countingFakeLLM) CompactFacts(_ context.Context, _ string, _ []string) ([]string, error) {
	return nil, nil
}

func (f *countingFakeLLM) GenerateResponse(_ context.Context, _ string, _ string, _ []string, _ []string) (string, error) {
	f.onGenerate()
	return "generated", nil
}
