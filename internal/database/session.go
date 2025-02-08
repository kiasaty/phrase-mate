package database

import (
	"time"

	"github.com/kiasaty/phrase-mate/models"
	"gorm.io/gorm"
)

func (c *Client) CreateSession(session *models.Session) (*models.Session, error) {
	if err := c.DB.Create(session).Error; err != nil {
		return nil, err
	}
	return session, nil
}

func (c *Client) EndSession(sessionID uint) error {
	now := time.Now()
	return c.DB.Model(&models.Session{}).
		Where("id = ?", sessionID).
		Update("ended_at", now).Error
}

func (c *Client) FindActiveSession(userID uint) (*models.Session, error) {
	var session models.Session

	err := c.DB.Where("user_id = ? AND ended_at IS NULL", userID).First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &session, nil
}
