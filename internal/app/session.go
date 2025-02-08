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

	phrase, err := app.getNextPhraseToReview(activeSession)
	if err != nil {
		return nil, err
	}
	if phrase != nil {
		return activeSession, err
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
		now := time.Now()

		var err error

		session, err = tx.CreateSession(&models.Session{
			UserID:    userID,
			StartedAt: now,
		})

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (app *App) endSession(session *models.Session) (*models.Session, error) {
	if session.EndedAt != nil {
		return session, nil
	}

	now := time.Now()
	session.EndedAt = &now

	err := app.DB.UpdateSession(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}
