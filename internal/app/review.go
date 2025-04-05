package app

import (
	"errors"
	"log"
	"math"
	"time"

	"github.com/kiasaty/phrase-mate/models"
)

func (app *App) ReviewPhrase(
	phraseID, userID, sessionID uint,
	recallQuality models.RecallQuality,
) (*models.Review, error) {
	if !recallQuality.IsValid() {
		return nil, errors.New("invalid recall quality")
	}

	// Fetch the last review for the given PhraseID and UserID
	lastReview, err := app.DB.FindReview(userID, phraseID)
	if err != nil {
		return nil, err
	}

	// Check if review is not due yet
	if lastReview != nil && lastReview.NextReviewAt != nil && lastReview.NextReviewAt.After(time.Now()) {
		return nil, nil
	}

	// Default values for new phrases
	previousEaseFactor := 2.5
	previousInterval := uint16(0)

	// Use the last review's values if available
	if lastReview != nil {
		previousEaseFactor = lastReview.EaseFactor
		previousInterval = lastReview.Interval
	}

	// Adjust EaseFactor based on RecallQuality
	qualityDiff := float64(5 - recallQuality)
	newEaseFactor := previousEaseFactor + (0.1 - qualityDiff*(0.08+qualityDiff*0.02))
	newEaseFactor = math.Round(newEaseFactor*100) / 100 // Round to 2 decimal places
	if newEaseFactor < 1.3 {
		newEaseFactor = 1.3
	}

	// Adjust Interval based on RecallQuality
	var newInterval uint16
	switch {
	case recallQuality < models.QualityRemembered:
		// Reset interval for low recall quality
		newInterval = 1
	case previousInterval == 0:
		// Set to 1 for the first review
		newInterval = 1
	default:
		// Increase interval based on ease factor
		newInterval = uint16(float64(previousInterval) * newEaseFactor)
	}

	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	nextReviewAt := startOfToday.AddDate(0, 0, int(newInterval))

	review := &models.Review{
		PhraseID:      phraseID,
		UserID:        userID,
		SessionID:     sessionID,
		RecallQuality: recallQuality,
		EaseFactor:    newEaseFactor,
		Interval:      newInterval,
		ReviewedAt:    &now,
		NextReviewAt:  &nextReviewAt,
	}

	if err := app.storeReview(review); err != nil {
		return nil, err
	}

	// Check if the phrase should be retired
	if newInterval >= uint16(app.Config.MaxIntervalDays) {
		if err := app.markPhraseAsMastered(phraseID); err != nil {
			return nil, err
		}
	}

	return review, nil
}

func (app *App) SendNextPhraseToReviewForAllUsers() {
	users, err := app.DB.GetAllUsers()
	if err != nil {
		log.Printf("Error retrieving users: %v", err)
		return
	}

	for _, user := range users {
		app.SendNextPhraseToReviewForUser(user)
	}
}

func (app *App) SendNextPhraseToReviewForUser(user *models.User) {
	session, err := app.GetOrStartSession(user.ID)
	if err != nil {
		log.Printf("Fetching the active session failed: %v", err)
		return
	}
	if session == nil {
		log.Printf("No active session was found for user: %d", user.ID)
		return
	}

	phrase, err := app.getNextPhraseToReview(session)
	if err != nil {
		log.Printf("Finding the next phrase to review failed: %v", err)
		return
	}
	if phrase == nil {
		log.Printf("No phrase was found to review for session: %d", session.ID)
		return
	}

	err = app.SendPhrase(
		user.TelegramChatID,
		session.ID,
		phrase.ID,
		removeHashtags(phrase.Text),
	)
	if err != nil {
		log.Printf("Sending the phrase failed: %v", err)
		return
	}
}

func (app *App) getNextPhraseToReview(session *models.Session) (*models.Phrase, error) {
	sessionSize := app.Config.SessionSize
	now := time.Now()

	dueReview, err := app.DB.GetDueReview(session.UserID, now, sessionSize)
	if err != nil {
		return nil, err
	}

	if dueReview != nil {
		phrase, err := app.DB.FindPhrase(dueReview.PhraseID)
		if err != nil {
			return nil, err
		}

		// Skip mastered phrases
		if phrase.IsMastered {
			return app.getNextPhraseToReview(session)
		}

		return phrase, nil
	}

	newPhraseIDs, err := app.DB.FindNewPhrasesToReview(session.UserID, 1)
	if err != nil {
		return nil, err
	}
	if len(newPhraseIDs) == 0 {
		return nil, nil
	}

	phrase, err := app.DB.FindPhrase(newPhraseIDs[0])
	if err != nil {
		return nil, err
	}

	// Skip mastered phrases
	if phrase.IsMastered {
		return app.getNextPhraseToReview(session)
	}

	return phrase, nil
}

func (app *App) storeReview(review *models.Review) error {
	return app.DB.CreateReview(review)
}

func (app *App) GetReviewHistory(userID uint, phraseID uint) ([]*models.ReviewHistory, error) {
	return app.DB.FindReviewHistory(userID, phraseID)
}

func (app *App) markPhraseAsMastered(phraseID uint) error {
	return app.DB.MarkPhraseAsMastered(phraseID)
}
