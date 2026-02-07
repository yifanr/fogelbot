package triggers

import (
	"fogelbot/config"
	"fogelbot/llm"
)

// RegisterAll registers all triggers in priority order (first-match-wins).
// llmClient and factProvider may be nil if LLM is not configured.
func RegisterAll(r *Registry, sentiment SentimentAnalyzer, lang LanguageDetector, llmClient llm.LLM, factProvider FactProvider) {
	// 1. Fogel mention (code-driven: complex sentiment + @mention logic)
	r.Add(NewFogelTrigger(llmClient, factProvider))

	// 2. Quick reply (code-driven: stateful bot-reply timestamps)
	r.Add(&QuickReplyTrigger{})

	// 3. Keyword triggers (DSL)
	r.On("HoodKing").Match(`\b(michael|jordan|king|prince|hood|ghetto)\b`).Respond("Yeah but he's the hood king!")
	r.On("Odenigbo").Match(`\bod.{0,2}n.{0,2}bo\b`).Respond("That's why I'm being odenigbo about it!")
	r.On("LoveHate").Match(`\bi (just )?(love|like|dislike|hate|want|need)\b`).Respond("I know you do \U0001F609")
	r.On("Kanye").Match(`\b(kanye|west)\b`).Respond("Rough year for Kanye")
	r.On("Food").Match(`\b(food|eat|eating|hungry|delicious|yummy|tasty|cook)\b`).Respond("%s, thats moan-worthy")
	r.On("BDSM").Match(`bdsm`).Respond("We did the BDSM test this summer and let's just say, I was very confused")
	r.On("Hockey").Match(`hockey`).Respond("I've seen all of the hockey team player's dicks")
	r.On("Ben").Match(`\b(ben|benjamin)\b`).Respond("Ben always gives me the satisfaction")
	r.On("Shame").Match(`\bsham(e|ing)\b`).Respond("Public shaming is a good form of shaming")
	r.On("Grades").Match(`\b(grade|grades|smart|academic)\b`).Respond("My grade went down\u2026 cuz it couldn't really go up! HEHEHEHE")
	r.On("IThink").Match(`\bi (think|believe)\b`).RespondOneOf(ThinkResponses...)

	// 4. Language detection (code-driven: uses language detector)
	r.Add(&LanguageTrigger{})

	// 5. User-specific (DSL)
	r.On("ElectroshkRare").User(config.ElectroshkID, `\byifan\b`).Probability(0.01).Respond("Did you really do the YEEEbowl without me?!?")
	r.On("Electroshk").User(config.ElectroshkID, `\byifan\b`).Probability(0.05).Respond("YEEEEEEEEEEEEEEEEEEEEEE")
	r.On("Modriver").User(config.ModriverID, `\b(raul|rahul)\b`).Probability(0.05).RespondOneOf("RAUL WHAT THE HELL?!", "Shut up, RAUL")

	// 6. Random negative (code-driven: probability-first optimization)
	r.Add(&RandomNegativeTrigger{})

	// 7. Generated response (code-driven: 5% LLM-generated response)
	if llmClient != nil && factProvider != nil {
		r.Add(NewGeneratedResponseTrigger(llmClient, factProvider))
	}

	// 8. Random quotes (DSL)
	r.On("RandomQuote").Probability(0.02).RespondOneOf(RandomQuotes...)
}
