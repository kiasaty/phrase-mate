package app

import (
	"fmt"
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

type testSetup struct {
	app    *App
	user   *models.User
	phrase *models.Phrase
}

func setupTestReview(t *testing.T) *testSetup {
	db := setupTestDB(t)
	app := NewApp(db, nil, GetDefaultConfig())

	// Create test user
	user := &models.User{
		TelegramChatID: 123,
		FirstName:      "Test",
		LastName:       "User",
	}
	if err := db.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create test phrase
	phrase := &models.Phrase{
		UserID:            user.ID,
		TelegramMessageID: 456,
		Text:              "Test phrase",
	}
	if err := db.DB.Create(phrase).Error; err != nil {
		t.Fatalf("Failed to create phrase: %v", err)
	}

	return &testSetup{
		app:    app,
		user:   user,
		phrase: phrase,
	}
}

func TestReviewPhrase(t *testing.T) {
	// Test cases for different recall qualities
	testCases := []struct {
		name             string
		recallQuality    models.RecallQuality
		expectedEase     float64
		expectedInterval uint16
	}{
		{
			name:             "Perfect recall",
			recallQuality:    models.QualityPerfect,
			expectedEase:     2.6,
			expectedInterval: 1,
		},
		{
			name:             "Fluent recall",
			recallQuality:    models.QualityFluent,
			expectedEase:     2.5,
			expectedInterval: 1,
		},
		{
			name:             "Remembered recall",
			recallQuality:    models.QualityRemembered,
			expectedEase:     2.36,
			expectedInterval: 1,
		},
		{
			name:             "Hesitant recall",
			recallQuality:    models.QualityHesitant,
			expectedEase:     2.18,
			expectedInterval: 1,
		},
		{
			name:             "Forgot recall",
			recallQuality:    models.QualityForgot,
			expectedEase:     1.96,
			expectedInterval: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh setup for each test case
			setup := setupTestReview(t)

			// Create review
			review, err := setup.app.ReviewPhrase(setup.phrase.ID, setup.user.ID, 1, tc.recallQuality)
			assert.NoError(t, err)
			assert.NotNil(t, review)
			fmt.Println("review", review.EaseFactor, review.Interval, review.RecallQuality)
			assert.Equal(t, tc.expectedEase, review.EaseFactor)
			assert.Equal(t, tc.expectedInterval, review.Interval)
			assert.Equal(t, tc.recallQuality, review.RecallQuality)
		})
	}
}

func TestReviewPhraseDueDate(t *testing.T) {
	setup := setupTestReview(t)

	// First review
	review, err := setup.app.ReviewPhrase(setup.phrase.ID, setup.user.ID, 1, models.QualityPerfect)
	assert.NoError(t, err)
	assert.NotNil(t, review)

	// Try to review again before due date
	review, err = setup.app.ReviewPhrase(setup.phrase.ID, setup.user.ID, 1, models.QualityPerfect)
	assert.NoError(t, err)
	assert.Nil(t, review) // Should return nil as it's not due yet

	// Move the next review date to the past
	err = setup.app.DB.(*database.Client).DB.Model(&models.Review{}).
		Where("phrase_id = ? AND user_id = ?", setup.phrase.ID, setup.user.ID).
		Update("next_review_at", time.Now().AddDate(0, 0, -1)).Error
	assert.NoError(t, err)

	// Now should be able to review
	review, err = setup.app.ReviewPhrase(setup.phrase.ID, setup.user.ID, 1, models.QualityPerfect)
	assert.NoError(t, err)
	assert.NotNil(t, review)
}

func TestReviewPhraseInvalidQuality(t *testing.T) {
	setup := setupTestReview(t)

	// Test with invalid quality
	review, err := setup.app.ReviewPhrase(setup.phrase.ID, setup.user.ID, 1, models.RecallQuality(6))
	assert.Error(t, err)
	assert.Nil(t, review)
}
