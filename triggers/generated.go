package triggers

import (
	"context"
	"log"
	"time"

	"fogelbot/llm"
)

const generatedResponseProbability = 0.05

// GeneratedResponseTrigger fires with a 5% chance, generating an LLM response using user facts.
type GeneratedResponseTrigger struct {
	llm          llm.LLM
	factProvider FactProvider
}

// NewGeneratedResponseTrigger creates a new trigger backed by the given LLM and fact provider.
func NewGeneratedResponseTrigger(llmClient llm.LLM, fp FactProvider) *GeneratedResponseTrigger {
	return &GeneratedResponseTrigger{
		llm:          llmClient,
		factProvider: fp,
	}
}

func (t *GeneratedResponseTrigger) Name() string { return "GeneratedResponse" }

func (t *GeneratedResponseTrigger) Check(ctx *Context) (string, bool) {
	if ctx.Rand.Float64() >= generatedResponseProbability {
		return "", false
	}

	facts, err := t.factProvider.GetFacts(ctx.Message.Author.ID)
	if err != nil {
		log.Printf("generated trigger: get facts: %v", err)
		return "", false
	}

	llmCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := t.llm.GenerateResponse(
		llmCtx,
		ctx.Message.Author.Username,
		ctx.Message.Content,
		facts,
		RandomQuotes,
	)
	if err != nil {
		log.Printf("generated trigger: LLM error: %v", err)
		return "", false
	}
	if resp == "" {
		return "", false
	}

	return resp, true
}
