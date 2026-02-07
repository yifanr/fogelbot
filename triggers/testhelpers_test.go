package triggers

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"fogelbot/state"
)

// --- Fakes ---

type fakeSentiment struct {
	score float64
}

func (f *fakeSentiment) Compound(text string) float64 { return f.score }

type fakeLanguage struct {
	nonEnglish bool
}

func (f *fakeLanguage) IsNonEnglish(text string) bool { return f.nonEnglish }

type fixedRand struct {
	intIdx     int
	floatIdx   int
	intValues  []int
	floatValues []float64
}

func (f *fixedRand) Intn(n int) int {
	if len(f.intValues) == 0 {
		return 0
	}
	v := f.intValues[f.intIdx%len(f.intValues)]
	f.intIdx++
	return v % n
}

func (f *fixedRand) Float64() float64 {
	if len(f.floatValues) == 0 {
		return 0
	}
	v := f.floatValues[f.floatIdx%len(f.floatValues)]
	f.floatIdx++
	return v
}

func newFixedRand(ints []int, floats []float64) *fixedRand {
	return &fixedRand{intValues: ints, floatValues: floats}
}

// newAlwaysRand returns a Rand that always returns the given int and float values.
func newAlwaysRand(i int, f float64) *fixedRand {
	return &fixedRand{intValues: []int{i}, floatValues: []float64{f}}
}

type fakeClock struct {
	now time.Time
}

func (f *fakeClock) Now() time.Time { return f.now }

func (f *fakeClock) Advance(d time.Duration) {
	f.now = f.now.Add(d)
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)}
}

// --- Context builder ---

type testContextBuilder struct {
	content    string
	authorID   string
	authorName string
	channelID  string
	mentions   []*discordgo.User
	botID      string
	sentiment  float64
	nonEnglish bool
	rand       Rand
	clock      Clock
	cooldowns  *state.CooldownManager
}

func newTestContext(content string) *testContextBuilder {
	return &testContextBuilder{
		content:    content,
		authorID:   "user_id",
		authorName: "TestUser",
		channelID:  "channel_1",
		botID:      "bot_id",
		mentions:   []*discordgo.User{},
	}
}

func (b *testContextBuilder) WithAuthor(id, name string) *testContextBuilder {
	b.authorID = id
	b.authorName = name
	return b
}

func (b *testContextBuilder) WithChannel(id string) *testContextBuilder {
	b.channelID = id
	return b
}

func (b *testContextBuilder) WithMentions(users ...*discordgo.User) *testContextBuilder {
	b.mentions = users
	return b
}

func (b *testContextBuilder) WithBotID(id string) *testContextBuilder {
	b.botID = id
	return b
}

func (b *testContextBuilder) WithSentiment(score float64) *testContextBuilder {
	b.sentiment = score
	return b
}

func (b *testContextBuilder) WithNonEnglish(v bool) *testContextBuilder {
	b.nonEnglish = v
	return b
}

func (b *testContextBuilder) WithRand(r Rand) *testContextBuilder {
	b.rand = r
	return b
}

func (b *testContextBuilder) WithClock(c Clock) *testContextBuilder {
	b.clock = c
	return b
}

func (b *testContextBuilder) WithCooldowns(cm *state.CooldownManager) *testContextBuilder {
	b.cooldowns = cm
	return b
}

func (b *testContextBuilder) Build() *Context {
	clk := b.clock
	if clk == nil {
		clk = newFakeClock()
	}
	r := b.rand
	if r == nil {
		r = newAlwaysRand(0, 0.0)
	}
	cm := b.cooldowns
	if cm == nil {
		cm = state.NewCooldownManager(clk)
	}
	return &Context{
		Session: &discordgo.Session{
			State: &discordgo.State{
				Ready: discordgo.Ready{
					User: &discordgo.User{ID: b.botID},
				},
			},
		},
		Message: &discordgo.MessageCreate{
			Message: &discordgo.Message{
				Content:   b.content,
				Author:    &discordgo.User{ID: b.authorID, Username: b.authorName},
				Mentions:  b.mentions,
				ChannelID: b.channelID,
			},
		},
		Cooldowns: cm,
		Sentiment: &fakeSentiment{score: b.sentiment},
		Language:  &fakeLanguage{nonEnglish: b.nonEnglish},
		Rand:      r,
		Clock:     clk,
	}
}

// --- Assertion helpers ---

func assertMatched(t *testing.T, trigger Trigger, ctx *Context) string {
	t.Helper()
	resp, matched := trigger.Check(ctx)
	if !matched {
		t.Errorf("%s: expected match, got no match", trigger.Name())
	}
	if resp == "" {
		t.Errorf("%s: expected non-empty response", trigger.Name())
	}
	return resp
}

func assertNotMatched(t *testing.T, trigger Trigger, ctx *Context) {
	t.Helper()
	resp, matched := trigger.Check(ctx)
	if matched {
		t.Errorf("%s: expected no match, got response %q", trigger.Name(), resp)
	}
}

func assertResponseOneOf(t *testing.T, got string, allowed []string) {
	t.Helper()
	for _, a := range allowed {
		if got == a {
			return
		}
	}
	t.Errorf("response %q not in allowed set %v", got, allowed)
}

func assertContains(t *testing.T, got, substr string) {
	t.Helper()
	if len(got) == 0 || len(substr) == 0 {
		t.Errorf("assertContains: got=%q, substr=%q", got, substr)
		return
	}
	if !contains(got, substr) {
		t.Errorf("expected %q to contain %q", got, substr)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
