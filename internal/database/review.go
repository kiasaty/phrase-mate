package database

import (
	"time"

	"github.com/kiasaty/phrase-mate/models"
	"gorm.io/gorm"
)

func (c *Client) CreateReview(review *models.Review) error {
	// First, find existing review
	existingReview, err := c.FindReview(review.UserID, review.PhraseID)
	if err != nil {
		return err
	}

	if existingReview != nil {
		// Update the existing review
		review.ID = existingReview.ID
		if err := c.UpdateReview(review); err != nil {
			return err
		}
	} else {
		// Create new review if it doesn't exist
		if err := c.DB.Create(review).Error; err != nil {
			return err
		}
	}

	// Store the review in history
	history := &models.ReviewHistory{
		PhraseID:      review.PhraseID,
		UserID:        review.UserID,
		SessionID:     review.SessionID,
		RecallQuality: review.RecallQuality,
		EaseFactor:    review.EaseFactor,
		Interval:      review.Interval,
		ReviewedAt:    review.ReviewedAt,
		NextReviewAt:  review.NextReviewAt,
	}

	return c.CreateReviewHistory(history)
}

func (c *Client) UpdateReview(review *models.Review) error {
	return c.DB.Save(review).Error
}

func (c *Client) FindReview(userID, phraseID uint) (*models.Review, error) {
	var review models.Review

	err := c.DB.Where("phrase_id = ? AND user_id = ?", phraseID, userID).
		First(&review).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &review, nil
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
	var review models.Review

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

	return &review, nil
}

// Review history operations
func (c *Client) CreateReviewHistory(review *models.ReviewHistory) error {
	return c.DB.Create(review).Error
}

func (c *Client) FindReviewHistory(userID uint, phraseID uint) ([]*models.ReviewHistory, error) {
	var history []*models.ReviewHistory
	err := c.DB.Where("user_id = ? AND phrase_id = ?", userID, phraseID).
		Order("reviewed_at DESC").
		Find(&history).Error
	if err != nil {
		return nil, err
	}
	return history, nil
}
