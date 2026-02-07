package triggers

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TriggerBuilder provides a fluent API for defining data-driven triggers.
type TriggerBuilder struct {
	registry *Registry
	name     string
	match    *regexp.Regexp
	userID   string
	userRe   *regexp.Regexp
	prob     float64
	sent     string // "positive", "negative", or ""
	cooldown time.Duration
}

// On starts building a new trigger with the given name.
func (r *Registry) On(name string) *TriggerBuilder {
	return &TriggerBuilder{
		registry: r,
		name:     name,
	}
}

func (b *TriggerBuilder) Match(pattern string) *TriggerBuilder {
	b.match = regexp.MustCompile(pattern)
	return b
}

func (b *TriggerBuilder) User(id, fallbackPattern string) *TriggerBuilder {
	b.userID = id
	if fallbackPattern != "" {
		b.userRe = regexp.MustCompile(fallbackPattern)
	}
	return b
}

func (b *TriggerBuilder) Probability(p float64) *TriggerBuilder {
	b.prob = p
	return b
}

func (b *TriggerBuilder) Sentiment(s string) *TriggerBuilder {
	b.sent = s
	return b
}

func (b *TriggerBuilder) Cooldown(d time.Duration) *TriggerBuilder {
	b.cooldown = d
	return b
}

func (b *TriggerBuilder) Respond(response string) {
	b.registry.Add(&DataDrivenTrigger{
		name:      b.name,
		match:     b.match,
		userID:    b.userID,
		userRe:    b.userRe,
		prob:      b.prob,
		sent:      b.sent,
		cooldown:  b.cooldown,
		responses: []string{response},
	})
}

func (b *TriggerBuilder) RespondOneOf(responses ...string) {
	b.registry.Add(&DataDrivenTrigger{
		name:      b.name,
		match:     b.match,
		userID:    b.userID,
		userRe:    b.userRe,
		prob:      b.prob,
		sent:      b.sent,
		cooldown:  b.cooldown,
		responses: responses,
	})
}

// DataDrivenTrigger is the Trigger implementation backing the builder DSL.
type DataDrivenTrigger struct {
	name      string
	match     *regexp.Regexp
	userID    string
	userRe    *regexp.Regexp
	prob      float64
	sent      string
	cooldown  time.Duration
	responses []string
}

func (t *DataDrivenTrigger) Name() string { return t.name }

func (t *DataDrivenTrigger) Check(ctx *Context) (string, bool) {
	contentLower := strings.ToLower(ctx.Message.Content)

	// 1. Match (regex on lowercased content)
	if t.match != nil {
		if !t.match.MatchString(contentLower) {
			return "", false
		}
	}

	// 2. User (author ID, mentions, or fallback regex)
	if t.userID != "" || t.userRe != nil {
		matched := false
		if t.userID != "" {
			if ctx.Message.Author.ID == t.userID {
				matched = true
			}
			if !matched {
				for _, m := range ctx.Message.Mentions {
					if m.ID == t.userID {
						matched = true
						break
					}
				}
			}
		}
		if !matched && t.userRe != nil {
			matched = t.userRe.MatchString(contentLower)
		}
		if !matched {
			return "", false
		}
	}

	// 3. Sentiment
	if t.sent != "" {
		score := ctx.Sentiment.Compound(ctx.Message.Content)
		switch t.sent {
		case "positive":
			if score < 0.05 {
				return "", false
			}
		case "negative":
			if score > -0.05 {
				return "", false
			}
		}
	}

	// 4. Probability
	if t.prob > 0 {
		if ctx.Rand.Float64() >= t.prob {
			return "", false
		}
	}

	// 5. Cooldown
	if t.cooldown > 0 {
		key := fmt.Sprintf("dsl:%s:%s", t.name, ctx.Message.ChannelID)
		if !ctx.Cooldowns.CheckAndSet(key, t.cooldown) {
			return "", false
		}
	}

	// 6. Pick response
	resp := t.responses[0]
	if len(t.responses) > 1 {
		resp = t.responses[ctx.Rand.Intn(len(t.responses))]
	}

	// Replace %s with author mention
	if strings.Contains(resp, "%s") {
		resp = fmt.Sprintf(resp, ctx.Message.Author.Mention())
	}

	return resp, true
}
