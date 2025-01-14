package app

import (
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kiasaty/phrase-mate/internal/database"
)

type App struct {
	DB          database.DatabaseClient
	TelegramBot *tgbotapi.BotAPI
}

func NewApp(databaseClient database.DatabaseClient, telegramBot *tgbotapi.BotAPI) App {
	return App{
		DB:          databaseClient,
		TelegramBot: telegramBot,
	}
}

func (app *App) HandleCommand() {
	if len(os.Args) < 2 {
		fmt.Println("List of existing commands.")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "fetch-updates":
		app.FetchTelegramUpdates()
	case "migrate-database":
		app.DB.Migrate()
	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("List of existing commands.")
		os.Exit(1)
	}
}
