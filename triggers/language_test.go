package triggers

import (
	"testing"
	"time"

	"fogelbot/state"
)

func TestLanguageTrigger_NonEnglish(t *testing.T) {
	trigger := &LanguageTrigger{}
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)

	ctx := newTestContext("Hola como estas").
		WithNonEnglish(true).
		WithClock(clk).
		WithCooldowns(cm).
		Build()

	resp := assertMatched(t, trigger, ctx)
	if resp != "SPEAK ENGLISH!!" {
		t.Errorf("expected 'SPEAK ENGLISH!!', got %q", resp)
	}
}

func TestLanguageTrigger_English(t *testing.T) {
	trigger := &LanguageTrigger{}
	ctx := newTestContext("Hello how are you").WithNonEnglish(false).Build()
	assertNotMatched(t, trigger, ctx)
}

func TestLanguageTrigger_Cooldown(t *testing.T) {
	trigger := &LanguageTrigger{}
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)

	// First trigger fires
	ctx := newTestContext("Hola").
		WithNonEnglish(true).
		WithClock(clk).
		WithCooldowns(cm).
		Build()
	assertMatched(t, trigger, ctx)

	// Second within cooldown doesn't fire
	clk.Advance(1 * time.Minute)
	ctx2 := newTestContext("Bonjour").
		WithNonEnglish(true).
		WithClock(clk).
		WithCooldowns(cm).
		Build()
	assertNotMatched(t, trigger, ctx2)
}

func TestLanguageTrigger_CooldownExpires(t *testing.T) {
	trigger := &LanguageTrigger{}
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)

	ctx := newTestContext("Hola").
		WithNonEnglish(true).
		WithClock(clk).
		WithCooldowns(cm).
		Build()
	assertMatched(t, trigger, ctx)

	clk.Advance(6 * time.Minute)
	ctx2 := newTestContext("Bonjour").
		WithNonEnglish(true).
		WithClock(clk).
		WithCooldowns(cm).
		Build()
	assertMatched(t, trigger, ctx2)
}

func TestLanguageTrigger_PerChannelIsolation(t *testing.T) {
	trigger := &LanguageTrigger{}
	clk := newFakeClock()
	cm := state.NewCooldownManager(clk)

	// Fire in channel_1
	ctx1 := newTestContext("Hola").
		WithNonEnglish(true).
		WithChannel("channel_1").
		WithClock(clk).
		WithCooldowns(cm).
		Build()
	assertMatched(t, trigger, ctx1)

	// Still fires in channel_2 (different channel)
	ctx2 := newTestContext("Bonjour").
		WithNonEnglish(true).
		WithChannel("channel_2").
		WithClock(clk).
		WithCooldowns(cm).
		Build()
	assertMatched(t, trigger, ctx2)
}
