package triggers

import (
	"context"
	"log"
	"strings"
	"time"

	"fogelbot/config"
	"fogelbot/llm"
)

type FogelTrigger struct {
	llm          llm.LLM
	factProvider FactProvider
}

func NewFogelTrigger(llmClient llm.LLM, fp FactProvider) *FogelTrigger {
	return &FogelTrigger{llm: llmClient, factProvider: fp}
}

func (t *FogelTrigger) Name() string {
	return "FogelMention"
}

func (t *FogelTrigger) Check(ctx *Context) (string, bool) {
	contentLower := strings.ToLower(ctx.Message.Content)
	keywordMatch := strings.Contains(contentLower, "fogel")

	atMentioned := false
	for _, user := range ctx.Message.Mentions {
		if user.ID == ctx.Session.State.User.ID {
			atMentioned = true
			break
		}
	}

	if !keywordMatch && !atMentioned {
		return "", false
	}

	if keywordMatch && !atMentioned && ctx.Rand.Float64() >= fogelKeywordProbability {
		return "", false
	}

	// Direct @mentions can take a low-probability LLM-generated response path.
	if atMentioned && t.llm != nil && t.factProvider != nil {
		if ctx.Rand.Float64() < mentionGeneratedProbability {
			if resp := t.tryGenerate(ctx); resp != "" {
				return resp, true
			}
		}
	}

	// Fall through to canned sentiment-based responses
	score := ctx.Sentiment.Compound(ctx.Message.Content)
	if score >= config.SentimentThreshold {
		return PositiveResponses[ctx.Rand.Intn(len(PositiveResponses))], true
	} else if score <= -config.SentimentThreshold {
		return NegativeResponses[ctx.Rand.Intn(len(NegativeResponses))], true
	} else {
		return PositiveResponses[ctx.Rand.Intn(len(PositiveResponses))], true
	}
}

func (t *FogelTrigger) tryGenerate(ctx *Context) string {
	facts, err := t.factProvider.GetFacts(ctx.Message.Author.ID)
	if err != nil {
		log.Printf("fogel trigger: get facts: %v", err)
		return ""
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
		log.Printf("fogel trigger: LLM error: %v", err)
		return ""
	}
	return resp
}
