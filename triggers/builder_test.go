package triggers

import (
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"fogelbot/state"
)

func TestBuilder_Match(t *testing.T) {
	r := NewRegistry()
	r.On("Test").Match(`\bhello\b`).Respond("world")

	ctx := newTestContext("hello there").Build()
	resp := r.Process(ctx)
	if resp != "world" {
		t.Errorf("expected 'world', got %q", resp)
	}

	ctx2 := newTestContext("goodbye").Build()
	resp2 := r.Process(ctx2)
	if resp2 != "" {
		t.Errorf("expected no match, got %q", resp2)
	}
}

func TestBuilder_MatchCaseInsensitive(t *testing.T) {
	r := NewRegistry()
	r.On("Test").Match(`\bhello\b`).Respond("world")

	ctx := newTestContext("HELLO there").Build()
	resp := r.Process(ctx)
	if resp != "world" {
		t.Errorf("expected 'world', got %q (case insensitive match failed)", resp)
	}
}

func TestBuilder_User_ByAuthorID(t *testing.T) {
	r := NewRegistry()
	r.On("Test").User("target_user", `\bfallback\b`).Respond("matched user")

	ctx := newTestContext("anything").WithAuthor("target_user", "Target").Build()
	resp := r.Process(ctx)
	if resp != "matched user" {
		t.Errorf("expected 'matched user', got %q", resp)
	}
}

func TestBuilder_User_ByMention(t *testing.T) {
	r := NewRegistry()
	r.On("Test").User("target_user", `\bfallback\b`).Respond("matched mention")

	ctx := newTestContext("hey @target").
		WithAuthor("other_user", "Other").
		WithMentions(makeUser("target_user")).
		Build()
	resp := r.Process(ctx)
	if resp != "matched mention" {
		t.Errorf("expected 'matched mention', got %q", resp)
	}
}

func TestBuilder_User_ByFallbackRegex(t *testing.T) {
	r := NewRegistry()
	r.On("Test").User("target_user", `\bfallback\b`).Respond("matched fallback")

	ctx := newTestContext("some fallback text").
		WithAuthor("other_user", "Other").
		Build()
	resp := r.Process(ctx)
	if resp != "matched fallback" {
		t.Errorf("expected 'matched fallback', got %q", resp)
	}
}

func TestBuilder_User_NoMatch(t *testing.T) {
	r := NewRegistry()
	r.On("Test").User("target_user", `\bfallback\b`).Respond("matched")

	ctx := newTestContext("nothing here").
		WithAuthor("other_user", "Other").
		Build()
	resp := r.Process(ctx)
	if resp != "" {
		t.Errorf("expected no match, got %q", resp)
	}
}

func TestBuilder_Probability_Fires(t *testing.T) {
	r := NewRegistry()
	r.On("Test").Probability(0.10).Respond("lucky")

	// Roll 0.05 < 0.10 threshold
	ctx := newTestContext("anything").
		WithRand(newFixedRand(nil, []float64{0.05})).
		Build()
	resp := r.Process(ctx)
	if resp != "lucky" {
		t.Errorf("expected 'lucky', got %q", resp)
	}
}

func TestBuilder_Probability_DoesNotFire(t *testing.T) {
	r := NewRegistry()
	r.On("Test").Probability(0.10).Respond("lucky")

	// Roll 0.10 >= 0.10 threshold
	ctx := newTestContext("anything").
		WithRand(newFixedRand(nil, []float64{0.10})).
		Build()
	resp := r.Process(ctx)
	if resp != "" {
		t.Errorf("expected no match, got %q", resp)
	}
}

func TestBuilder_Probability_BoundaryJustBelow(t *testing.T) {
	r := NewRegistry()
	r.On("Test").Probability(0.02).Respond("edge")

	// Roll 0.019 < 0.02
	ctx := newTestContext("x").
		WithRand(newFixedRand(nil, []float64{0.019})).
		Build()
	resp := r.Process(ctx)
	if resp != "edge" {
		t.Errorf("expected 'edge', got %q", resp)
	}
}

func TestBuilder_Probability_BoundaryExact(t *testing.T) {
	r := NewRegistry()
	r.On("Test").Probability(0.02).Respond("edge")

	// Roll 0.02 >= 0.02 should NOT fire
	ctx := newTestContext("x").
		WithRand(newFixedRand(nil, []float64{0.02})).
		Build()
	resp := r.Process(ctx)
	if resp != "" {
		t.Errorf("expected no match, got %q", resp)
	}
}

func TestBuilder_Sentiment_Positive(t *testing.T) {
	r := NewRegistry()
	r.On("Test").Sentiment("positive").Respond("positive!")

	ctx := newTestContext("great").WithSentiment(0.5).Build()
	if r.Process(ctx) != "positive!" {
		t.Error("positive sentiment should match")
	}

	ctx2 := newTestContext("bad").WithSentiment(-0.5).Build()
	if r.Process(ctx2) != "" {
		t.Error("negative sentiment should not match positive filter")
	}
}

