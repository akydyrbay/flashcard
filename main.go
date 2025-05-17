package main

import (
	"flag"
	"flashcard/clients/telegram"
	"fmt"
	"log"
)

const (
	tgBotHost = "api.telegram.org"
)

func main() {
	tgClient := telegram.New(tgBotHost, mustToken())
	fmt.Println(tgClient)
}

func mustToken() string {
	token := flag.String("get-bot-token", "", "token for accessing tg bot")

	flag.Parse()

	if *token == "" {
		log.Fatal("token is not specified")
	}
	return *token
}
