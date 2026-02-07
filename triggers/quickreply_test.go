package triggers

import (
	"testing"
	"time"

	"fogelbot/state"
)

func TestQuickReply_WithinWindow(t *testing.T) {
	trigger := &QuickReplyTrigger{}
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)

	// Simulate bot replied 10 seconds ago
	cm.SetLastBotReply("channel_1")
	clk.Advance(10 * time.Second)

	ctx := newTestContext("quick response").
		WithSentiment(0.5).
		WithRand(newAlwaysRand(0, 0)).
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, PositiveResponses)
}

func TestQuickReply_OutsideWindow(t *testing.T) {
	trigger := &QuickReplyTrigger{}
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)

	cm.SetLastBotReply("channel_1")
	clk.Advance(30 * time.Second) // Beyond 20s window

	ctx := newTestContext("too slow").
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	assertNotMatched(t, trigger, ctx)
}

func TestQuickReply_NoBotReply(t *testing.T) {
	trigger := &QuickReplyTrigger{}
	ctx := newTestContext("hello").Build()
	assertNotMatched(t, trigger, ctx)
}

func TestQuickReply_CooldownPreventsRepeat(t *testing.T) {
	trigger := &QuickReplyTrigger{}
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)

	// First: bot replies, user replies within window
	cm.SetLastBotReply("channel_1")
	clk.Advance(5 * time.Second)

	ctx := newTestContext("first quick reply").
		WithSentiment(0.5).
		WithRand(newAlwaysRand(0, 0)).
		WithClock(clk).
		WithCooldowns(cm).
		Build()
	assertMatched(t, trigger, ctx)

	// Second: bot replies again, user replies within window, but cooldown is active
	clk.Advance(10 * time.Second)
	cm.SetLastBotReply("channel_1")
	clk.Advance(5 * time.Second)

	ctx2 := newTestContext("second quick reply").
		WithSentiment(0.5).
		WithRand(newAlwaysRand(0, 0)).
		WithClock(clk).
		WithCooldowns(cm).
		Build()
	assertNotMatched(t, trigger, ctx2)
}

func TestQuickReply_CooldownExpires(t *testing.T) {
	trigger := &QuickReplyTrigger{}
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)

	// First trigger
	cm.SetLastBotReply("channel_1")
	clk.Advance(5 * time.Second)
	ctx := newTestContext("quick").
		WithSentiment(0.5).
		WithRand(newAlwaysRand(0, 0)).
		WithClock(clk).
		WithCooldowns(cm).
		Build()
	assertMatched(t, trigger, ctx)

	// Advance past cooldown
	clk.Advance(6 * time.Minute)
	cm.SetLastBotReply("channel_1")
	clk.Advance(5 * time.Second)

	ctx2 := newTestContext("quick again").
		WithSentiment(0.5).
		WithRand(newAlwaysRand(0, 0)).
		WithClock(clk).
		WithCooldowns(cm).
		Build()
	assertMatched(t, trigger, ctx2)
}

func TestQuickReply_NegativeSentiment(t *testing.T) {
	trigger := &QuickReplyTrigger{}
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)

	cm.SetLastBotReply("channel_1")
	clk.Advance(5 * time.Second)

	ctx := newTestContext("I hate this").
		WithSentiment(-0.8).
		WithRand(newAlwaysRand(0, 0)).
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, NegativeResponses)
}
