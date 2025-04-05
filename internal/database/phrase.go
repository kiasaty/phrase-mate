package database

import (
	"github.com/kiasaty/phrase-mate/models"
)

func (c *Client) CreatePhrase(phrase *models.Phrase) (*models.Phrase, error) {
	if err := c.DB.Create(phrase).Error; err != nil {
		return nil, err
	}
	return phrase, nil
}

func (c *Client) FindPhraseByMessageId(messageID int) (phrase *models.Phrase) {
	var p models.Phrase
	if err := c.DB.Preload("User").Preload("Tags").Where("telegram_message_id = ?", messageID).First(&p).Error; err != nil {
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

func (c *Client) FindPhrase(phraseID uint) (*models.Phrase, error) {
	var phrase models.Phrase

	result := c.DB.First(&phrase, phraseID)

	if result.Error != nil {
		return nil, result.Error
	}

	return &phrase, nil
}

func (c *Client) FindNewPhrasesToReview(userID uint, limit int) ([]uint, error) {
	var phraseIDs []uint
	query := `
		SELECT 
			phrases.id AS phrase_id 
		FROM phrases
		LEFT JOIN reviews ON phrases.id = reviews.phrase_id AND reviews.user_id = ?
		WHERE reviews.id IS NULL
		AND phrases.is_mastered = false
		LIMIT ?
	`
	err := c.DB.Raw(query, userID, limit).Scan(&phraseIDs).Error
	if err != nil {
		return nil, err
	}

	return phraseIDs, nil
}

func (c *Client) MarkPhraseAsMastered(phraseID uint) error {
	return c.DB.Model(&models.Phrase{}).
		Where("id = ?", phraseID).
		Update("is_mastered", true).
		Error
}
