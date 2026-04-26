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
	r.On("HoodKing").Match(`\b(michael|jordan|king|prince|hood|ghetto)\b`).Probability(keywordTriggerProbability).Respond("Yeah but he's the hood king!")
	r.On("Odenigbo").Match(`\bod.{0,2}n.{0,2}bo\b`).Probability(keywordTriggerProbability).Respond("That's why I'm being odenigbo about it!")
	r.On("LoveHate").Match(`\bi (just )?(love|like|dislike|hate|want|need)\b`).Probability(keywordTriggerProbability).Respond("I know you do \U0001F609")
	r.On("Kanye").Match(`\b(kanye|west)\b`).Probability(keywordTriggerProbability).Respond("Rough year for Kanye")
	r.On("Food").Match(`\b(food|eat|eating|hungry|delicious|yummy|tasty|cook)\b`).Probability(keywordTriggerProbability).Respond("%s, thats moan-worthy")
	r.On("BDSM").Match(`bdsm`).Probability(keywordTriggerProbability).Respond("We did the BDSM test this summer and let's just say, I was very confused")
	r.On("Hockey").Match(`hockey`).Probability(keywordTriggerProbability).Respond("I've seen all of the hockey team player's dicks")
	r.On("Ben").Match(`\b(ben|benjamin)\b`).Probability(keywordTriggerProbability).Respond("Ben always gives me the satisfaction")
	r.On("Shame").Match(`\bsham(e|ing)\b`).Probability(keywordTriggerProbability).Respond("Public shaming is a good form of shaming")
	r.On("Grades").Match(`\b(grade|grades|smart|academic)\b`).Probability(keywordTriggerProbability).Respond("My grade went down\u2026 cuz it couldn't really go up! HEHEHEHE")
	r.On("IThink").Match(`\bi (think|believe)\b`).Probability(keywordTriggerProbability).RespondOneOf(ThinkResponses...)

	// 4. Language detection (code-driven: uses language detector)
	r.Add(&LanguageTrigger{})

	// 5. User-specific (DSL)
	r.On("ElectroshkRare").User(config.ElectroshkID, `\byifan\b`).Probability(electroshkRareProbability).Respond("Did you really do the YEEEbowl without me?!?")
	r.On("Electroshk").User(config.ElectroshkID, `\byifan\b`).Probability(electroshkProbability).Respond("YEEEEEEEEEEEEEEEEEEEEEE")
	r.On("Modriver").User(config.ModriverID, `\b(raul|rahul)\b`).Probability(modriverProbability).RespondOneOf("RAUL WHAT THE HELL?!", "Shut up, RAUL")

	// 6. Random negative (code-driven: probability-first optimization)
	r.Add(&RandomNegativeTrigger{})

	// 7. Generated response (code-driven: low-probability LLM-generated response)
	if llmClient != nil && factProvider != nil {
		r.Add(NewGeneratedResponseTrigger(llmClient, factProvider))
	}

	// 8. Random quotes (DSL)
	r.On("RandomQuote").Probability(randomQuoteProbability).RespondOneOf(RandomQuotes...)
}
