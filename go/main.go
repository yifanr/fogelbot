package main

import (
	"fogelbot/bot"
	"fogelbot/config"
)

func main() {
	config.Load()
	
	fogelBot := bot.New()
	fogelBot.Run()
}