package app

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *App) SendPhrase(chatID int64, phraseID uint, phraseText string) error {
	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Easy", "review:"+strconv.Itoa(int(phraseID))+":easy"),
		tgbotapi.NewInlineKeyboardButtonData("Medium", "review:"+strconv.Itoa(int(phraseID))+":medium"),
		tgbotapi.NewInlineKeyboardButtonData("Hard", "review:"+strconv.Itoa(int(phraseID))+":hard"),
	}

	msg := tgbotapi.NewMessage(chatID, phraseText)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)

	_, err := app.TelegramBot.Send(msg)
	return err
}
