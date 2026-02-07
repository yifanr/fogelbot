package bot

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"fogelbot/config"
	"fogelbot/facts"
	"fogelbot/language"
	"fogelbot/llm"
	"fogelbot/sentiment"
	"fogelbot/state"
	"fogelbot/triggers"
)

type Bot struct {
	Session         *discordgo.Session
	Cooldowns       *state.CooldownManager
	Sentiment       *sentiment.Analyzer
	Language        *language.Detector
	TriggerRegistry *triggers.Registry
	FactCollector   *facts.Collector
}

func New() *Bot {
	dg, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		log.Fatal("Error creating Discord session, ", err)
	}

	clock := triggers.RealClock()

	bot := &Bot{
		Session:         dg,
		Cooldowns:       state.NewCooldownManager(clock),
		Sentiment:       sentiment.NewAnalyzer(),
		Language:        language.NewDetector(),
		TriggerRegistry: triggers.NewRegistry(),
	}

	var llmClient llm.LLM
	var factProvider triggers.FactProvider

	if config.GeminiAPIKey != "" {
		gemini := llm.NewGemini(config.GeminiAPIKey)
		store, err := facts.NewFactStore(config.FactDBPath)
		if err != nil {
			log.Fatal("Error opening fact store: ", err)
		}
		collector := facts.NewCollector(store, gemini, clock)
		bot.FactCollector = collector
		llmClient = gemini
		factProvider = collector
	}

	triggers.RegisterAll(bot.TriggerRegistry, bot.Sentiment, bot.Language, llmClient, factProvider)

	dg.AddHandler(bot.OnReady)
	dg.AddHandler(bot.OnMessage)

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	return bot
}

func (b *Bot) Run() {
	if b.FactCollector != nil {
		b.FactCollector.Start()
		defer b.FactCollector.Stop()
	}

	err := b.Session.Open()
	if err != nil {
		log.Fatal("Error opening connection, ", err)
	}
	defer b.Session.Close()

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
