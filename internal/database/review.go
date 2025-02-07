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

func (c *Client) CountNotReviewedPhrasesBySessionId(sessionID uint) (int64, error) {
	var count int64

	err := c.DB.Model(&models.Review{}).
		Where("session_id = ? AND next_review_at <= ? AND reviewed_at IS NULL", sessionID, time.Now()).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (c *Client) GetDueReviews(userID uint, now time.Time, limit int) ([]models.Review, error) {
	var reviews []models.Review

	err := c.DB.
		Where("user_id = ? AND next_review_at <= ?", userID, now).
		Order("next_review_at ASC, ease_factor ASC").
		Limit(limit).
		Find(&reviews).Error

	if err != nil {
		return nil, err
	}

	return reviews, nil
}

func (c *Client) GetNewReviews(userID uint, sessionID uint, limit int) ([]models.Review, error) {
	var reviews []models.Review
	query := `
		SELECT 
			p.id AS phrase_id, 
			? AS user_id, 
			? AS session_id, 
			2.5 AS ease_factor, 
			0 AS interval, 
			NULL AS next_review_at, 
			NOW() AS reviewed_at
		FROM phrases p
		LEFT JOIN phrase_reviews pr ON p.id = pr.phrase_id AND pr.user_id = ?
		WHERE pr.id IS NULL
		LIMIT ?
	`
	err := c.DB.Raw(query, userID, sessionID, userID, limit).Scan(&reviews).Error
	if err != nil {
		return nil, err
	}

	return reviews, nil
}
