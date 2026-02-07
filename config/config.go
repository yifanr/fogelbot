package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var (
	DiscordToken string
	ElectroshkID string
	ModriverID   string
)

const (
	QuickReplyWindow     = 20 * time.Second
	QuickReplyCooldown   = 5 * time.Minute
	SpeakEnglishCooldown = 5 * time.Minute
	SentimentThreshold   = 0.05
)

func Load() {
	// Try loading from .env file, but don't fail if it doesn't exist (prod env might not have it)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	DiscordToken = os.Getenv("DISCORD_TOKEN")
	if DiscordToken == "" {
		log.Fatal("Error: DISCORD_TOKEN not found in environment")
	}

	ElectroshkID = os.Getenv("ELECTROSHK_ID")
	ModriverID = os.Getenv("MODRIVER_ID")
}
