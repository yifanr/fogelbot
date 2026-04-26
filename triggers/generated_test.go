package triggers

import (
	"context"
	"fmt"
	"testing"
)

type fakeLLM struct {
	generateResult string
	generateErr    error
}

func (f *fakeLLM) ExtractFacts(_ context.Context, _ string, _ []string) ([]string, error) {
	return nil, nil
}

func (f *fakeLLM) CompactFacts(_ context.Context, _ string, _ []string) ([]string, error) {
	return nil, nil
}

func (f *fakeLLM) GenerateResponse(_ context.Context, _ string, _ string, _ []string, _ []string) (string, error) {
	return f.generateResult, f.generateErr
}

type fakeFactProvider struct {
	facts []string
	err   error
}

func (f *fakeFactProvider) GetFacts(_ string) ([]string, error) {
	return f.facts, f.err
}

func TestGeneratedResponseTrigger_ProbabilityGate(t *testing.T) {
	fl := &fakeLLM{generateResult: "response"}
	fp := &fakeFactProvider{facts: []string{"likes Go"}}
	trigger := NewGeneratedResponseTrigger(fl, fp)

	// High roll (0.99) -> should not fire (0.99 >= 0.025)
	ctx := newTestContext("hello").
		WithRand(newAlwaysRand(0, 0.99)).
		Build()

	_, matched := trigger.Check(ctx)
	if matched {
		t.Error("expected no match with high random roll")
	}
}

func TestGeneratedResponseTrigger_Fires(t *testing.T) {
	fl := &fakeLLM{generateResult: "Nobody asked you."}
	fp := &fakeFactProvider{facts: []string{"likes Go"}}
	trigger := NewGeneratedResponseTrigger(fl, fp)

	// Low roll (0.01) -> should fire (0.01 < 0.025)
	ctx := newTestContext("hello").
		WithRand(newAlwaysRand(0, 0.01)).
		Build()

	resp, matched := trigger.Check(ctx)
	if !matched {
		t.Fatal("expected match with low random roll")
	}
	if resp != "Nobody asked you." {
		t.Errorf("unexpected response: %q", resp)
	}
}

func TestGeneratedResponseTrigger_LLMError(t *testing.T) {
	fl := &fakeLLM{generateErr: fmt.Errorf("api down")}
	fp := &fakeFactProvider{facts: []string{"likes Go"}}
	trigger := NewGeneratedResponseTrigger(fl, fp)

	ctx := newTestContext("hello").
		WithRand(newAlwaysRand(0, 0.01)).
		Build()

	_, matched := trigger.Check(ctx)
	if matched {
		t.Error("expected no match when LLM errors")
	}
}

func TestGeneratedResponseTrigger_EmptyResponse(t *testing.T) {
	fl := &fakeLLM{generateResult: ""}
	fp := &fakeFactProvider{facts: []string{"likes Go"}}
	trigger := NewGeneratedResponseTrigger(fl, fp)

	ctx := newTestContext("hello").
		WithRand(newAlwaysRand(0, 0.01)).
		Build()

	_, matched := trigger.Check(ctx)
	if matched {
		t.Error("expected no match with empty response")
	}
}

func TestGeneratedResponseTrigger_FactProviderError(t *testing.T) {
	fl := &fakeLLM{generateResult: "response"}
	fp := &fakeFactProvider{err: fmt.Errorf("db error")}
	trigger := NewGeneratedResponseTrigger(fl, fp)

	ctx := newTestContext("hello").
		WithRand(newAlwaysRand(0, 0.01)).
		Build()

	_, matched := trigger.Check(ctx)
	if matched {
		t.Error("expected no match when fact provider errors")
	}
}

func TestGeneratedResponseTrigger_NoFacts(t *testing.T) {
	fl := &fakeLLM{generateResult: "Whatever."}
	fp := &fakeFactProvider{facts: nil}
	trigger := NewGeneratedResponseTrigger(fl, fp)

	ctx := newTestContext("hello").
		WithRand(newAlwaysRand(0, 0.01)).
		Build()

	resp, matched := trigger.Check(ctx)
	if !matched {
		t.Fatal("expected match even with no facts")
	}
	if resp != "Whatever." {
		t.Errorf("unexpected response: %q", resp)
	}
}
