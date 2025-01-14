package app

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kiasaty/phrase-mate/models"
)

func (app *App) SaveUser(user *tgbotapi.User) (*models.User, error) {
	existingUser, err := app.DB.FindUserByTelegramID(user.ID)
	if err != nil && err.Error() != "record not found" {
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
	// Initialize Telegram updates configuration
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Start receiving updates
	updates := app.TelegramBot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Save the user information
		user, err := app.SaveUser(update.Message.From)
		if err != nil {
			log.Printf("Error saving user: %v", err)
			continue
		}

		// Extract message text and hashtags
		messageText := update.Message.Text
		messageID := update.Message.MessageID
		hashtags := extractHashtags(messageText)

		if len(hashtags) == 0 {
			log.Printf("No hashtags found in message: %s", messageText)
			continue
		}

		// Check if the phrase already exists
		existingPhrase := app.DB.FindPhraseByMessageId(messageID)
		if existingPhrase != nil {
			log.Printf("Phrase with MessageID %d already exists", messageID)
			continue
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
					continue
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
			continue
		}

		log.Printf("Phrase added for user %s: %s", user.Username, messageText)
	}
}
