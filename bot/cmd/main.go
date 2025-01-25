package main

import (
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	setupBot "github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/setup"

	"log"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/config"
	_ "time/tzdata"
)

func main() {
	cfg := config.Get()
	b, err := bot.New(cfg)
	if err != nil {
		log.Panic(err)
	}

	setupBot.Setup(b)

	b.Start()
}
