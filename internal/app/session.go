package app

import (
	"time"

	"github.com/kiasaty/phrase-mate/internal/database"
	"github.com/kiasaty/phrase-mate/models"
)

func (app *App) GetOrStartSession(userID uint) (*models.Session, error) {
	activeSession, err := app.findActiveSession(userID)
	if err != nil {
		return nil, err
	}

	if activeSession == nil {
		return app.startSession(userID)
	}

	notReviewedPhrasesCount, err := app.DB.CountNotReviewedPhrasesBySessionId(activeSession.ID)
	if err != nil {
		return nil, err
	}

	if notReviewedPhrasesCount != 0 {
		return activeSession, nil
	}

	now := time.Now()
	activeSession.EndedAt = &now
	err = app.DB.UpdateSession(activeSession)
	if err != nil {
		return nil, err
	}

	activeSession, err = app.startSession(userID)
	if err != nil {
		return nil, err
	}

	return activeSession, nil
}

func (app *App) findActiveSession(userID uint) (*models.Session, error) {
	return app.DB.FindActiveSession(userID)
}

func (app *App) startSession(userID uint) (*models.Session, error) {
	var session *models.Session

	err := app.DB.Transaction(func(tx database.DatabaseClient) error {
		sessionSize := app.Config.SessionSize
		now := time.Now()

		var err error

		session, err = tx.CreateSession(&models.Session{
			UserID:    userID,
			StartedAt: now,
		})

		if err != nil {
			return err
		}

		dueSessionReviews, err := tx.GetDueReviews(userID, now, sessionSize)
		if err != nil {
			return err
		}

		remainingSlots := sessionSize - len(dueSessionReviews)
		if remainingSlots > 0 {
			newReviews, err := tx.GetNewReviews(userID, session.ID, remainingSlots)
			if err != nil {
				return err
			}

			dueSessionReviews = append(dueSessionReviews, newReviews...)
		}

		for i := range dueSessionReviews {
			dueSessionReviews[i].SessionID = session.ID
		}

		if err := tx.CreateReviews(dueSessionReviews); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return session, nil
}
