package triggers

import (
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
		return PositiveResponses[ctx.Rand.Intn(len(PositiveResponses))], true
	} else if score <= -config.SentimentThreshold {
		return NegativeResponses[ctx.Rand.Intn(len(NegativeResponses))], true
	} else {
		return PositiveResponses[ctx.Rand.Intn(len(PositiveResponses))], true
	}
}
