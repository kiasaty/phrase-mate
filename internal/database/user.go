package database

import (
	"github.com/kiasaty/phrase-mate/models"
	"gorm.io/gorm"
)

func (c *Client) CreateUser(user *models.User) (*models.User, error) {
	if err := c.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (c *Client) FindUserByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User

	err := c.DB.Where("telegram_chat_id = ?", telegramID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (c *Client) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	if err := c.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
