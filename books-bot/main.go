package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	setupCommands(bot)

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	bookSearcherList := []BookSearcher{
		//NewZLibrarySearcher(),
		NewFlibustaSearcher(),
	}

	globalBookSearcher := NewGlobalSearchBookService(bookSearcherList)
	handler := NewHandler(bot, globalBookSearcher)

	for update := range updates {
		err := handler.HandleUpdate(&update)
		if err != nil {
			log.Println(err)
		}
	}
}

func setupCommands(bot *tgbotapi.BotAPI) {
	config := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     "/state",
			Description: "Get projects state",
		},
		tgbotapi.BotCommand{
			Command:     "/status",
			Description: "Check bot is alive",
		},
		tgbotapi.BotCommand{
			Command:     "/help",
			Description: "Get help",
		},
	)

	_, _ = bot.Request(config)
}
