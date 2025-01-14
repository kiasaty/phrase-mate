package database

import "github.com/kiasaty/phrase-mate/models"

func (c *Client) CreatePhrase(phrase *models.Phrase) (*models.Phrase, error) {
	if err := c.DB.Create(phrase).Error; err != nil {
		return nil, err
	}
	return phrase, nil
}

func (c *Client) FindPhraseByMessageId(messageID int) (phrase *models.Phrase) {
	var p models.Phrase
	if err := c.DB.Preload("User").Preload("Tags").Where("message_id = ?", messageID).First(&p).Error; err != nil {
		return nil
	}
	return &p
}

func (c *Client) UpdatePhrase(phrase *models.Phrase) error {
	return c.DB.Save(phrase).Error
}

func (c *Client) UpdatePhraseTags(phrase *models.Phrase, tags *[]models.Tag) error {
	if err := c.DB.Model(phrase).Association("Tags").Replace(tags); err != nil {
		return err
	}
	return nil
}
