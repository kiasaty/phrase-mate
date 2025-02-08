package app

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kiasaty/phrase-mate/models"
)

func (app *App) SaveUser(user *tgbotapi.User) (*models.User, error) {
	existingUser, err := app.DB.FindUserByTelegramID(user.ID)
	if err != nil {
		return nil, err
	}

	if existingUser == nil {
		newUser := &models.User{
			TelegramChatID: user.ID,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Username:       user.UserName,
			LanguageCode:   user.LanguageCode,
			IsBot:          user.IsBot,
		}

		savedUser, err := app.DB.CreateUser(newUser)
		if err != nil {
			return nil, err
		}

		return savedUser, nil
	}

	return existingUser, nil
}

func (app *App) FetchTelegramUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := app.TelegramBot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			app.handleCallbackQuery(update.CallbackQuery)
			continue
		}

		if update.Message != nil {
			app.handleNewPhrase(update.Message)
			continue
		}
	}
}

func (app *App) handleNewPhrase(message *tgbotapi.Message) {
	// Save the user information
	user, err := app.SaveUser(message.From)
	if err != nil {
		log.Printf("Error saving user: %v", err)
		return
	}

	// Extract message text and hashtags
	messageText := message.Text
	messageID := message.MessageID
	hashtags := extractHashtags(messageText)

	if len(hashtags) == 0 {
		log.Printf("No hashtags found in message: %s", messageText)
		return
	}

	// Check if the phrase already exists
	existingPhrase := app.DB.FindPhraseByMessageId(messageID)
	if existingPhrase != nil {
		log.Printf("Phrase with MessageID %d already exists", messageID)
		return
	}

	// Create or find tags
	var tags []models.Tag
	for _, hashtag := range hashtags {
		hashtag := strings.ToLower(hashtag)

		tag, err := app.DB.FindTagByName(hashtag)
		if err != nil {
			// Create tag if it doesn't exist
			tag, err = app.DB.CreateTag(&models.Tag{Name: hashtag})
			if err != nil {
				log.Printf("Error creating tag %s: %v", hashtag, err)
				return
			}
		}
		tags = append(tags, *tag)
	}

	// Create a new phrase with user reference and tags
	phrase := &models.Phrase{
		UserID:            user.ID,
		TelegramMessageID: messageID,
		Text:              messageText,
		Tags:              tags,
	}

	if _, err := app.DB.CreatePhrase(phrase); err != nil {
		log.Printf("Error creating phrase: %v", err)
		return
	}

	log.Printf("Phrase added for user %s: %s", user.Username, messageText)
}

func (app *App) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	data := strings.Split(callbackQuery.Data, ":")

	if len(data) != 3 || data[0] != "review" {
		log.Printf("Invalid callback data: %s", callbackQuery.Data)
		return
	}

	user, err := app.DB.FindUserByTelegramID(callbackQuery.From.ID)
	if err != nil {
		log.Printf("User not found: %v", err)
		return
	}

	phraseID, err := strconv.Atoi(data[1])
	if err != nil {
		log.Printf("Invalid phrase ID: %v", err)
		return
	}

	recallQualityNumber, err := strconv.ParseUint(data[2], 10, 8)
	if err != nil {
		fmt.Println("Invalid recall quality:", err)
		return
	}

	recallQuality := models.RecallQuality(recallQualityNumber)

	if !recallQuality.IsValid() {
		fmt.Println("Invalid RecallQuality value:", recallQuality)
		return
	}

	err = app.handleReview(user, uint(phraseID), recallQuality)
	if err == nil {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Thank you for your feedback!")
		if _, err := app.TelegramBot.Request(callback); err != nil {
			log.Printf("Failed to send callback response: %v", err)
		}
	}
}

func (app *App) handleReview(user *models.User, phraseID uint, recallQuality models.RecallQuality) error {
	review, err := app.ReviewPhrase(phraseID, user.ID, uint(1), recallQuality)

	if err != nil {
		log.Printf("Failed to review the phrase: %v", err)
		return err
	}

	if review == nil {
		return nil
	}

	err = app.DB.CreateReview(review)

	if err != nil {
		log.Printf("Failed to save phrase review: %v", err)
		return err
	}

	return nil
}
