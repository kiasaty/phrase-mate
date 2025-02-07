package app

import (
	"errors"
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
	lastReview, err := app.DB.FindPhraseLastReview(userID, phraseID)
	if err != nil {
		return nil, err
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
	newEaseFactor := previousEaseFactor + (0.1 - float64(5-recallQuality)*(0.08+float64(5-recallQuality)*0.02))
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

	review := &models.Review{
		PhraseID:      phraseID,
		UserID:        userID,
		SessionID:     sessionID,
		RecallQuality: recallQuality,
		EaseFactor:    newEaseFactor,
		Interval:      newInterval,
		ReviewedAt:    now,
		NextReviewAt:  now.AddDate(0, 0, int(newInterval)).Truncate(24 * time.Hour), // Calculate NextReviewAt (as a date)
	}

	return review, nil
}
