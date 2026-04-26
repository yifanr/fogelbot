package triggers

import (
	"fogelbot/config"
)

type RandomNegativeTrigger struct{}

func (t *RandomNegativeTrigger) Name() string {
	return "RandomNegative"
}

func (t *RandomNegativeTrigger) Check(ctx *Context) (string, bool) {
	if ctx.Rand.Float64() < randomNegativeProbability {
		score := ctx.Sentiment.Compound(ctx.Message.Content)
		if score <= -config.SentimentThreshold {
			return NegativeResponses[ctx.Rand.Intn(len(NegativeResponses))], true
		}
	}
	return "", false
}
