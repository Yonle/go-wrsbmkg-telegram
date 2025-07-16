package main

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
)

func main() {
	ReadConfig("wrsbmkg_telegrambot_config.yaml")

	if config.ChatID == 0 {
		panic("ChatID is not provided!")
	}

	if len(config.MsgMemoryDir) > 0 {
		go MemDirHouseKeeping()
	}

	ctx := context.Background()
	opts := []bot.Option{}

	b, err := bot.New(config.Token, opts...)
	if err != nil {
		panic(err)
	}

	log.Println("Ready. New reports will be ready in 15 seconds...")
	startBMKG(ctx, b)
}
