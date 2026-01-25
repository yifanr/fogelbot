package triggers

import (
	"math/rand"
	"strings"
	"regexp"
	"fogelbot/config"
)

type ElectroshkTrigger struct{}

func (t *ElectroshkTrigger) Name() string {
	return "Electroshk"
}

func (t *ElectroshkTrigger) Check(ctx *Context) (string, bool) {
	contentLower := strings.ToLower(ctx.Message.Content)
	triggered := false
	
	// Check ID if configured
	if config.ElectroshkID != "" {
		if ctx.Message.Author.ID == config.ElectroshkID {
			triggered = true
		} else {
			for _, m := range ctx.Message.Mentions {
				if m.ID == config.ElectroshkID {
					triggered = true
					break
				}
			}
		}
	}
	
	// Check "yifan" regex
	if !triggered {
		match, _ := regexp.MatchString(`\byifan\b`, contentLower)
		if match {
			triggered = true
		}
	}

	if !triggered {
		return "", false
	}

	roll := rand.Float64()
	if roll < 0.01 {
		return "Did you really do the YEEEbowl without me?!?", true
	} else if roll < 0.06 { // 0.01 + 0.05
		return "YEEEEEEEEEEEEEEEEEEEEEE", true
	}

	return "", false
}

type ModriverTrigger struct{}

func (t *ModriverTrigger) Name() string {
	return "Modriver"
}

func (t *ModriverTrigger) Check(ctx *Context) (string, bool) {
	contentLower := strings.ToLower(ctx.Message.Content)
	triggered := false

	if config.ModriverID != "" {
		if ctx.Message.Author.ID == config.ModriverID {
			triggered = true
		} else {
			for _, m := range ctx.Message.Mentions {
				if m.ID == config.ModriverID {
					triggered = true
					break
				}
			}
		}
	}

	if !triggered {
		match, _ := regexp.MatchString(`\b(raul|rahul)\b`, contentLower)
		if match {
			triggered = true
		}
	}

	if !triggered {
		return "", false
	}

	if rand.Float64() < 0.05 {
		resps := []string{"RAUL WHAT THE HELL?!", "Shut up, RAUL"}
		return resps[rand.Intn(len(resps))], true
	}
	
	return "", false
}
