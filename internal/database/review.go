package database

import (
	"time"

	"github.com/kiasaty/phrase-mate/models"
	"gorm.io/gorm"
)

func (c *Client) CreateReview(review *models.Review) error {
	return c.DB.Create(review).Error
}

func (c *Client) CreateReviews(reviews []models.Review) error {
	return c.DB.Create(&reviews).Error
}

func (c *Client) FindPhraseLastReview(userID, phraseID uint) (*models.Review, error) {
	var lastReview models.Review

	err := c.DB.Where("phrase_id = ? AND user_id = ?", phraseID, userID).
		Order("reviewed_at DESC").
		First(&lastReview).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &lastReview, nil
}

func (c *Client) CountReviewedPhrasesInSession(sessionID uint) (uint, error) {
	var count int64

	err := c.DB.Model(&models.Review{}).
		Where("session_id = ?", sessionID, time.Now()).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return uint(count), nil
}

func (c *Client) GetDueReview(userID uint, now time.Time, limit uint) (*models.Review, error) {
	var review *models.Review

	err := c.DB.
		Where("user_id = ? AND next_review_at <= ?", userID, now).
		Order("next_review_at ASC, ease_factor ASC").
		Limit(int(limit)).
		First(&review).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return review, nil
}
