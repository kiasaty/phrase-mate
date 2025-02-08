package app

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *App) SendPhrase(chatID int64, phraseID uint, phraseText string) error {
	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("1", "review:"+strconv.Itoa(int(phraseID))+":1"),
		tgbotapi.NewInlineKeyboardButtonData("2", "review:"+strconv.Itoa(int(phraseID))+":2"),
		tgbotapi.NewInlineKeyboardButtonData("3", "review:"+strconv.Itoa(int(phraseID))+":3"),
		tgbotapi.NewInlineKeyboardButtonData("4", "review:"+strconv.Itoa(int(phraseID))+":4"),
		tgbotapi.NewInlineKeyboardButtonData("5", "review:"+strconv.Itoa(int(phraseID))+":5"),
	}

	msg := tgbotapi.NewMessage(chatID, phraseText)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)

	_, err := app.TelegramBot.Send(msg)
	return err
}
