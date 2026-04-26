package triggers

import (
	"strings"
	"unicode"

	"fogelbot/config"
)

const quickReplyMinLetters = 8

type QuickReplyTrigger struct{}

func (t *QuickReplyTrigger) Name() string {
	return "QuickReply"
}

func (t *QuickReplyTrigger) AllowDuringRecentBotReply() bool {
	return true
}

func (t *QuickReplyTrigger) Check(ctx *Context) (string, bool) {
	if meaningfulLetterCount(ctx.Message.Content) < quickReplyMinLetters {
		return "", false
	}

	channelID := ctx.Message.ChannelID
	lastReply := ctx.Cooldowns.GetLastBotReply(channelID)

	if lastReply.IsZero() {
		return "", false
	}

	now := ctx.Clock.Now()
	if now.Sub(lastReply) < config.QuickReplyMinDelay {
		return "", false
	}
	if now.Sub(lastReply) > config.QuickReplyWindow {
		return "", false
	}

	lastTrigger := ctx.Cooldowns.GetLastQuickReplyTrigger(channelID)
	if !lastTrigger.IsZero() && now.Sub(lastTrigger) < config.QuickReplyCooldown {
		return "", false
	}

	if ctx.Rand.Float64() >= quickReplyProbability {
		return "", false
	}

	ctx.Cooldowns.SetLastQuickReplyTrigger(channelID)

	score := ctx.Sentiment.Compound(ctx.Message.Content)
	if score >= config.SentimentThreshold {
		return PositiveResponses[ctx.Rand.Intn(len(PositiveResponses))], true
	} else if score <= -config.SentimentThreshold {
		return NegativeResponses[ctx.Rand.Intn(len(NegativeResponses))], true
	} else {
		return PositiveResponses[ctx.Rand.Intn(len(PositiveResponses))], true
	}
}

func meaningfulLetterCount(text string) int {
	count := 0
	for _, r := range strings.TrimSpace(text) {
		if unicode.IsLetter(r) {
			count++
		}
	}
	return count
}
