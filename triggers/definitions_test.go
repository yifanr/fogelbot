package triggers

import (
	"strings"
	"testing"

	"fogelbot/state"
)

func buildRegistryForTest(clk *fakeClock, cm *state.CooldownManager) *Registry {
	sentiment := &fakeSentiment{score: 0.0}
	lang := &fakeLanguage{nonEnglish: false}
	r := NewRegistry()
	RegisterAll(r, sentiment, lang)
	return r
}

func TestDSLTriggers_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMatch bool
		contains  string
	}{
		{"HoodKing matches michael", "michael is great", true, "hood king"},
		{"HoodKing matches jordan", "jordan rocks", true, "hood king"},
		{"HoodKing matches king", "the king is here", true, "hood king"},
		{"HoodKing no match", "zzz qqq xxx", false, ""},
		{"Odenigbo matches", "odenigbo said hi", true, "odenigbo"},
		{"Odenigbo fuzzy match", "odenbo test", true, "odenigbo"},
		{"LoveHate matches love", "i love this", true, "I know you do"},
		{"LoveHate matches hate", "i hate mondays", true, "I know you do"},
		{"LoveHate matches want", "i want pizza", true, "I know you do"},
		{"Kanye matches", "kanye west", true, "Rough year"},
		{"Food matches eat", "let's eat now", true, "moan-worthy"},
		{"Food matches hungry", "i am hungry", true, "moan-worthy"},
		{"BDSM matches", "bdsm is weird", true, "BDSM test"},
		{"Hockey matches", "hockey game tonight", true, "hockey team"},
		{"Ben matches", "ask ben", true, "Ben always"},
		{"Shame matches", "no shame", true, "shaming"},
		{"Grades matches", "my grade sucks", true, "grade went down"},
		{"IThink matches think", "i think so", true, ""},
		{"IThink matches believe", "i believe it", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clk := newFakeClock()
			cm := state.NewCooldownManager(clk)
			r := buildRegistryForTest(clk, cm)

			// Use high float (0.99) so probability-gated triggers (user, random) don't fire
			ctx := newTestContext(tt.input).
				WithRand(newAlwaysRand(0, 0.99)).
				WithClock(clk).
				WithCooldowns(cm).
				Build()

			resp := r.Process(ctx)
			if tt.wantMatch {
				if resp == "" {
					t.Errorf("expected match for %q, got none", tt.input)
				}
				if tt.contains != "" && !strings.Contains(resp, tt.contains) {
					t.Errorf("expected response to contain %q, got %q", tt.contains, resp)
				}
			} else {
				if resp != "" {
					t.Errorf("expected no match for %q, got %q", tt.input, resp)
				}
			}
		})
	}
}

func TestDSLTriggers_IThinkResponses(t *testing.T) {
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)
	r := buildRegistryForTest(clk, cm)

	// With fixedRand index 0, should get first ThinkResponse
	ctx := newTestContext("i think so").
		WithRand(newAlwaysRand(0, 0.99)).
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	resp := r.Process(ctx)
	if resp != "DING DING DING" {
		t.Errorf("expected 'DING DING DING', got %q", resp)
	}

	// With index 1, should get "Very good, <@user_123>!"
	clk2 := newFakeClock()
	cm2 := state.NewCooldownManager(clk2)
	r2 := buildRegistryForTest(clk2, cm2)

	ctx2 := newTestContext("i think so").
		WithRand(newAlwaysRand(1, 0.99)).
		WithAuthor("user_123", "TestUser").
		WithClock(clk2).
		WithCooldowns(cm2).
		Build()

	resp2 := r2.Process(ctx2)
	if !strings.Contains(resp2, "Very good") || !strings.Contains(resp2, "<@user_123>") {
		t.Errorf("expected 'Very good, <@user_123>!', got %q", resp2)
	}
}

func TestDSLTriggers_UserSpecific_Electroshk(t *testing.T) {
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)
	r := buildRegistryForTest(clk, cm)

	// "yifan" matches fallback regex, low roll -> ElectroshkRare (0.01 threshold)
	ctx := newTestContext("hey yifan").
		WithRand(newFixedRand(nil, []float64{0.005})).
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	resp := r.Process(ctx)
	if resp != "Did you really do the YEEEbowl without me?!?" {
		t.Errorf("expected YEEEbowl, got %q", resp)
	}
}

func TestDSLTriggers_UserSpecific_ElectroshkYEEE(t *testing.T) {
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)
	r := buildRegistryForTest(clk, cm)

	// "yifan" matches, ElectroshkRare misses (roll > 0.01), Electroshk hits (roll < 0.05)
	ctx := newTestContext("hey yifan").
		WithRand(newFixedRand(nil, []float64{0.02, 0.01})).
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	resp := r.Process(ctx)
	if resp != "YEEEEEEEEEEEEEEEEEEEEEE" {
		t.Errorf("expected YEEE, got %q", resp)
	}
}

func TestDSLTriggers_Modriver(t *testing.T) {
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)
	r := buildRegistryForTest(clk, cm)

	// "raul" matches Modriver fallback regex.
	// ElectroshkRare and Electroshk also have User matchers, but "raul" won't match
	// their `\byifan\b` fallback or empty user ID. So the float64 values go:
	// - ElectroshkRare: User check fails, no float consumed
	// - Electroshk: User check fails, no float consumed
	// - Modriver: User check passes (raul matches fallback), consumes float 0.01 < 0.05 -> fires
	ctx := newTestContext("hey raul").
		WithRand(newFixedRand([]int{0}, []float64{0.01})).
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	resp := r.Process(ctx)
	expected := []string{"RAUL WHAT THE HELL?!", "Shut up, RAUL"}
	assertResponseOneOf(t, resp, expected)
}

func TestDSLTriggers_RandomQuote(t *testing.T) {
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)
	r := buildRegistryForTest(clk, cm)

	// Message that doesn't match any keyword/user triggers.
	// The only probability-gated triggers that could fire are:
	// - ElectroshkRare, Electroshk, Modriver: user check fails first, no float consumed
	// - RandomNegative: probability check, consumes float
	// - RandomQuote: probability check, consumes float
	// So floats: [0.99 (RandomNeg miss), 0.005 (RandomQuote hit)]
	ctx := newTestContext("zzz qqq xyz").
		WithRand(newFixedRand([]int{0}, []float64{0.99, 0.005})).
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	resp := r.Process(ctx)
	assertResponseOneOf(t, resp, RandomQuotes)
}
