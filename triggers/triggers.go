package triggers

import (
	"math/rand"
	"time"

	"fogelbot/config"
	"fogelbot/state"
	"github.com/bwmarrin/discordgo"
)

// Injectable dependency interfaces

type SentimentAnalyzer interface {
	Compound(text string) float64
}

type LanguageDetector interface {
	IsNonEnglish(text string) bool
}

type Rand interface {
	Intn(n int) int
	Float64() float64
}

type Clock interface {
	Now() time.Time
}

// FactProvider retrieves stored facts about a user.
type FactProvider interface {
	GetFacts(userID string) ([]string, error)
}

// Production implementations

type stdRand struct{}

func (stdRand) Intn(n int) int   { return rand.Intn(n) }
func (stdRand) Float64() float64 { return rand.Float64() }

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// StdRand returns the production Rand implementation.
func StdRand() Rand { return stdRand{} }

// RealClock returns the production Clock implementation.
func RealClock() Clock { return realClock{} }

// Context carries all dependencies through the trigger pipeline.
type Context struct {
	Session   *discordgo.Session
	Message   *discordgo.MessageCreate
	Cooldowns *state.CooldownManager
	Sentiment SentimentAnalyzer
	Language  LanguageDetector
	Rand      Rand
	Clock     Clock
}

func (ctx *Context) DirectlyMentionsBot() bool {
	if ctx == nil || ctx.Session == nil || ctx.Session.State == nil || ctx.Session.State.User == nil || ctx.Message == nil {
		return false
	}
	for _, user := range ctx.Message.Mentions {
		if user.ID == ctx.Session.State.User.ID {
			return true
		}
	}
	return false
}

func (ctx *Context) RecentBotReplyWithin(d time.Duration) bool {
	if ctx == nil || ctx.Message == nil || ctx.Cooldowns == nil || ctx.Clock == nil {
		return false
	}
	lastReply := ctx.Cooldowns.GetLastBotReply(ctx.Message.ChannelID)
	if lastReply.IsZero() {
		return false
	}
	return ctx.Clock.Now().Sub(lastReply) < d
}

type Trigger interface {
	Name() string
	Check(ctx *Context) (string, bool)
}

type recentReplyExempt interface {
	AllowDuringRecentBotReply() bool
}

type Registry struct {
	triggers []Trigger
}

func NewRegistry() *Registry {
	return &Registry{
		triggers: make([]Trigger, 0),
	}
}

func (r *Registry) Add(t Trigger) {
	r.triggers = append(r.triggers, t)
}

// Register is an alias for Add for backward compatibility.
func (r *Registry) Register(t Trigger) {
	r.Add(t)
}

func (r *Registry) Process(ctx *Context) string {
	suppressAmbient := ctx.RecentBotReplyWithin(config.RecentReplyCooldown) && !ctx.DirectlyMentionsBot()
	for _, t := range r.triggers {
		if suppressAmbient {
			exempt, ok := t.(recentReplyExempt)
			if !ok || !exempt.AllowDuringRecentBotReply() {
				continue
			}
		}
		if response, matched := t.Check(ctx); matched {
			return response
		}
	}
	return ""
}
