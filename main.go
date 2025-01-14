package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/kiasaty/phrase-mate/internal/app"
	"github.com/kiasaty/phrase-mate/internal/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the .env file: %v", err)
	}

	databaseDSN := os.Getenv("DATABASE_DSN")
	if databaseDSN == "" {
		log.Fatalf("DATABASE_DSN not set in .env file!")
	}

	databaseClient, err := database.NewDatabaseClient(databaseDSN)
	if err != nil {
		log.Fatalf("Failed to create the database client: %v", err)
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatalf("TELEGRAM_BOT_TOKEN not set in .env file!")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	app := app.NewApp(databaseClient, bot)

	app.HandleCommand()
}
