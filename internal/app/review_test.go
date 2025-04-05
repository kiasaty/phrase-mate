package app

import (
	"testing"
	"time"

	"github.com/kiasaty/phrase-mate/internal/database"
	"github.com/kiasaty/phrase-mate/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *database.Client {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.Phrase{}, &models.Review{}, &models.ReviewHistory{}, &models.Session{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return &database.Client{DB: db}
}

func TestReviewPhrase(t *testing.T) {
	db := setupTestDB(t)
	app := NewApp(db, nil, GetDefaultConfig())

	// Create test user and phrase
	user := &models.User{
		TelegramChatID: 123,
		FirstName:      "Test",
		LastName:       "User",
	}
	if err := db.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	phrase := &models.Phrase{
		UserID:            user.ID,
		TelegramMessageID: 456,
		Text:              "Test phrase",
	}
	if err := db.DB.Create(phrase).Error; err != nil {
		t.Fatalf("Failed to create phrase: %v", err)
	}

	// Test first review with perfect recall
	review, err := app.ReviewPhrase(phrase.ID, user.ID, 1, models.QualityPerfect)
	assert.NoError(t, err)
	assert.NotNil(t, review)
	assert.Equal(t, float64(2.6), review.EaseFactor) // Initial ease factor for perfect review
	assert.Equal(t, uint16(1), review.Interval)      // First review interval
	assert.Equal(t, models.QualityPerfect, review.RecallQuality)

	// Move the next review date to the past to allow immediate review
	err = db.DB.Model(&models.Review{}).Where("phrase_id = ? AND user_id = ?", phrase.ID, user.ID).
		Update("next_review_at", time.Now().AddDate(0, 0, -1)).Error
	assert.NoError(t, err)

	// Test second review with hesitant recall
	review, err = app.ReviewPhrase(phrase.ID, user.ID, 1, models.QualityHesitant)
	assert.NoError(t, err)
	assert.NotNil(t, review)
	assert.True(t, review.EaseFactor < 2.6)     // Ease factor should decrease
	assert.Equal(t, uint16(1), review.Interval) // Interval should reset for low quality
	assert.Equal(t, models.QualityHesitant, review.RecallQuality)

	// Verify review history
	history, err := app.GetReviewHistory(user.ID, phrase.ID)
	assert.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, models.QualityHesitant, history[0].RecallQuality) // Latest review first
	assert.Equal(t, models.QualityPerfect, history[1].RecallQuality)  // First review second

	// Verify current review matches latest history
	currentReview, err := app.DB.FindReview(user.ID, phrase.ID)
	assert.NoError(t, err)
	assert.NotNil(t, currentReview)
	assert.Equal(t, history[0].RecallQuality, currentReview.RecallQuality)
	assert.Equal(t, history[0].EaseFactor, currentReview.EaseFactor)
	assert.Equal(t, history[0].Interval, currentReview.Interval)
}

func TestReviewPhraseDueDate(t *testing.T) {
	db := setupTestDB(t)
	app := NewApp(db, nil, GetDefaultConfig())

	// Create test user and phrase
	user := &models.User{
		TelegramChatID: 123,
		FirstName:      "Test",
		LastName:       "User",
	}
	if err := db.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	phrase := &models.Phrase{
		UserID:            user.ID,
		TelegramMessageID: 456,
		Text:              "Test phrase",
	}
	if err := db.DB.Create(phrase).Error; err != nil {
		t.Fatalf("Failed to create phrase: %v", err)
	}

	// First review
	review, err := app.ReviewPhrase(phrase.ID, user.ID, 1, models.QualityPerfect)
	assert.NoError(t, err)
	assert.NotNil(t, review)

	// Try to review again before due date
	review, err = app.ReviewPhrase(phrase.ID, user.ID, 1, models.QualityPerfect)
	assert.NoError(t, err)
	assert.Nil(t, review) // Should return nil as it's not due yet

	// Move the next review date to the past
	err = db.DB.Model(&models.Review{}).Where("phrase_id = ? AND user_id = ?", phrase.ID, user.ID).
		Update("next_review_at", time.Now().AddDate(0, 0, -1)).Error
	assert.NoError(t, err)

	// Now should be able to review
	review, err = app.ReviewPhrase(phrase.ID, user.ID, 1, models.QualityPerfect)
	assert.NoError(t, err)
	assert.NotNil(t, review)
}

func TestReviewPhraseInvalidQuality(t *testing.T) {
	db := setupTestDB(t)
	app := NewApp(db, nil, GetDefaultConfig())

	// Create test user and phrase
	user := &models.User{
		TelegramChatID: 123,
		FirstName:      "Test",
		LastName:       "User",
	}
	if err := db.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	phrase := &models.Phrase{
		UserID:            user.ID,
		TelegramMessageID: 456,
		Text:              "Test phrase",
	}
	if err := db.DB.Create(phrase).Error; err != nil {
		t.Fatalf("Failed to create phrase: %v", err)
	}

	// Test with invalid quality
	review, err := app.ReviewPhrase(phrase.ID, user.ID, 1, models.RecallQuality(6))
	assert.Error(t, err)
	assert.Nil(t, review)
}
