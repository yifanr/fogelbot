package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"fogelbot/triggers"
)

func (b *Bot) OnReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

func (b *Bot) OnMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	ctx := &triggers.Context{
		Session:   s,
		Message:   m,
		Cooldowns: b.Cooldowns,
		Sentiment: b.Sentiment,
		Language:  b.Language,
		Rand:      triggers.StdRand(),
		Clock:     triggers.RealClock(),
	}

	response := b.TriggerRegistry.Process(ctx)
	if response != "" {
		_, err := s.ChannelMessageSend(m.ChannelID, response)
		if err != nil {
			log.Println("Error sending message:", err)
		} else {
			b.Cooldowns.SetLastBotReply(m.ChannelID)
		}
	}
}
