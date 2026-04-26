package triggers

import (
	"testing"
)

func TestRandomNegative_LowRollNegativeSentiment(t *testing.T) {
	trigger := &RandomNegativeTrigger{}
	// Roll 0.01 < 0.025 threshold, negative sentiment
	ctx := newTestContext("this is terrible").
		WithSentiment(-0.8).
		WithRand(newFixedRand([]int{0}, []float64{0.01})).
		Build()
	resp := assertMatched(t, trigger, ctx)
	assertResponseOneOf(t, resp, NegativeResponses)
}

func TestRandomNegative_HighRollDoesNotFire(t *testing.T) {
	trigger := &RandomNegativeTrigger{}
	// Roll 0.10 >= 0.025 threshold
	ctx := newTestContext("this is terrible").
		WithSentiment(-0.8).
		WithRand(newFixedRand([]int{0}, []float64{0.10})).
		Build()
	assertNotMatched(t, trigger, ctx)
}

func TestRandomNegative_NonNegativeSentiment(t *testing.T) {
	trigger := &RandomNegativeTrigger{}
	// Low roll but positive sentiment
	ctx := newTestContext("this is great").
		WithSentiment(0.5).
		WithRand(newFixedRand([]int{0}, []float64{0.01})).
		Build()
	assertNotMatched(t, trigger, ctx)
}

func TestRandomNegative_NeutralSentiment(t *testing.T) {
	trigger := &RandomNegativeTrigger{}
	// Low roll but neutral sentiment (above -0.05)
	ctx := newTestContext("okay").
		WithSentiment(0.0).
		WithRand(newFixedRand([]int{0}, []float64{0.01})).
		Build()
	assertNotMatched(t, trigger, ctx)
}

func TestRandomNegative_BoundaryRoll(t *testing.T) {
	trigger := &RandomNegativeTrigger{}
	// Roll exactly at 0.025 should NOT fire (< 0.025 required)
	ctx := newTestContext("bad").
		WithSentiment(-0.8).
		WithRand(newFixedRand([]int{0}, []float64{0.025})).
		Build()
	assertNotMatched(t, trigger, ctx)
}

func TestRandomNegative_ResponseSelection(t *testing.T) {
	trigger := &RandomNegativeTrigger{}
	// Select second response (index 1)
	ctx := newTestContext("terrible").
		WithSentiment(-0.8).
		WithRand(newFixedRand([]int{1}, []float64{0.01})).
		Build()
	resp := assertMatched(t, trigger, ctx)
	if resp != NegativeResponses[1] {
		t.Errorf("expected %q, got %q", NegativeResponses[1], resp)
	}
}
