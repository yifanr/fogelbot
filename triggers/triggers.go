package triggers

import (
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"fogelbot/state"
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

// Production implementations

type stdRand struct{}

func (stdRand) Intn(n int) int      { return rand.Intn(n) }
func (stdRand) Float64() float64    { return rand.Float64() }

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

func (r *Registry) Add(t Trigger) {
	r.triggers = append(r.triggers, t)
}

// Register is an alias for Add for backward compatibility.
func (r *Registry) Register(t Trigger) {
	r.Add(t)
}

func (r *Registry) Process(ctx *Context) string {
	for _, t := range r.triggers {
		if response, matched := t.Check(ctx); matched {
			return response
		}
	}
	return ""
}
