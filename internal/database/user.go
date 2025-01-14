package database

import "github.com/kiasaty/phrase-mate/models"

func (c *Client) CreateUser(user *models.User) (*models.User, error) {
	if err := c.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (c *Client) FindUserByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User
	if err := c.DB.Where("telegram_chat_id = ?", telegramID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