func TestBuilder_Sentiment_Negative(t *testing.T) {
	r := NewRegistry()
	r.On("Test").Sentiment("negative").Respond("negative!")

	ctx := newTestContext("terrible").WithSentiment(-0.5).Build()
	if r.Process(ctx) != "negative!" {
		t.Error("negative sentiment should match")
	}

	ctx2 := newTestContext("great").WithSentiment(0.5).Build()
	if r.Process(ctx2) != "" {
		t.Error("positive sentiment should not match negative filter")
	}
}

func TestBuilder_Cooldown(t *testing.T) {
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)

	r := NewRegistry()
	r.On("Test").Match(`\btest\b`).Cooldown(5 * time.Minute).Respond("cooled")

	ctx := newTestContext("test").WithClock(clk).WithCooldowns(cm).Build()
	resp := r.Process(ctx)
	if resp != "cooled" {
		t.Errorf("first should fire, got %q", resp)
	}

	// Within cooldown
	clk.Advance(1 * time.Minute)
	ctx2 := newTestContext("test").WithClock(clk).WithCooldowns(cm).Build()
	resp2 := r.Process(ctx2)
	if resp2 != "" {
		t.Errorf("within cooldown should not fire, got %q", resp2)
	}

	// After cooldown
	clk.Advance(5 * time.Minute)
	ctx3 := newTestContext("test").WithClock(clk).WithCooldowns(cm).Build()
	resp3 := r.Process(ctx3)
	if resp3 != "cooled" {
		t.Errorf("after cooldown should fire, got %q", resp3)
	}
}

func TestBuilder_RespondOneOf(t *testing.T) {
	r := NewRegistry()
	r.On("Test").Match(`\bpick\b`).RespondOneOf("a", "b", "c")

	// Index 1 → "b"
	ctx := newTestContext("pick me").WithRand(newAlwaysRand(1, 0)).Build()
	resp := r.Process(ctx)
	if resp != "b" {
		t.Errorf("expected 'b', got %q", resp)
	}

	// Index 2 → "c"
	ctx2 := newTestContext("pick me").WithRand(newAlwaysRand(2, 0)).Build()
	resp2 := r.Process(ctx2)
	if resp2 != "c" {
		t.Errorf("expected 'c', got %q", resp2)
	}
}

func TestBuilder_MentionSubstitution(t *testing.T) {
	r := NewRegistry()
	r.On("Test").Match(`\bfood\b`).Respond("%s, thats moan-worthy")

	ctx := newTestContext("food").WithAuthor("123", "TestUser").Build()
	resp := r.Process(ctx)
	if !strings.Contains(resp, "thats moan-worthy") {
		t.Errorf("expected mention substitution, got %q", resp)
	}
	// The mention format is <@123>
	if !strings.Contains(resp, "<@123>") {
		t.Errorf("expected user mention <@123> in response, got %q", resp)
	}
}

func TestBuilder_MatchAndProbabilityCombined(t *testing.T) {
	r := NewRegistry()
	r.On("Test").Match(`\bpizza\b`).Probability(0.5).Respond("yum")

	// Match + low roll → fire
	ctx := newTestContext("pizza time").WithRand(newFixedRand(nil, []float64{0.1})).Build()
	if r.Process(ctx) != "yum" {
		t.Error("match + probability below threshold should fire")
	}

	// Match + high roll → no fire
	ctx2 := newTestContext("pizza time").WithRand(newFixedRand(nil, []float64{0.9})).Build()
	if r.Process(ctx2) != "" {
		t.Error("match + probability above threshold should not fire")
	}

	// No match → no fire regardless of roll
	ctx3 := newTestContext("burger time").WithRand(newFixedRand(nil, []float64{0.1})).Build()
	if r.Process(ctx3) != "" {
		t.Error("no match should not fire")
	}
}

func TestBuilder_UserAndProbabilityCombined(t *testing.T) {
	r := NewRegistry()
	r.On("Test").User("u1", `\bname\b`).Probability(0.05).Respond("gotcha")

	// User match + low roll
	ctx := newTestContext("x").WithAuthor("u1", "U").WithRand(newFixedRand(nil, []float64{0.01})).Build()
	if r.Process(ctx) != "gotcha" {
		t.Error("user match + low roll should fire")
	}

	// User match + high roll
	ctx2 := newTestContext("x").WithAuthor("u1", "U").WithRand(newFixedRand(nil, []float64{0.9})).Build()
	if r.Process(ctx2) != "" {
		t.Error("user match + high roll should not fire")
	}
}

// helper to make a discordgo.User pointer
func makeUser(id string) *discordgo.User {
	return &discordgo.User{ID: id}
}
