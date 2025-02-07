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
	Config      Config
}

type Config struct {
	SessionSize int
}

func NewApp(databaseClient database.DatabaseClient, telegramBot *tgbotapi.BotAPI, config Config) App {
	return App{
		DB:          databaseClient,
		TelegramBot: telegramBot,
		Config:      config,
	}
}

func GetDefaultConfig() Config {
	return Config{
		SessionSize: 20,
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
	case "send-next-phrases-to-review":
		app.SendNextPhraseToReviewForAllUsers()
	case "migrate-database":
		app.DB.Migrate()
	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("List of existing commands.")
		os.Exit(1)
	}
}
