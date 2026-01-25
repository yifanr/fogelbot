package triggers

import (
	"time"
	"fogelbot/config"
)

type LanguageTrigger struct{}

func (t *LanguageTrigger) Name() string {
	return "LanguageDetection"
}

func (t *LanguageTrigger) Check(ctx *Context) (string, bool) {
	if !ctx.Language.IsNonEnglish(ctx.Message.Content) {
		return "", false
	}

	channelID := ctx.Message.ChannelID
	now := time.Now()
	
	lastTrigger := ctx.Cooldowns.GetLastSpeakEnglishTrigger(channelID)
	if !lastTrigger.IsZero() && now.Sub(lastTrigger) < config.SpeakEnglishCooldown {
		return "", false
	}

	ctx.Cooldowns.SetLastSpeakEnglishTrigger(channelID)
	return "SPEAK ENGLISH!!", true
}
