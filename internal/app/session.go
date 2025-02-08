package app

import (
	"time"

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
	session, err := app.DB.CreateSession(&models.Session{
		UserID:    userID,
		StartedAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (app *App) endSession(sessionID uint) error {
	return app.DB.EndSession(sessionID)
}
