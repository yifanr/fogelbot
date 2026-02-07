package triggers

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

type KeywordTrigger struct {
	regex    *regexp.Regexp
	response func(*Context) string
	name     string
}

func (t *KeywordTrigger) Name() string {
	return t.name
}

func (t *KeywordTrigger) Check(ctx *Context) (string, bool) {
	contentLower := strings.ToLower(ctx.Message.Content)
	if t.regex.MatchString(contentLower) {
		return t.response(ctx), true
	}
	return "", false
}

func NewKeywordTriggers() []Trigger {
	triggers := []Trigger{}

	// Helper to add simple response trigger
	add := func(name, pattern string, resp string) {
		re := regexp.MustCompile(pattern)
		triggers = append(triggers, &KeywordTrigger{
			name:  name,
			regex: re,
			response: func(ctx *Context) string {
				return resp
			},
		})
	}
    
    // Helper for dynamic response
    addFunc := func(name, pattern string, f func(*Context) string) {
        re := regexp.MustCompile(pattern)
        triggers = append(triggers, &KeywordTrigger{
            name: name,
            regex: re,
            response: f,
        })
    }

	// 1. michael/jordan/king/prince/hood/ghetto
	add("HoodKing", `\b(michael|jordan|king|prince|hood|ghetto)\b`, "Yeah but he's the hood king!")

	// 2. Odenigbo
	add("Odenigbo", `\bod.{0,2}n.{0,2}bo\b`, "That's why I'm being odenigbo about it!")

	// 3. Love/like/dislike/hate/want/need
	add("LoveHate", `\bi (just )?(love|like|dislike|hate|want|need)\b`, "I know you do 😉")

	// 4. Kanye/West
	add("Kanye", `\b(kanye|west)\b`, "Rough year for Kanye")

	// 5. Food/eat/eating...
	addFunc("Food", `\b(food|eat|eating|hungry|delicious|yummy|tasty|cook)\b`, func(ctx *Context) string {
		return fmt.Sprintf("%s, thats moan-worthy", ctx.Message.Author.Mention())
	})

	// 6. BDSM (simple string check in python)
	add("BDSM", `bdsm`, "We did the BDSM test this summer and let's just say, I was very confused")

	// 7. Hockey
	add("Hockey", `hockey`, "I've seen all of the hockey team player's dicks")

	// 8. Ben/Benjamin
	add("Ben", `\b(ben|benjamin)\b`, "Ben always gives me the satisfaction")

	// 9. Shame/shaming
	add("Shame", `\bsham(e|ing)\b`, "Public shaming is a good form of shaming")

	// 10. Grade/academic
	add("Grades", `\b(grade|grades|smart|academic)\b`, "My grade went down… cuz it couldn't really go up! HEHEHEHE")

	// 11. I think/believe
	addFunc("IThink", `\bi (think|believe)\b`, func(ctx *Context) string {
		resp := ThinkResponses[rand.Intn(len(ThinkResponses))]
		return fmt.Sprintf(resp, ctx.Message.Author.Mention())
	})

	return triggers
}
