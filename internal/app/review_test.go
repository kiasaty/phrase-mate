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

func TestPhraseMastery(t *testing.T) {
	// Test 1: Perfect reviews should lead to mastery
	t.Run("Perfect reviews lead to mastery", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(db)

		config := GetDefaultConfig()
		config.MaxIntervalDays = 180 // 6 months
		app := NewApp(db, nil, config)

		// Create test user and phrase
		user := &models.User{
			TelegramChatID: 123,
			IsBot:          false,
		}
		user, err := db.CreateUser(user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		phrase := &models.Phrase{
			UserID:            user.ID,
			TelegramMessageID: 456,
			Text:              "test phrase",
		}
		phrase, err = db.CreatePhrase(phrase)
		if err != nil {
			t.Fatalf("Failed to create phrase: %v", err)
		}

		// Create a session
		session := &models.Session{
			UserID: user.ID,
		}
		session, err = db.CreateSession(session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Simulate perfect reviews
		var lastReview *models.Review
		for i := 0; i < 7; i++ {
			review, err := app.ReviewPhrase(phrase.ID, user.ID, session.ID, models.QualityPerfect)
			if err != nil {
				t.Fatalf("Failed to review phrase: %v", err)
			}
			if review == nil {
				t.Fatalf("Expected review to be created")
			}
			lastReview = review

			// Move the next review date to the past to allow immediate review
			if i < 6 {
				pastTime := time.Now().AddDate(0, 0, -1)
				lastReview.NextReviewAt = &pastTime
				if err := db.UpdateReview(lastReview); err != nil {
					t.Fatalf("Failed to update review: %v", err)
				}
			}
		}

		// Verify the phrase is marked as mastered
		updatedPhrase, err := db.FindPhrase(phrase.ID)
		if err != nil {
			t.Fatalf("Failed to find phrase: %v", err)
		}
		if !updatedPhrase.IsMastered {
			t.Error("Expected phrase to be marked as mastered")
		}
	})

	// Test 2: Imperfect reviews should not lead to mastery
	t.Run("Imperfect reviews do not lead to mastery", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(db)

		config := GetDefaultConfig()
		config.MaxIntervalDays = 180 // 6 months
		app := NewApp(db, nil, config)

		// Create test user and phrase
		user := &models.User{
			TelegramChatID: 123,
			IsBot:          false,
		}
		user, err := db.CreateUser(user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		phrase := &models.Phrase{
			UserID:            user.ID,
			TelegramMessageID: 789,
			Text:              "test phrase 2",
		}
		phrase, err = db.CreatePhrase(phrase)
		if err != nil {
			t.Fatalf("Failed to create phrase: %v", err)
		}

		// Create a session
		session := &models.Session{
			UserID: user.ID,
		}
		session, err = db.CreateSession(session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Simulate imperfect reviews
		var lastReview *models.Review
		for i := 0; i < 10; i++ { // More reviews than needed for mastery
			review, err := app.ReviewPhrase(phrase.ID, user.ID, session.ID, models.QualityHesitant)
			if err != nil {
				t.Fatalf("Failed to review phrase: %v", err)
			}
			if review == nil {
				t.Fatalf("Expected review to be created")
			}
			lastReview = review

			// Move the next review date to the past to allow immediate review
			pastTime := time.Now().AddDate(0, 0, -1)
			lastReview.NextReviewAt = &pastTime
			if err := db.UpdateReview(lastReview); err != nil {
				t.Fatalf("Failed to update review: %v", err)
			}
		}

		// Verify the phrase is not marked as mastered
		updatedPhrase, err := db.FindPhrase(phrase.ID)
		if err != nil {
			t.Fatalf("Failed to find phrase: %v", err)
		}
		if updatedPhrase.IsMastered {
			t.Error("Expected phrase to not be marked as mastered")
		}
	})

	// Test 3: Mixed reviews with eventual perfect ones should lead to mastery
	t.Run("Mixed reviews with eventual perfect ones lead to mastery", func(t *testing.T) {
		db := setupTestDB(t)
		defer cleanupTestDB(db)

		config := GetDefaultConfig()
		config.MaxIntervalDays = 180 // 6 months
		app := NewApp(db, nil, config)

		// Create test user and phrase
		user := &models.User{
			TelegramChatID: 123,
			IsBot:          false,
		}
		user, err := db.CreateUser(user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		phrase := &models.Phrase{
			UserID:            user.ID,
			TelegramMessageID: 101112,
			Text:              "test phrase 3",
		}
		phrase, err = db.CreatePhrase(phrase)
		if err != nil {
			t.Fatalf("Failed to create phrase: %v", err)
		}

		// Create a session
		session := &models.Session{
			UserID: user.ID,
		}
		session, err = db.CreateSession(session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Simulate mixed reviews
		qualities := []models.RecallQuality{
			models.QualityHesitant,
			models.QualityRemembered,
			models.QualityPerfect,
			models.QualityPerfect,
			models.QualityPerfect,
			models.QualityPerfect,
			models.QualityPerfect,
			models.QualityPerfect,
		}

		var lastReview *models.Review
		for i, quality := range qualities {
			review, err := app.ReviewPhrase(phrase.ID, user.ID, session.ID, quality)
			if err != nil {
				t.Fatalf("Failed to review phrase: %v", err)
			}
			if review == nil {
				t.Fatalf("Expected review to be created")
			}
			lastReview = review

			// Move the next review date to the past to allow immediate review
			if i < len(qualities)-1 { // Don't move the date after the last review
				pastTime := time.Now().AddDate(0, 0, -1)
				lastReview.NextReviewAt = &pastTime
				if err := db.UpdateReview(lastReview); err != nil {
					t.Fatalf("Failed to update review: %v", err)
				}
			}
		}

		// Verify the phrase is marked as mastered
		updatedPhrase, err := db.FindPhrase(phrase.ID)
		if err != nil {
			t.Fatalf("Failed to find phrase: %v", err)
		}
		if !updatedPhrase.IsMastered {
			t.Error("Expected phrase to be marked as mastered")
		}
	})
}

func cleanupTestDB(db *database.Client) {
	// Drop all tables
	db.DB.Migrator().DropTable(
		&models.User{},
		&models.Tag{},
		&models.Phrase{},
		&models.Review{},
		&models.ReviewHistory{},
		&models.Session{},
	)
}
