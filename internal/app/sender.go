package app

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *App) SendPhrase(chatID int64, sessionID uint, phraseID uint, phraseText string) error {
	buttonKeyPrefix := "review:" + strconv.Itoa(int(sessionID)) + ":" + strconv.Itoa(int(phraseID))
	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("1", buttonKeyPrefix+":1"),
		tgbotapi.NewInlineKeyboardButtonData("2", buttonKeyPrefix+":2"),
		tgbotapi.NewInlineKeyboardButtonData("3", buttonKeyPrefix+":3"),
		tgbotapi.NewInlineKeyboardButtonData("4", buttonKeyPrefix+":4"),
		tgbotapi.NewInlineKeyboardButtonData("5", buttonKeyPrefix+":5"),
	}

	msg := tgbotapi.NewMessage(chatID, phraseText)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)

	_, err := app.TelegramBot.Send(msg)
	return err
}
