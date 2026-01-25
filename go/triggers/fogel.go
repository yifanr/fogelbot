package triggers

import (
	"math/rand"
	"strings"
	"fogelbot/config"
)

type FogelTrigger struct{}

func (t *FogelTrigger) Name() string {
	return "FogelMention"
}

func (t *FogelTrigger) Check(ctx *Context) (string, bool) {
	contentLower := strings.ToLower(ctx.Message.Content)
	mentioned := false
	if strings.Contains(contentLower, "fogel") {
		mentioned = true
	}
	
	// Check mentions
	for _, user := range ctx.Message.Mentions {
		if user.ID == ctx.Session.State.User.ID {
			mentioned = true
			break
		}
	}

	if !mentioned {
		return "", false
	}

	score := ctx.Sentiment.Compound(ctx.Message.Content)
	if score >= config.SentimentThreshold {
		return PositiveResponses[rand.Intn(len(PositiveResponses))], true
	} else if score <= -config.SentimentThreshold {
		return NegativeResponses[rand.Intn(len(NegativeResponses))], true
	} else {
		// Neutral - pick randomly from positive
		return PositiveResponses[rand.Intn(len(PositiveResponses))], true
	}
}
