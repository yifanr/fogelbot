package triggers

import (
	"math/rand"
	"fogelbot/config"
)

type RandomNegativeTrigger struct{}

func (t *RandomNegativeTrigger) Name() string {
	return "RandomNegative"
}

func (t *RandomNegativeTrigger) Check(ctx *Context) (string, bool) {
	if rand.Float64() < 0.05 {
		score := ctx.Sentiment.Compound(ctx.Message.Content)
		if score <= -config.SentimentThreshold {
			return NegativeResponses[rand.Intn(len(NegativeResponses))], true
		}
	}
	return "", false
}

type RandomQuoteTrigger struct{}

func (t *RandomQuoteTrigger) Name() string {
	return "RandomQuote"
}

func (t *RandomQuoteTrigger) Check(ctx *Context) (string, bool) {
	if rand.Float64() < 0.02 {
		return RandomQuotes[rand.Intn(len(RandomQuotes))], true
	}
	return "", false
}
