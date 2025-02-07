package database

import (
	"github.com/kiasaty/phrase-mate/models"
	"gorm.io/gorm"
)

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

func (c *Client) FindNextPhraseToReviewBySessionID(sessionID uint) (*models.Phrase, error) {
	var review models.Review

	err := c.DB.Where("session_id = ? AND reviewed_at IS NULL", sessionID).
		Order("next_review_at ASC, ease_factor ASC").
		First(&review).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return c.FindPhrase(review.PhraseID)
}

func (c *Client) FindPhrase(phraseID uint) (*models.Phrase, error) {
	var phrase models.Phrase

	result := c.DB.First(&phrase, phraseID)

	if result.Error != nil {
		return nil, result.Error
	}

	return &phrase, nil
}
