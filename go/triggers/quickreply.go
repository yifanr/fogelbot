package triggers

import (
	"math/rand"
	"time"
	"fogelbot/config"
)

type QuickReplyTrigger struct{}

func (t *QuickReplyTrigger) Name() string {
	return "QuickReply"
}

func (t *QuickReplyTrigger) Check(ctx *Context) (string, bool) {
	channelID := ctx.Message.ChannelID
	lastReply := ctx.Cooldowns.GetLastBotReply(channelID)
	
	// If bot hasn't replied ever, skip
	if lastReply.IsZero() {
		return "", false
	}

	now := time.Now()
	// Check window
	if now.Sub(lastReply) > config.QuickReplyWindow {
		return "", false
	}

	// Check cooldown
	lastTrigger := ctx.Cooldowns.GetLastQuickReplyTrigger(channelID)
	if !lastTrigger.IsZero() && now.Sub(lastTrigger) < config.QuickReplyCooldown {
		return "", false
	}

	// Update trigger time
	ctx.Cooldowns.SetLastQuickReplyTrigger(channelID)

	score := ctx.Sentiment.Compound(ctx.Message.Content)
	if score >= config.SentimentThreshold {
		return PositiveResponses[rand.Intn(len(PositiveResponses))], true
	} else if score <= -config.SentimentThreshold {
		return NegativeResponses[rand.Intn(len(NegativeResponses))], true
	} else {
		return PositiveResponses[rand.Intn(len(PositiveResponses))], true
	}
}
