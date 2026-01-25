package triggers

import (
	"github.com/bwmarrin/discordgo"
	"fogelbot/sentiment"
	"fogelbot/language"
	"fogelbot/state"
)

type Context struct {
	Session   *discordgo.Session
	Message   *discordgo.MessageCreate
	Cooldowns *state.CooldownManager
	Sentiment *sentiment.Analyzer
	Language  *language.Detector
}

type Trigger interface {
	Name() string
	Check(ctx *Context) (string, bool)
}

type Registry struct {
	triggers []Trigger
}

func NewRegistry() *Registry {
	return &Registry{
		triggers: make([]Trigger, 0),
	}
}

func (r *Registry) Register(t Trigger) {
	r.triggers = append(r.triggers, t)
}

func (r *Registry) Process(ctx *Context) string {
	for _, t := range r.triggers {
		if response, matched := t.Check(ctx); matched {
			return response
		}
	}
	return ""
}
