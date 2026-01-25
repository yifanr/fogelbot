package bot

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"fogelbot/config"
	"fogelbot/language"
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
}

func New() *Bot {
	dg, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		log.Fatal("Error creating Discord session, ", err)
	}

	bot := &Bot{
		Session:   dg,
		Cooldowns: state.NewCooldownManager(),
		Sentiment: sentiment.NewAnalyzer(),
		Language:  language.NewDetector(),
		TriggerRegistry: triggers.NewRegistry(),
	}

	// Register triggers in priority order
	// 1. Fogel mention
	bot.TriggerRegistry.Register(&triggers.FogelTrigger{})
	// 2. Quick reply
	bot.TriggerRegistry.Register(&triggers.QuickReplyTrigger{})
	// 3. Keyword triggers
	for _, t := range triggers.NewKeywordTriggers() {
		bot.TriggerRegistry.Register(t)
	}
	// 4. Language detection
	bot.TriggerRegistry.Register(&triggers.LanguageTrigger{})
	// 5. User-specific
	bot.TriggerRegistry.Register(&triggers.ElectroshkTrigger{})
	bot.TriggerRegistry.Register(&triggers.ModriverTrigger{})
	// 6. Random negative
	bot.TriggerRegistry.Register(&triggers.RandomNegativeTrigger{})
	// 7. Random quotes
	bot.TriggerRegistry.Register(&triggers.RandomQuoteTrigger{})

	// Add handlers
	dg.AddHandler(bot.OnReady)
	dg.AddHandler(bot.OnMessage)

	// Identify intents
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	return bot
}

func (b *Bot) Run() {
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
